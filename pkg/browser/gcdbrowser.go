package browser

import (
	"context"
	"os"

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

type GCDBrowser struct {
	g *gcd.Gcd
}

func NewGCDBrowser() *GCDBrowser {
	b := &GCDBrowser{}
	b.g = gcd.NewChromeDebugger()
	return b
}

func (b *GCDBrowser) Init() error {
	return b.Restart()
}

func (b *GCDBrowser) Restart() error {
	b.g.AddFlags(startupFlags)
	if err := b.g.StartProcess("/usr/bin/chromium-browser", "/tmp/blarg", "9022"); err != nil {
		return errors.Wrap(err, "failed to start browser")
	}
	return nil
}

func (b *GCDBrowser) Close() error {
	os.RemoveAll("/tmp/blarg")
	return b.g.ExitProcess()
}

func (b *GCDBrowser) Load(ctx context.Context, address *am.ScanGroupAddress, scheme, port string) (string, string, []*am.HTTPResponse, error) {
	t, err := b.g.NewTab()
	if err != nil {
		return "", "", nil, errors.Wrap(err, "error creating tab")
	}
	defer b.g.CloseTab(t)

	tab := NewTab(t, address)
	tab.CaptureNetworkTraffic(ctx, address, port)

	url := scheme + "://"

	if address.HostAddress != "" {
		url += address.HostAddress
		tab.InjectIP(port, scheme)
	} else {
		// no host address, just use IP
		url += address.IPAddress
	}

	if (scheme == "http" && port != "80") || (scheme == "https" && port != "443") {
		url += ":" + port
	}

	err = tab.LoadPage(ctx, url)
	if err != nil {
		log.Warn().Err(err).Msg("loading page")
	}

	img, err := tab.TakeScreenshot(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("unable to take screenshot")
	}

	log.Info().Msg("closed tab")
	return img, tab.SerializeDOM(), tab.GetNetworkTraffic(), nil
}
