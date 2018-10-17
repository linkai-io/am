package web

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gammazero/workerpool"
	"github.com/linkai-io/am/pkg/parsers"

	"github.com/linkai-io/am/pkg/retrier"

	"github.com/linkai-io/am/pkg/browser"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module"
	"github.com/linkai-io/am/services/module/brute/state"
	"github.com/miekg/dns"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	oneHour = 60 * 60
)

var (
	ErrEmptyWebData     = errors.New("webData was empty from load")
	ErrEmptyHostAddress = errors.New("hostaddress was empty")
)

var schemes = []string{"http", "https"}

// Web will brute force and mutate subdomains to attempt to find
// additional hosts
type Web struct {
	st       state.Stater
	dc       *dnsclient.Client
	browsers browser.Browser
	// for closing subscriptions to listen for group updates
	exitContext context.Context
	cancel      context.CancelFunc
	// concurrent safe cache of scan groups updated via Subscribe callbacks
	groupCache *cache.ScanGroupSubscriber
}

// New brute force module
func New(browsers browser.Browser, dc *dnsclient.Client, st state.Stater) *Web {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Web{browsers: browsers, st: st, exitContext: ctx, cancel: cancel}
	b.dc = dc
	// start cache subscriber and listen for updates
	b.groupCache = cache.NewScanGroupSubscriber(ctx, st)
	return b
}

// Init the web module
func (w *Web) Init() error {
	return nil
}

func (w *Web) defaultPortConfig() *am.PortModuleConfig {
	return &am.PortModuleConfig{
		RequestsPerSecond: 50,
		CustomPorts:       []int32{80, 443, 8443, 8000, 9200},
	}
}

func (w *Web) defaultWebConfig() *am.WebModuleConfig {
	return &am.WebModuleConfig{
		TakeScreenShots:       true,
		RequestsPerSecond:     50,
		MaxLinks:              1,
		ExtractJS:             true,
		FingerprintFrameworks: true,
	}
}

// shouldAnalyze determines if we should analyze the specific address or not.
func (w *Web) shouldAnalyze(ctx context.Context, logger zerolog.Logger, address *am.ScanGroupAddress) bool {
	if address.IsWildcardZone {
		return false
	}

	switch uint16(address.NSRecord) {
	case dns.TypeMX, dns.TypeNS, dns.TypeSRV:
		return false
	}

	if address.UserConfidenceScore > 75 {
		return true
	}

	if address.ConfidenceScore < 75 {
		logger.Info().Float32("confidence", address.ConfidenceScore).Msg("score too low")
		return false
	}

	return true
}

// Analyze will attempt to find additional domains by extracting hosts from a website as well
// as capture any network traffic, save images, dom, and responses to s3/disk
func (w *Web) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {

	portCfg := w.defaultPortConfig()
	// := w.defaultWebConfig()
	logger := log.With().
		Int("OrgID", userContext.GetOrgID()).
		Int("UserID", userContext.GetUserID()).
		Str("TraceID", userContext.GetTraceID()).
		Str("IPAddress", address.IPAddress).
		Str("HostAddress", address.HostAddress).
		Int64("AddressID", address.AddressID).
		Str("AddressHash", address.AddressHash).Logger()

	webRecords := make(map[string]*am.ScanGroupAddress, 0)
	if !w.shouldAnalyze(ctx, logger, address) {
		logger.Info().Msg("not analyzing")
		return address, webRecords, nil
	}

	if group, err := w.groupCache.GetGroupByIDs(address.OrgID, address.GroupID); err != nil {
		logger.Warn().Err(err).Msg("unable to find group id in cache, using default settings")
	} else {
		portCfg = group.ModuleConfigurations.PortModule
	}

	for _, port := range portCfg.CustomPorts {
		// do stuff
		logger.Info().Int32("port", port).Msg("analyzing")
		portStr := strconv.Itoa(int(port))
		for _, scheme := range schemes {

			webData := &am.WebData{}
			retryErr := retrier.RetryAttempts(func() error {
				var err error
				webData, err = w.browsers.Load(ctx, address, scheme, portStr)
				return err
			}, 3)
			if retryErr != nil {
				continue
			}

			hosts, err := w.processWebData(ctx, logger, address, webData)
			if err != nil {
				continue
			}
			for k, v := range hosts {
				webRecords[k] = v
			}
		}
	}

	return address, webRecords, nil

}

func (w *Web) processWebData(ctx context.Context, logger zerolog.Logger, address *am.ScanGroupAddress, webData *am.WebData) (map[string]*am.ScanGroupAddress, error) {
	newAddresses := make(map[string]*am.ScanGroupAddress, 0)

	if webData == nil {
		return nil, ErrEmptyWebData
	}

	etld, err := parsers.GetETLD(address.HostAddress)
	if err != nil {
		return nil, ErrEmptyHostAddress
	}

	if webData.Responses != nil {
		extractedHosts := w.extractHostsFromResponses(etld, webData)
		resolvedAddresses := w.resolveNewDomains(ctx, logger, address, extractedHosts, am.DiscoveryWebCrawler)
		for k, v := range resolvedAddresses {
			newAddresses[k] = v
		}
	}
	return newAddresses, nil
}

func (w *Web) resolveNewDomains(ctx context.Context, logger zerolog.Logger, address *am.ScanGroupAddress, newAddresses map[string]struct{}, discoveryMethod string) map[string]*am.ScanGroupAddress {
	webRecords := make(map[string]*am.ScanGroupAddress, 0)

	numAddresses := len(newAddresses)
	rps := 10
	if numAddresses < 10 {
		rps = len(newAddresses)
	}
	pool := workerpool.New(rps)

	type results struct {
		R        []*dnsclient.Results
		Hostname string
		Err      error
	}

	out := make(chan *results)
	wg := &sync.WaitGroup{}
	wg.Add(numAddresses)

	for newHost := range newAddresses {
		task := func(ctx context.Context, host string, wg *sync.WaitGroup, out chan<- *results) func() {
			return func() {
				r, err := w.dc.ResolveName(ctx, host)
				out <- &results{Hostname: host, R: r, Err: err}
				wg.Done()
			}
		}
		pool.Submit(task(ctx, newHost, wg, out))
	}

	go func() {
		wg.Wait()
		close(out)
		pool.Stop()
		logger.Info().Msg("all tasks completed")
	}()

	for result := range out {
		if result.Err != nil {
			continue
		}

		for _, rr := range result.R {
			for _, ip := range rr.IPs {
				logger.Info().Str("hostname", result.Hostname).Str("ip_address", ip).Msg("found new record")
				newAddress := module.NewAddressFromDNS(address, ip, result.Hostname, discoveryMethod, uint(rr.RecordType))
				newAddress.ConfidenceScore = module.CalculateConfidence(logger, address, newAddress)
				webRecords[newAddress.AddressHash] = newAddress
			}
		}
	}

	return webRecords
}

func (w *Web) extractHostsFromResponses(etld string, webData *am.WebData) map[string]struct{} {
	allHosts := make(map[string]struct{}, 0)

	zone := strings.Replace(etld, ".", "\\.", -1)

	needle, err := regexp.Compile(zone)
	if err != nil {
		return allHosts
	}

	needles := make([]*regexp.Regexp, 1)
	needles[0] = needle

	for _, resp := range webData.Responses {
		found := parsers.ExtractHostsFromResponse(needles, resp.RawBody)
		for k := range found {
			allHosts[k] = struct{}{}
		}
		if resp.WebCertificate != nil {
			//resp.WebCertificate.SubjectName
		}

	}

	return allHosts
}
