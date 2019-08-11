package browser

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/webtech"

	"github.com/wirepair/gcd/gcdmessage"

	"github.com/linkai-io/am/am"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/pkg/errors"
	"github.com/wirepair/gcd"
)

var startupFlags = []string{
	//"--allow-insecure-localhost",
	"--enable-automation",
	"--enable-features=NetworkService",
	"--test-type",
	//"--ignore-certificate-errors",
	//"--ignore-ssl-errors",
	//"--ignore-certificate-errors-spki-list",
	"--disable-client-side-phishing-detection",
	"--disable-component-update",
	"--disable-infobars",
	"--disable-ntp-popular-sites",
	"--disable-ntp-most-likely-favicons-from-server",
	"--disable-sync-app-list",
	"--disable-domain-reliability",
	"--disable-background-networking",
	"--disable-sync",
	"--disable-new-browser-first-run",
	"--disable-default-apps",
	"--disable-popup-blocking",
	"--disable-extensions",
	"--disable-features=TranslateUI",
	"--disable-gpu",
	"--disable-dev-shm-usage",
	"--no-sandbox",
	//"--metrics-recording-only",
	"--allow-running-insecure-content",
	"--no-first-run",
	"--window-size=1024,768",
	"--safebrowsing-disable-auto-update",
	"--safebrowsing-disable-download-protection",
	//"--deterministic-fetch",

	"--password-store=basic",
	//"--proxy-server=localhost:8080",
	// TODO: re-investigate headless periodically, currently intercepting TLS requests and replacing
	// hostnames with ip addresses fails.
	"--headless",
	"about:blank",
}

var (
	ErrBrowserClosing = errors.New("unable to load, as closing down")
)

type GCDBrowserPool struct {
	profileDir       string
	maxBrowsers      int
	acquiredBrowsers int32
	acquireErrors    int32
	browsers         chan *gcd.Gcd
	browserTimeout   time.Duration
	closing          int32
	display          string
	detector         webtech.Detector
	leaser           LeaserService
	startCount       int32
	logger           zerolog.Logger
}

func NewGCDBrowserPool(maxBrowsers int, leaser LeaserService, techDetect webtech.Detector) *GCDBrowserPool {
	b := &GCDBrowserPool{}
	b.detector = techDetect
	b.maxBrowsers = maxBrowsers
	b.browserTimeout = time.Second * 45
	b.leaser = leaser
	b.browsers = make(chan *gcd.Gcd, b.maxBrowsers)
	return b
}

// UseDisplay (to be called before Init()) tells chrome to start using an Xvfb display
func (b *GCDBrowserPool) UseDisplay(display string) {
	b.display = fmt.Sprintf("DISPLAY=%s", display)
}

// Init starts the browser/Browser pool
func (b *GCDBrowserPool) Init() error {
	if _, err := b.leaser.Cleanup(); err != nil {
		return err
	}
	return b.Start()
}

// SetAPITimeout tells gcd how long to wait for a response from the browser for all API calls
func (b *GCDBrowserPool) SetAPITimeout(duration time.Duration) {
	b.browserTimeout = duration
}

// Start the browser with a random profile directory and create Browsers
func (b *GCDBrowserPool) Start() error {
	// allow 3 seconds per Browser
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(b.maxBrowsers*3))
	defer cancel()
	// clean up just in case we are restarting
	if _, err := b.leaser.Cleanup(); err != nil {
		panic("failed to clean up browsers")
	}

	log.Info().Int("browsers", b.maxBrowsers).Msg("creating browsers")
	b.browsers = make(chan *gcd.Gcd, b.maxBrowsers)

	atomic.AddInt32(&b.startCount, 1)
	currentCount := atomic.LoadInt32(&b.startCount)
	// always have 2 browsers ready
	for i := 0; i < b.maxBrowsers; i++ {
		b.Return(timeoutCtx, nil, currentCount) // passing nil will just create a new one for us
		log.Info().Int("i", i).Msg("browser created")
	}

	time.Sleep(time.Second * 2) // give time for browser to settle
	return nil
}

// Acquire a Browser, unless context expired. If expired, increment our Browser error count
// which is used to restart the entire browser process after a max limit on errors
// is reached
func (b *GCDBrowserPool) Acquire(ctx context.Context) *gcd.Gcd {

	select {
	case browser := <-b.browsers:
		if browser != nil {
			atomic.AddInt32(&b.acquiredBrowsers, 1)
		}
		return browser
	case <-ctx.Done():
		log.Warn().Err(ctx.Err()).Msg("failed to acquire Browser from pool")
		atomic.AddInt32(&b.acquireErrors, 1)
		b.shouldRestart()
		return nil
	}
}

// Closing a channel that may be being read will cause a panic, which is fine because
// then we just restart anyways
func (b *GCDBrowserPool) shouldRestart() {
	acquired := atomic.LoadInt32(&b.acquiredBrowsers)
	errored := atomic.LoadInt32(&b.acquireErrors)
	count, _ := b.leaser.Count()
	log.Warn().Int32("acquired", acquired).Int32("errored", errored).Str("leaser_count", count).Msg("force restarting due to failure to acquire browsers")
	// flag as shutting down and clear out errors
	atomic.StoreInt32(&b.closing, 1)
	atomic.StoreInt32(&b.acquiredBrowsers, 0)
	atomic.StoreInt32(&b.acquireErrors, 0)
	// empty pool
	for {
		select {
		case <-b.browsers:
			log.Info().Msg("emptying browser")
		default:
			goto EMPTY
		}
	}
EMPTY:
	time.Sleep(1 * time.Second)
	log.Info().Msg("calling restart")
	if err := b.Start(); err != nil {
		panic("restarting due to failure to restart browsers process")
	}
	atomic.StoreInt32(&b.closing, 0)
}

// Return a browser
func (b *GCDBrowserPool) Return(ctx context.Context, browser *gcd.Gcd, startCount int32) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	doneCh := make(chan struct{})

	go b.closeAndCreateBrowser(browser, doneCh, startCount)

	select {
	case <-timeoutCtx.Done():
		log.Error().Msg("failed to closeAndCreateBrowser in time")
	case <-doneCh:
		return
	}
}

// closeAndCreateBrowser takes an optional Browser to close, and creates a new one, closing doneCh
// to signal it completed (although it may be a nil browser if error occurred).
func (b *GCDBrowserPool) closeAndCreateBrowser(browser *gcd.Gcd, doneCh chan struct{}, startCount int32) {
	if browser != nil {
		if err := b.leaser.Return(browser.Port()); err != nil {
			log.Error().Err(err).Msg("failed to return browser")
		}
		atomic.AddInt32(&b.acquiredBrowsers, -1)
	}

	// if we've restarted and this browser was still leased, we don't want to create another one
	currentCount := atomic.LoadInt32(&b.startCount)
	if currentCount != startCount {
		close(doneCh)
		return
	}

	browser = gcd.NewChromeDebugger()
	port, err := b.leaser.Acquire()
	if err != nil {
		log.Warn().Err(err).Msg("unable to acquire new browser")
		b.browsers <- nil
		close(doneCh)
		return
	}

	if err := browser.ConnectToInstance("localhost", string(port)); err != nil {
		log.Warn().Err(err).Msg("failed to connect to instance")
		browser = nil
	}

	b.browsers <- browser
	close(doneCh)
}

func (b *GCDBrowserPool) LoadForDiff(ctx context.Context, address *am.ScanGroupAddress, scheme, port string) (string, string, error) {
	var browser *gcd.Gcd
	startCount := atomic.LoadInt32(&b.startCount) // track if we've restarted so we can throw away bad browsers

	if atomic.LoadInt32(&b.closing) == 1 {
		return "", "", ErrBrowserClosing
	}

	// if nil, do not return browser
	if browser = b.Acquire(ctx); browser == nil {
		return "", "", errors.New("browser acquisition failed during LoadForDiff")
	}
	defer b.Return(ctx, browser, startCount)

	logger := log.With().Str("HostAddress", address.HostAddress).Str("port", port).Str("call", "LoadForDiff").Logger()
	ctx = logger.WithContext(ctx)

	log.Ctx(ctx).Info().Int32("acquired", atomic.LoadInt32(&b.acquiredBrowsers)).Int32("errors", atomic.LoadInt32(&b.acquireErrors)).Msg("acquired browser")

	t, err := browser.GetFirstTab()
	if err != nil {
		log.Error().Err(err).Msg("failed to get first tab")
		return "", "", err
	}

	defer browser.CloseTab(t) // closes websocket go routines

	t.SetApiTimeout(b.browserTimeout)

	tab := NewTab(ctx, t, address)
	defer tab.Close()
	url := b.buildURL(tab, address, scheme, port)

	//log.Ctx(ctx).Info().Msg("loading url")

	if err := tab.LoadPage(ctx, url); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("loading page")
		if err == ErrNavigationTimedOut {
			return "", "", err
		}
		if chromeErr, ok := err.(*gcdmessage.ChromeApiTimeoutErr); ok {
			return "", "", errors.Wrap(chromeErr, "failed to load page due to timeout")
		}

		if errors.Cause(err) == ErrNavigating || errors.Cause(err) == ErrTabCrashed {
			return "", "", err
		}
	}

	// not necessary, but to make sure 'timing' is as close to possible as the actual Load to capture, keep for now
	//log.Ctx(ctx).Info().Str("url", url).Msg("taking screenshot")
	_, err = tab.TakeScreenshot(ctx)
	//log.Ctx(ctx).Info().Str("url", url).Msg("screenshot taken")
	if err != nil {
		log.Warn().Err(err).Msg("unable to take screenshot")
	}
	dom := tab.SerializeDOM()
	loadURL := tab.GetURL(ctx)
	log.Ctx(ctx).Info().Msg("closed browser")
	return loadURL, dom, nil
}

// Load an address of scheme and port, returning an image, the dom, all text based responses or an error.
func (b *GCDBrowserPool) Load(ctx context.Context, address *am.ScanGroupAddress, scheme, port string) (*am.WebData, error) {
	var browser *gcd.Gcd
	startCount := atomic.LoadInt32(&b.startCount) // track if we've restarted so we can throw away bad browsers

	requestPort, _ := strconv.Atoi(port) // safe to assume this won't fail
	if atomic.LoadInt32(&b.closing) == 1 {
		return nil, ErrBrowserClosing
	}

	// if nil, do not return browser
	if browser = b.Acquire(ctx); browser == nil {
		return nil, errors.New("browser acquisition failed during Load")
	}
	defer b.Return(ctx, browser, startCount)

	logger := log.With().Str("HostAddress", address.HostAddress).Str("port", port).Str("call", "Load").Logger()
	ctx = logger.WithContext(ctx)
	log.Ctx(ctx).Info().Int32("acquired", atomic.LoadInt32(&b.acquiredBrowsers)).Int32("errors", atomic.LoadInt32(&b.acquireErrors)).Msg("acquired browser")

	t, err := browser.GetFirstTab()
	if err != nil {
		log.Error().Err(err).Msg("failed to get first tab")
		return nil, err
	}

	defer browser.CloseTab(t) // closes websocket go routines

	t.SetApiTimeout(b.browserTimeout)

	tab := NewTab(ctx, t, address)
	defer tab.Close()

	tab.CaptureNetworkTraffic(ctx, address, port)

	url := b.buildURL(tab, address, scheme, port)

	log.Ctx(ctx).Info().Msg("loading url")

	start := time.Now().UnixNano()
	if err := tab.LoadPage(ctx, url); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("loading page")
		if err == ErrNavigationTimedOut {
			return nil, err
		}
		if chromeErr, ok := err.(*gcdmessage.ChromeApiTimeoutErr); ok {
			return nil, errors.Wrap(chromeErr, "failed to load page due to timeout")
		}

		if errors.Cause(err) == ErrNavigating || errors.Cause(err) == ErrTabCrashed {
			return nil, err
		}
	}

	log.Ctx(ctx).Info().Str("url", url).Msg("taking screenshot")
	img, err := tab.TakeScreenshot(ctx)
	log.Ctx(ctx).Info().Str("url", url).Msg("screenshot taken")
	if err != nil {
		log.Warn().Err(err).Msg("unable to take screenshot")
	}
	dom := tab.SerializeDOM()

	loadResponse, allResponses := tab.GetNetworkTraffic()
	// if for some reason we can't find the load response, just fake it
	// with current address details :(
	if loadResponse == nil {
		log.Ctx(ctx).Warn().Msg("unable to find original load response, using address data")
		loadResponse = &am.HTTPResponse{
			Headers:      make(map[string]string, 0),
			HostAddress:  address.HostAddress,
			IPAddress:    address.IPAddress,
			ResponsePort: port,
			URL:          url,
			Scheme:       scheme,
		}
	}

	responsePort, err := strconv.Atoi(loadResponse.ResponsePort)
	if err != nil {
		responsePort = requestPort
	}

	domMatches := b.detector.DOM(dom)
	headerMatches := b.detector.Headers(loadResponse.Headers)
	jsResults, err := tab.InjectJS(b.detector.JSToInject())
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to inject web tech detection js")
	}
	jsObjects := b.detector.JSResultsToObjects(jsResults)
	jsMatches := b.detector.JS(jsObjects)

	webTech := b.detector.MergeMatches([]map[string][]*webtech.Match{domMatches, headerMatches, jsMatches})

	// TODO: follow redirect loadResponse and use that.
	finalURL := tab.GetURL(ctx)
	if finalURL == "" {
		finalURL = loadResponse.URL
	}

	webData := &am.WebData{
		Address:             address,
		Responses:           allResponses,
		Snapshot:            img,
		URL:                 finalURL,
		AddressHash:         convert.HashAddress(loadResponse.IPAddress, loadResponse.HostAddress),
		HostAddress:         loadResponse.HostAddress,
		IPAddress:           loadResponse.IPAddress,
		ResponsePort:        responsePort,
		RequestedPort:       requestPort,
		Scheme:              loadResponse.Scheme,
		SerializedDOM:       dom,
		SerializedDOMHash:   convert.HashData([]byte(dom)),
		ResponseTimestamp:   loadResponse.ResponseTimestamp,
		DetectedTech:        webTech,
		URLRequestTimestamp: start,
		LoadURL:             url,
	}

	log.Ctx(ctx).Info().Msg("closed browser")
	return webData, nil
}

// Close all browsers and return. TODO: make this not terrible.
func (b *GCDBrowserPool) Close(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&b.closing, 0, 1) {
		return nil
	}

	for {
		browser := b.Acquire(ctx)
		if browser != nil {
			if err := b.leaser.Return(browser.Port()); err != nil {
				log.Error().Err(err).Msg("failed to return browser")
			}
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if len(b.browsers) == 0 {
			return nil
		}
	}
}

// buildURL and signal the browser to inject IP address if we have an IP/Host pair
// TODO: renable injecting IP once fixed/resolved...
func (b *GCDBrowserPool) buildURL(tab *Tab, address *am.ScanGroupAddress, scheme, port string) string {
	url := scheme + "://"
	if address.HostAddress != "" {
		url += address.HostAddress
		//tab.InjectIP(scheme, port)
	} else {
		// no host address, just use IP
		url += address.IPAddress
	}

	if (scheme == "http" && port != "80") || (scheme == "https" && port != "443") {
		url += ":" + port
	}
	return url
}
