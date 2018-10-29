package web

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/linkai-io/am/pkg/filestorage"

	"github.com/gammazero/workerpool"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/browser"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/pkg/parsers"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/services/module"
	"github.com/linkai-io/am/services/module/web/state"
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
	storage  filestorage.Storage
	// for closing subscriptions to listen for group updates
	exitContext context.Context
	cancel      context.CancelFunc
	// concurrent safe cache of scan groups updated via Subscribe callbacks
	groupCache *cache.ScanGroupSubscriber
}

// New web analysis module
func New(browsers browser.Browser, dc *dnsclient.Client, st state.Stater, storage filestorage.Storage) *Web {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Web{st: st, exitContext: ctx, cancel: cancel}

	b.browsers = browsers
	b.dc = dc
	b.storage = storage
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
		CustomPorts:       []int32{80, 443},
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

	host := address.HostAddress
	if host == "" {
		host = address.IPAddress
	}

	shouldWeb, err := w.st.DoWebDomain(ctx, address.OrgID, address.GroupID, oneHour, host)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to check do web domain")
		return false
	}

	if !shouldWeb {
		logger.Info().Msg("not analyzing web for domain, as it is already complete")
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

			// don't bother trying https for port 80 and http for port 443
			if (port == 80 && scheme != "http") || (port == 443 && scheme != "https") {
				continue
			}

			webData := &am.WebData{}
			retryErr := retrier.RetryAttempts(func() error {
				var err error
				logger.Info().Int32("port", port).Str("scheme", scheme).Msg("calling load")
				webData, err = w.browsers.Load(ctx, address, scheme, portStr)
				return err
			}, 2)

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
	var hash string
	var link string
	var err error

	newAddresses := make(map[string]*am.ScanGroupAddress, 0)

	if webData == nil {
		return nil, ErrEmptyWebData
	}

	_, link, err = w.storage.Write(ctx, address, []byte(webData.Snapshot))
	if err != nil {
		logger.Warn().Err(err).Msg("failed to write snapshot data to storage")
	}
	webData.SnapshotLink = link

	hash, link, err = w.storage.Write(ctx, address, []byte(webData.SerializedDOM))
	if err != nil {
		logger.Warn().Err(err).Msg("failed to write serialized dom data to storage")

	}
	webData.SerializedDOMHash = hash
	webData.SerializedDOMLink = link

	if webData.Responses != nil {
		extractedHosts := w.processResponses(ctx, logger, address, webData)
		resolvedAddresses := w.resolveNewDomains(ctx, logger, address, extractedHosts, am.DiscoveryWebCrawler)
		for k, v := range resolvedAddresses {
			newAddresses[k] = v
		}
	}
	return newAddresses, nil
}

// processResponses iterates over all responses, extracting additional domains and creating a hash of the
// body data and save it to file storage
func (w *Web) processResponses(ctx context.Context, logger zerolog.Logger, address *am.ScanGroupAddress, webData *am.WebData) map[string]struct{} {
	var extractHosts bool
	var zone string
	var needles []*regexp.Regexp

	allHosts := make(map[string]struct{}, 0)

	etld, err := parsers.GetETLD(address.HostAddress)
	if err == nil {
		extractHosts = true
	}

	if extractHosts {
		zone = strings.Replace(etld, ".", "\\.", -1)

		needle, err := regexp.Compile("(?i)" + zone)
		if err != nil {
			return allHosts
		}

		needles = make([]*regexp.Regexp, 1)
		needles[0] = needle
	}

	// iterate over responses and save to filestorage. If we have a proper etld,
	// extract hosts from the body and certificates.
	for _, resp := range webData.Responses {
		if resp == nil {
			continue
		}

		hash, link, err := w.storage.Write(ctx, address, []byte(resp.RawBody))
		if err != nil {
			logger.Warn().Err(err).Msg("unable to process hash/link for raw data")
		}
		resp.RawBodyHash = hash
		resp.RawBodyLink = link

		if !extractHosts {
			continue
		}

		found := parsers.ExtractHostsFromResponse(needles, resp.RawBody)
		for k := range found {
			allHosts[k] = struct{}{}
		}

		if resp.WebCertificate == nil {
			continue
		}

		allHosts[resp.WebCertificate.SubjectName] = struct{}{}

		if resp.WebCertificate.SanList == nil || len(resp.WebCertificate.SanList) == 0 {
			continue
		}

		// TODO: add 'verified' hosts to this suffix check
		for _, host := range resp.WebCertificate.SanList {
			if strings.HasSuffix(host, etld) {
				allHosts[host] = struct{}{}
			}
		}
	}
	return allHosts
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
		h := newHost
		pool.Submit(task(ctx, h, wg, out))
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
