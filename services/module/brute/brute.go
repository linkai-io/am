package brute

import (
	"bufio"
	"context"
	"io"
	"strings"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/pkg/parsers"
	"github.com/linkai-io/am/services/module"
	"github.com/linkai-io/am/services/module/brute/state"
	"github.com/miekg/dns"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	oneHour = 60 * 60
)

// Bruter will brute force and mutate subdomains to attempt to find
// additional hosts
type Bruter struct {
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
func New(dc *dnsclient.Client, st state.Stater) *Bruter {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Bruter{st: st, exitContext: ctx, cancel: cancel}
	b.dc = dc
	// start cache subscriber and listen for updates
	b.groupCache = cache.NewScanGroupSubscriber(ctx, st)
	return b
}

// Init the brute forcer with the initial input subdomain list
func (b *Bruter) Init(bruteFile io.Reader) error {
	fileScanner := bufio.NewScanner(bruteFile)
	b.subdomains = make([]string, 0)

	for fileScanner.Scan() {
		b.subdomains = append(b.subdomains, strings.TrimSpace(fileScanner.Text()))
	}
	return nil
}

func (b *Bruter) defaultModuleConfig() *am.BruteModuleConfig {
	return &am.BruteModuleConfig{
		MaxDepth:          2,
		RequestsPerSecond: 50,
		CustomSubNames:    make([]string, 0),
	}
}

// Analyze will attempt to find additional domains by brute forcing hosts. Note that while we will not brute force past
// max depth, we *will* attempt to mutate hosts past max depth.
func (b *Bruter) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {
	bmc := b.defaultModuleConfig()
	logger := log.With().
		Int("OrgID", userContext.GetOrgID()).
		Int("UserID", userContext.GetUserID()).
		Str("TraceID", userContext.GetTraceID()).
		Str("IPAddress", address.IPAddress).
		Str("HostAddress", address.HostAddress).
		Int64("AddressID", address.AddressID).
		Str("AddressHash", address.AddressHash).Logger()

	bruteRecords := make(map[string]*am.ScanGroupAddress, 0)
	if !b.shouldAnalyze(ctx, logger, address) {
		logger.Info().Msg("not analyzing")
		return address, bruteRecords, nil
	}

	if group, err := b.groupCache.GetGroupByIDs(address.OrgID, address.GroupID); err != nil {
		logger.Warn().Err(err).Msg("unable to find group id in cache, using default settings")
	} else {
		bmc = group.ModuleConfigurations.BruteModule
	}
	bruteRecords = b.bruteDomain(ctx, logger, bmc, address)

	mutateRecords := b.mutateDomain(ctx, logger, bmc, address)
	for k, v := range mutateRecords {
		bruteRecords[k] = v
	}
	return address, bruteRecords, nil
}

// shouldAnalyze determines if we should analyze the specific address or not. Updates address.IsWildcardZone
// if tested.
func (b *Bruter) shouldAnalyze(ctx context.Context, logger zerolog.Logger, address *am.ScanGroupAddress) bool {
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

func (b *Bruter) bruteDomain(ctx context.Context, logger zerolog.Logger, bmc *am.BruteModuleConfig, address *am.ScanGroupAddress) map[string]*am.ScanGroupAddress {
	bruteRecords := make(map[string]*am.ScanGroupAddress, 0)
	depth, err := parsers.GetDepth(address.HostAddress)
	if err != nil || int32(depth) > bmc.MaxDepth {
		logger.Info().Err(err).Int("depth", depth).Int32("max_depth", bmc.MaxDepth).Msg("not brute forcing due to depth")
		return bruteRecords
	}

	shouldBrute, err := b.st.DoBruteDomain(ctx, address.OrgID, address.GroupID, oneHour, address.HostAddress)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to check do brute force domain")
		return bruteRecords
	}

	if !shouldBrute {
		logger.Info().Msg("not brute forcing domain, as it is already complete")
		return bruteRecords
	}

	if isWildcard := b.dc.IsWildcard(ctx, address.HostAddress); isWildcard {
		address.IsWildcardZone = true
		logger.Info().Msg("not brute forcing due to wildcard zone")
		return bruteRecords
	}

	subdomains := BuildSubDomainList(b.subdomains, bmc.CustomSubNames)
	log.Info().Msg("start brute forcing")
	bruteRecords = b.bruteDomains(ctx, logger, address, address.HostAddress, subdomains, am.DiscoveryBruteSubDomain, int(bmc.RequestsPerSecond))
	log.Info().Msg("brute forcing complete")
	return bruteRecords
}

func (b *Bruter) mutateDomain(ctx context.Context, logger zerolog.Logger, bmc *am.BruteModuleConfig, address *am.ScanGroupAddress) map[string]*am.ScanGroupAddress {
	mutateRecords := make(map[string]*am.ScanGroupAddress, 0)
	depth, err := parsers.GetDepth(address.HostAddress)
	if err != nil || int32(depth) > bmc.MaxDepth {
		logger.Info().Int("depth", depth).Int32("max_depth", bmc.MaxDepth).Msg("not brute forcing due to depth")
		return mutateRecords
	}

	subDomain, domain, err := parsers.GetSubDomainAndDomain(address.HostAddress)
	if err != nil {
		logger.Warn().Err(err).Msg("not mutating")
		return mutateRecords
	}

	if subDomain == "" {
		logger.Info().Msg("no subdomain, not mutating")
		return mutateRecords
	}

	subdomains := NumberMutation(subDomain)
	if len(subdomains) == 0 {
		return mutateRecords
	}

	unmutatedSubDomain := UnmutateNumber(subDomain)

	shouldMutate, err := b.st.DoMutateDomain(ctx, address.OrgID, address.GroupID, oneHour, unmutatedSubDomain)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to check do mutate domain")
		return mutateRecords
	}

	if !shouldMutate {
		logger.Info().Msg("not brute forcing domain, as it is already complete")
		return mutateRecords
	}

	// although we are checking is wildcard 2x, this is the 'rare' case since we've usually already
	// checked this host for mutations.
	if isWildcard := b.dc.IsWildcard(ctx, address.HostAddress); isWildcard {
		address.IsWildcardZone = true
		return mutateRecords
	}

	log.Info().Msg("start mutating")
	mutateRecords = b.bruteDomains(ctx, logger, address, domain, subdomains, am.DiscoveryBruteMutator, int(bmc.RequestsPerSecond))
	log.Info().Msg("mutating complete")
	return mutateRecords
}

func (b *Bruter) bruteDomains(ctx context.Context, logger zerolog.Logger, address *am.ScanGroupAddress, hostAddress string, subdomains []string, discoveryMethod string, requestsPerSecond int) map[string]*am.ScanGroupAddress {

	newHosts := make(map[string]struct{}, len(subdomains))
	for _, subdomain := range subdomains {
		newHosts[subdomain+"."+hostAddress] = struct{}{}
	}

	return module.ResolveNewAddresses(ctx, logger, b.dc, &module.ResolverData{
		Address:           address,
		RequestsPerSecond: requestsPerSecond,
		NewAddresses:      newHosts,
		DiscoveryMethod:   discoveryMethod,
	})
}

// BuildSubDomainList merges the base list with any custom domains in the scan group config
func BuildSubDomainList(list, custom []string) []string {
	totalDomains := len(list) + len(custom)
	subdomains := make([]string, totalDomains)
	i := 0
	for ; i < len(list); i++ {
		subdomains[i] = strings.Trim(list[i], "\n\t ")
	}

	for j := 0; j < len(custom); j++ {
		subdomains[i] = strings.Trim(custom[j], "\n\t ")
		i++
	}

	return subdomains
}
