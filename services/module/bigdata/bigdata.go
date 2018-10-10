package brute

import (
	"context"
	"io"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/brute/state"
	"github.com/miekg/dns"
	"github.com/rs/zerolog"
)

const (
	oneHour = 60 * 60
)

// BigData will query our big data database looking for sub domains
// of a etld.
type BigData struct {
	st         state.Stater
	dc         *dnsclient.Client
	subdomains []string

	// for closing subscriptions to listen for group updates
	exitContext context.Context
	cancel      context.CancelFunc
	// concurrent safe cache of scan groups updated via Subscribe callbacks
	groupCache *cache.ScanGroupSubscriber
}

// New brute force module
func New(dc *dnsclient.Client, st state.Stater) *BigData {
	ctx, cancel := context.WithCancel(context.Background())
	b := &BigData{st: st, exitContext: ctx, cancel: cancel}
	b.dc = dc
	// start cache subscriber and listen for updates
	b.groupCache = cache.NewScanGroupSubscriber(ctx, st)
	return b
}

// Init the brute forcer with the initial input subdomain list
func (b *BigData) Init(bruteFile io.Reader) error {

	return nil
}

// Analyze will attempt to find additional domains by brute forcing hosts. Note that while we will not brute force past
// max depth, we *will* attempt to mutate hosts past max depth.
func (b *BigData) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {
	/*
		bmc := b.defaultModuleConfig()
		logger := log.With().
			Int("OrgID", userContext.GetOrgID()).
			Int("UserID", userContext.GetUserID()).
			Str("TraceID", userContext.GetTraceID()).
			Str("IPAddress", address.IPAddress).
			Str("HostAddress", address.HostAddress).
			Int64("AddressID", address.AddressID).
			Str("AddressHash", address.AddressHash).Logger()

		BigDataecords := make(map[string]*am.ScanGroupAddress, 0)
		if !b.shouldAnalyze(ctx, logger, address) {
			logger.Info().Msg("not analyzing")
			return address, BigDataecords, nil
		}

		if group, err := b.groupCache.GetGroupByIDs(address.OrgID, address.GroupID); err != nil {
			logger.Warn().Err(err).Msg("unable to find group id in cache, using default settings")
		} else {
			//bmc = group.ModuleConfigurations.BruteModule
		}
	*/
	return nil, nil, nil
}

// shouldAnalyze determines if we should analyze the specific address or not. Updates address.IsWildcardZone
// if tested.
func (b *BigData) shouldAnalyze(ctx context.Context, logger zerolog.Logger, address *am.ScanGroupAddress) bool {
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
