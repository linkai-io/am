package web

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/brute/state"
	"github.com/miekg/dns"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	oneHour = 60 * 60
)

// Web will brute force and mutate subdomains to attempt to find
// additional hosts
type Web struct {
	st state.Stater
	dc *dnsclient.Client

	// for closing subscriptions to listen for group updates
	exitContext context.Context
	cancel      context.CancelFunc
	// concurrent safe cache of scan groups updated via Subscribe callbacks
	groupCache *cache.ScanGroupSubscriber
}

// New brute force module
func New(dc *dnsclient.Client, st state.Stater) *Web {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Web{st: st, exitContext: ctx, cancel: cancel}
	b.dc = dc
	// start cache subscriber and listen for updates
	b.groupCache = cache.NewScanGroupSubscriber(ctx, st)
	return b
}

// Init the brute forcer with the initial input subdomain list
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

// Analyze will attempt to find additional domains by extracting hosts from a website as well
// as capture any network traffic
func (w *Web) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {

	portCfg := w.defaultPortConfig()
	webCfg := w.defaultWebConfig()
	logger := log.With().
		Int("OrgID", userContext.GetOrgID()).
		Int("UserID", userContext.GetUserID()).
		Str("TraceID", userContext.GetTraceID()).
		Str("IPAddress", address.IPAddress).
		Str("HostAddress", address.HostAddress).
		Int64("AddressID", address.AddressID).
		Str("AddressHash", address.AddressHash).Logger()

	WebRecords := make(map[string]*am.ScanGroupAddress, 0)
	if !w.shouldAnalyze(ctx, logger, address) {
		logger.Info().Msg("not analyzing")
		return address, WebRecords, nil
	}

	if group, err := w.groupCache.GetGroupByIDs(address.OrgID, address.GroupID); err != nil {
		logger.Warn().Err(err).Msg("unable to find group id in cache, using default settings")
	}

	return address, WebRecords, nil

	return nil, nil, nil
}

// shouldAnalyze determines if we should analyze the specific address or not. Updates address.IsWildcardZone
// if tested.
func (w *Web) shouldAnalyze(ctx context.Context, logger zerolog.Logger, address *am.ScanGroupAddress) bool {
	if address.HostAddress == "" || address.IsWildcardZone || address.IsHostedService {
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
