package web

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/pkg/filestorage"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/browser"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/pkg/parsers"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/services/module"
	"github.com/linkai-io/am/services/module/web/state"
	"github.com/miekg/dns"
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
var defaultPorts = []int32{80, 443}

// Web will brute force and mutate subdomains to attempt to find
// additional hosts
type Web struct {
	st            state.Stater
	dc            *dnsclient.Client
	webDataClient am.WebDataService
	browsers      browser.Browser
	storage       filestorage.Storage
	// for closing subscriptions to listen for group updates
	exitContext context.Context
	cancel      context.CancelFunc
	// concurrent safe cache of scan groups updated via Subscribe callbacks
	groupCache *cache.ScanGroupSubscriber
}

// New web analysis module
func New(browsers browser.Browser, webDataClient am.WebDataService, dc *dnsclient.Client, st state.Stater, storage filestorage.Storage) *Web {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Web{st: st, exitContext: ctx, cancel: cancel}

	b.browsers = browsers
	b.webDataClient = webDataClient
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

// shouldAnalyze determines if we should analyze the specific address or not.
func (w *Web) shouldAnalyze(ctx context.Context, address *am.ScanGroupAddress) bool {
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
		log.Ctx(ctx).Warn().Err(err).Msg("unable to check do web domain")
		return false
	}

	if !shouldWeb {
		log.Ctx(ctx).Info().Msg("not analyzing web for domain, as it is already complete")
		return false
	}

	if address.UserConfidenceScore > 75 {
		return true
	}

	if address.ConfidenceScore < 75 {
		log.Ctx(ctx).Info().Float32("confidence", address.ConfidenceScore).Msg("score too low")
		return false
	}

	return true
}

// Analyze will attempt to find additional domains by extracting hosts from a website as well
// as capture any network traffic, save images, dom, and responses to s3/disk
func (w *Web) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {
	portCfg := module.DefaultPortConfig()
	nsCfg := module.DefaultNSConfig()
	ctx = module.DefaultLogger(ctx, userContext, address)

	webRecords := make(map[string]*am.ScanGroupAddress, 0)

	if !w.shouldAnalyze(ctx, address) {
		log.Ctx(ctx).Info().Msg("not analyzing")
		return address, webRecords, nil
	}

	if group, err := w.groupCache.GetGroupByIDs(address.OrgID, address.GroupID); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("unable to find group id in cache, using default settings")
	} else {
		portCfg = group.ModuleConfigurations.PortModule
		nsCfg = group.ModuleConfigurations.NSModule
	}

	allPorts := make(map[int32]struct{}, 0)
	for _, port := range defaultPorts {
		allPorts[port] = struct{}{}
	}
	for _, port := range portCfg.CustomPorts {
		allPorts[port] = struct{}{}
	}

	for port := range allPorts {
		// do stuff
		log.Ctx(ctx).Info().Int32("port", port).Msg("analyzing")
		portStr := strconv.Itoa(int(port))
		for _, scheme := range schemes {

			// don't bother trying https for port 80 and http for port 443
			if (port == 80 && scheme != "http") || (port == 443 && scheme != "https") {
				continue
			}

			webData := &am.WebData{}
			retryErr := retrier.RetryAttempts(func() error {
				var err error
				log.Ctx(ctx).Info().Int32("port", port).Str("scheme", scheme).Msg("calling load")
				webData, err = w.browsers.Load(ctx, address, scheme, portStr)
				return err
			}, 2)

			if retryErr != nil {
				continue
			}

			hosts, err := w.processWebData(ctx, userContext, nsCfg, address, webData)
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

func (w *Web) processWebData(ctx context.Context, userContext am.UserContext, nsCfg *am.NSModuleConfig, address *am.ScanGroupAddress, webData *am.WebData) (map[string]*am.ScanGroupAddress, error) {
	var hash string
	var link string
	var err error

	newAddresses := make(map[string]*am.ScanGroupAddress, 0)

	if webData == nil {
		return nil, ErrEmptyWebData
	}

	snapshotData, err := base64.StdEncoding.DecodeString(webData.Snapshot)
	if err == nil {
		_, link, err = w.storage.Write(ctx, userContext, address, snapshotData)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("failed to write snapshot data to storage")
		}
	} else {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to decode snapshot data")
	}
	webData.SnapshotLink = link

	hash, link, err = w.storage.Write(ctx, userContext, address, []byte(webData.SerializedDOM))
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to write serialized dom data to storage")
	}
	webData.SerializedDOMHash = hash
	webData.SerializedDOMLink = link

	if webData.Responses != nil {
		extractedHosts := w.processResponses(ctx, userContext, address, webData)
		resolvedAddresses := module.ResolveNewAddresses(ctx, w.dc, &module.ResolverData{
			Address:           address,
			RequestsPerSecond: int(nsCfg.RequestsPerSecond),
			NewAddresses:      extractedHosts,
			DiscoveryMethod:   am.DiscoveryWebCrawler,
		})

		for k, v := range resolvedAddresses {
			newAddresses[k] = v
		}
	}

	if _, err := w.webDataClient.Add(ctx, userContext, webData); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to upload webdata to service")
	}

	return newAddresses, nil
}

// processResponses iterates over all responses, extracting additional domains and creating a hash of the
// body data and save it to file storage
func (w *Web) processResponses(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, webData *am.WebData) map[string]struct{} {
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
	certDuplicates := make(map[string]struct{}, 0)

	for _, resp := range webData.Responses {
		if resp == nil {
			continue
		}

		hash, link, err := w.storage.Write(ctx, userContext, address, []byte(resp.RawBody))
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Msg("unable to process hash/link for raw data")
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

		// make sure we don't have duplicate certificates otherwise webdataservice will throw an error
		// during insert
		unique := certificateHash(resp.WebCertificate)
		if _, ok := certDuplicates[unique]; ok {
			resp.WebCertificate = nil
		} else {
			certDuplicates[unique] = struct{}{}
		}

	}
	return allHosts
}

// certificateHash is for ensuring only unique certificates get added to the WebData structure. Only unique
// values may be UPSERT'd from a temporary table.
func certificateHash(cert *am.WebCertificate) string {
	data := fmt.Sprintf("%s.%s.%s.%d.%d", cert.SubjectName, cert.Cipher, cert.Mac, cert.ValidFrom, cert.ValidTo)
	return convert.HashData([]byte(data))
}
