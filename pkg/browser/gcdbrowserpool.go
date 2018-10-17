package browser

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"sync/atomic"
	"time"

	"github.com/wirepair/gcd/gcdmessage"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/rs/zerolog/log"

	"github.com/pkg/errors"
	"github.com/wirepair/gcd"
)

var startupFlags = []string{"--ignore-certificate-errors",
	"--test-type",
	"--disable-background-networking",
	"--disable-sync",
	"--no-sandbox",
	"--disable-default-apps",
	"--disable-popup-blocking",
	"--enable-automation",
	"--metrics-recording-only",
	"--allow-running-insecure-content",
	"--disable-new-Browser-first-run",
	"--no-first-run",
	"--window-size=1024,768",
	"--disable-features=TranslateUI",
	"--safebrowsing-disable-auto-update",
	"--disable-component-update",
	"--safebrowsing-disable-download-protection",
	"--disable-gpu",
	//"--deterministic-fetch",
	"--password-store=basic",
	// TODO: re-investigate headless periodically, currently intercepting TLS requests and replacing
	// hostnames with ip addresses fails.
	//"--headless",
}

var (
	ErrBrowserClosing = errors.New("unable to load, as closing down")
)

type GCDBrowserPool struct {
	profileDir       string
	maxBrowsers      int
	acquiredBrowsers int32
	browsers         chan *gcd.Gcd
	browserTimeout   time.Duration
	closing          int32
	display          string
}

func NewGCDBrowserPool(maxBrowsers int) *GCDBrowserPool {
	b := &GCDBrowserPool{}

	b.maxBrowsers = maxBrowsers
	b.browserTimeout = time.Second * 30
	b.browsers = make(chan *gcd.Gcd, maxBrowsers)
	return b
}

// UseDisplay (to be called before Init()) tells chrome to start using an Xvfb display
func (b *GCDBrowserPool) UseDisplay(display string) {
	b.display = fmt.Sprintf("DISPLAY=%s", display)
}

// Init starts the browser/Browser pool
func (b *GCDBrowserPool) Init() error {
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

	log.Info().Int("browsers", b.maxBrowsers).Msg("creating browsers")
	// always have 2 browsers ready
	for i := 0; i < b.maxBrowsers; i++ {
		b.Return(timeoutCtx, nil) // passing nil will just create a new one for us
		log.Info().Int("i", i).Msg("browser created")
	}

	time.Sleep(time.Second * 5) // give time for browser to settle
	return nil
}

// Acquire a Browser, unless context expired. If expired, increment our Browser error count
// which is used to restart the entire browser process aftere a max limit on errors
// is reached
func (b *GCDBrowserPool) Acquire(ctx context.Context) *gcd.Gcd {

	select {
	case browser := <-b.browsers:
		atomic.AddInt32(&b.acquiredBrowsers, 1)
		return browser
	case <-ctx.Done():
		log.Warn().Err(ctx.Err()).Msg("failed to acquire Browser from pool")
		return nil
	}
}

// Return a browser Browser, if we fail to return the browser increment the Browser error
// count.
func (b *GCDBrowserPool) Return(ctx context.Context, browser *gcd.Gcd) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	doneCh := make(chan struct{})

	go b.closeAndCreateBrowser(browser, doneCh)

	select {
	case <-timeoutCtx.Done():
	case <-doneCh:
		return
	}
}

// Close all browsers and return. TODO: make this not terrible.
func (b *GCDBrowserPool) Close(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&b.closing, 0, 1) {
		return nil
	}

	for {
		browser := b.Acquire(ctx)
		if browser != nil {
			browser.ExitProcess()
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if len(b.browsers) == 0 {
			return nil
		}
	}
}

// closeAndCreateBrowser takes an optional Browser to close, and creates a new one, closing doneCh
// to signal it was successfully created.
func (b *GCDBrowserPool) closeAndCreateBrowser(browser *gcd.Gcd, doneCh chan struct{}) {
	if browser != nil {
		browser.ExitProcess()
		atomic.AddInt32(&b.acquiredBrowsers, -1)
	}

	browser = gcd.NewChromeDebugger()
	browser.DeleteProfileOnExit()

	browser.AddFlags(startupFlags)
	if b.display != "" {
		browser.AddEnvironmentVars([]string{b.display})
	}

	if err := browser.StartProcess("/usr/bin/chromium-browser", b.randProfile(), b.randPort()); err != nil {
		log.Error().Err(err).Msg("failed to start browser")
		return
	}

	b.browsers <- browser
	close(doneCh)
}

// Load an address of scheme and port, returning an image, the dom, all text based responses or an error.
// Care is taken if we have an unsBrowserle browser and will signal a restart of the entire browser process
// if we reach max Browser errors
func (b *GCDBrowserPool) Load(ctx context.Context, address *am.ScanGroupAddress, scheme, port string) (*am.WebData, error) {
	var browser *gcd.Gcd

	if atomic.LoadInt32(&b.closing) == 1 {
		return nil, ErrBrowserClosing
	}

	if browser = b.Acquire(ctx); browser == nil {
		return nil, errors.New("browser acquisition failed during Load")
	}

	log.Info().Msg("acquired browser")
	defer b.Return(ctx, browser)
	t, err := browser.GetFirstTab()
	t.SetApiTimeout(b.browserTimeout)

	tab := NewTab(t, address)
	log.Info().Msg("capturing traffic")
	tab.CaptureNetworkTraffic(ctx, address, port)

	url := b.buildURL(tab, address, scheme, port)

	log.Info().Msg("loading url")
	if err := tab.LoadPage(ctx, url); err != nil {
		log.Warn().Err(err).Msg("loading page")
		if err == ErrNavigationTimedOut {
			return nil, err
		}
		if chromeErr, ok := err.(*gcdmessage.ChromeApiTimeoutErr); ok {
			return nil, errors.Wrap(chromeErr, "failed to load page due to timeout")
		}
	}

	log.Info().Str("url", url).Msg("taking screenshot")
	img, err := tab.TakeScreenshot(ctx)
	log.Info().Str("url", url).Msg("screenshot taken")
	if err != nil {
		log.Warn().Err(err).Msg("unable to take screenshot")
	}

	webData := &am.WebData{
		Address:           address,
		SerializedDOM:     tab.SerializeDOM(),
		Responses:         tab.GetNetworkTraffic(),
		Snapshot:          img,
		ResponseTimestamp: time.Now().UnixNano(),
	}

	log.Info().Msg("closed Browser")
	return webData, nil
}

// buildURL and signal the Browser to inject IP address if we have an IP/Host pair
func (b *GCDBrowserPool) buildURL(tab *Tab, address *am.ScanGroupAddress, scheme, port string) string {
	url := scheme + "://"
	if address.HostAddress != "" {
		url += address.HostAddress
		tab.InjectIP(scheme, port)
	} else {
		// no host address, just use IP
		url += address.IPAddress
	}

	if (scheme == "http" && port != "80") || (scheme == "https" && port != "443") {
		url += ":" + port
	}
	return url
}

func (b *GCDBrowserPool) randPort() string {
	var l net.Listener
	retryErr := retrier.Retry(func() error {
		var err error
		l, err = net.Listen("tcp", ":0")
		return err
	})

	if retryErr != nil {
		log.Warn().Err(retryErr).Msg("unable to get port using default 9022")
		return "9022"
	}
	_, randPort, _ := net.SplitHostPort(l.Addr().String())
	l.Close()
	return randPort
}

func (b *GCDBrowserPool) randProfile() string {
	profile, err := ioutil.TempDir("/tmp", "gcd")
	if err != nil {
		log.Error().Err(err).Msg("failed to create temporary profile directory")
		return "/tmp/gcd"
	}
	return profile
}
