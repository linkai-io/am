package browser

import (
	"context"
	"io/ioutil"
	"os"
	"sync/atomic"
	"time"

	"github.com/wirepair/gcd/gcdmessage"

	"github.com/linkai-io/am/am"
	"github.com/rs/zerolog/log"

	"github.com/pkg/errors"
	"github.com/wirepair/gcd"
)

var startupFlags = []string{"--test-type",
	"--ignore-certificate-errors",
	"--allow-running-insecure-content",
	"--disable-new-tab-first-run",
	"--no-first-run",
	"--disable-translate",
	"--safebrowsing-disable-auto-update",
	"--disable-component-update",
	"--safebrowsing-disable-download-protection",
	"--deterministic-fetch",
	"--headless",
}

var (
	ErrBrowserRestarting = errors.New("browser currently restarting")
)

type GCDBrowser struct {
	g            *gcd.Gcd
	profileDir   string
	maxTabs      int
	maxTabErrors int
	acquiredTabs int32
	closing      int32
	tabErrors    int32
	tabs         chan *gcd.ChromeTarget
	tabTimeout   time.Duration
}

func NewGCDBrowser(maxTabs, maxTabErrors int) *GCDBrowser {
	b := &GCDBrowser{}
	b.g = gcd.NewChromeDebugger()
	b.maxTabs = maxTabs
	b.maxTabErrors = maxTabErrors
	b.tabTimeout = time.Second * 30
	b.tabs = make(chan *gcd.ChromeTarget, maxTabs)
	return b
}

func (b *GCDBrowser) Init() error {
	return b.Start()
}

func (b *GCDBrowser) SetAPITimeout(duration time.Duration) {
	b.tabTimeout = duration
}

func (b *GCDBrowser) Start() error {
	b.profileDir = b.randProfile()
	b.g.AddFlags(startupFlags)
	if err := b.g.StartProcess("/usr/bin/chromium-browser", b.profileDir, "9022"); err != nil {
		return errors.Wrap(err, "failed to start browser")
	}

	b.createTabs()
	return nil
}

func (b *GCDBrowser) randProfile() string {
	profile, err := ioutil.TempDir("/tmp", "gcdbrowser")
	if err != nil {
		log.Error().Err(err).Msg("failed to create temporary profile directory")
		return "/tmp/gcdbrowser"
	}
	return profile
}

func (b *GCDBrowser) Close() error {
	os.RemoveAll(b.profileDir)
	return b.g.ExitProcess()
}

// createTabs and set closing to false.
func (b *GCDBrowser) createTabs() {
	// allow 2 seconds per tab
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(b.maxTabs*2))
	defer cancel()

	log.Info().Int("tabs", b.maxTabs).Msg("creating tabs")
	for i := 0; i < b.maxTabs; i++ {
		b.Return(timeoutCtx, nil) // passing nil will just create a new one for us
		log.Info().Int("i", i).Msg("tab created")
	}
	atomic.StoreInt32(&b.closing, 0)
	atomic.StoreInt32(&b.tabErrors, 0)
	time.Sleep(time.Second * 2) // give time for browser to settle
}

// Acquire a tab, unless context expired. If expired, increment our tab error count
// which is used to restart the entire browser process aftere a max limit on errors
// is reached
func (b *GCDBrowser) Acquire(ctx context.Context) *gcd.ChromeTarget {
	if atomic.LoadInt32(&b.closing) == 1 {
		return nil
	}

	select {
	case t := <-b.tabs:
		atomic.AddInt32(&b.acquiredTabs, 1)
		return t
	case <-ctx.Done():
		log.Warn().Err(ctx.Err()).Msg("failed to acquire tab from pool")
		atomic.AddInt32(&b.tabErrors, 1)
		return nil
	}
}

// Return a browser tab, if we fail to return the browser increment the tab error
// count.
func (b *GCDBrowser) Return(ctx context.Context, t *gcd.ChromeTarget) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	doneCh := make(chan struct{})

	go b.closeAndCreateTab(t, doneCh)

	select {
	case <-timeoutCtx.Done():
		atomic.AddInt32(&b.tabErrors, 1)
	case <-doneCh:
		return
	}
}

// closeAndCreateTab takes an optional tab to close, and creates a new one, closing doneCh
// to signal it was successfully created.
func (b *GCDBrowser) closeAndCreateTab(t *gcd.ChromeTarget, doneCh chan struct{}) {
	if t != nil {
		b.g.CloseTab(t)
		atomic.AddInt32(&b.acquiredTabs, -1)
	}

	t, err := b.g.NewTab()
	if err != nil {
		log.Warn().Err(err).Msg("unable to open new tab during return")
		return
	}
	b.tabs <- t
	close(doneCh)
}

func (b *GCDBrowser) drain() {
	for t := range b.tabs {
		b.g.CloseTab(t)
		atomic.AddInt32(&b.acquiredTabs, -1)
	}
}

func (b *GCDBrowser) signalRestart(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&b.closing, 0, 1) {
		return errors.New("already closing down browser")
	}

	log.Warn().Msg("signaling browser restart")
	ticker := time.NewTicker(time.Millisecond * 150)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*90)
	defer cancel()

	for {
		select {
		case <-ticker.C:
			log.Info().Int32("acquired_tabs", atomic.LoadInt32(&b.acquiredTabs)).Msg("tabs open")
			if atomic.LoadInt32(&b.acquiredTabs) != 0 {
				continue
			}
			if err := b.Close(); err != nil {
				log.Error().Err(err).Msg("error closing browser process")
			}
			log.Info().Msg("browser restarting")
			return b.Start()
		case <-timeoutCtx.Done():
			if err := b.Close(); err != nil {
				log.Error().Err(err).Msg("error closing browser process")
			}
			log.Warn().Msg("timed out waiting for browsers, restarting anyways")
			return b.Start()
		}
	}
}

// Load an address of scheme and port, returning an image, the dom, all text based responses or an error.
// Care is taken if we have an unstable browser and will signal a restart of the entire browser process
// if we reach max tab errors
func (b *GCDBrowser) Load(ctx context.Context, address *am.ScanGroupAddress, scheme, port string) (*am.WebData, error) {
	var t *gcd.ChromeTarget

	if atomic.LoadInt32(&b.closing) == 1 {
		return nil, ErrBrowserRestarting
	}

	if atomic.LoadInt32(&b.tabErrors) > int32(b.maxTabErrors) {
		if err := b.signalRestart(ctx); err != nil {
			return nil, errors.Wrap(err, "failed to restart browser during load")
		}
		log.Info().Msg("browser acquired after restart completed.")
	}

	if t = b.Acquire(ctx); t == nil {
		return nil, errors.New("browser acquisition failed during Load")
	}
	log.Info().Msg("acquired browser")
	defer b.Return(ctx, t)

	t.SetApiTimeout(b.tabTimeout)

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

	img, err := tab.TakeScreenshot(ctx)
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

	log.Info().Msg("closed tab")
	return webData, nil
}

// buildURL and signal the tab to inject IP address if we have an IP/Host pair
func (b *GCDBrowser) buildURL(tab *Tab, address *am.ScanGroupAddress, scheme, port string) string {
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
