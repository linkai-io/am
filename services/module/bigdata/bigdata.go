package bigdata

import (
	"context"
	"time"

	"github.com/linkai-io/am/services/module"

	"github.com/linkai-io/am/pkg/bq"
	"github.com/linkai-io/am/pkg/certstream"
	"github.com/linkai-io/am/pkg/retrier"

	"github.com/linkai-io/am/pkg/parsers"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/bigdata/state"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

const (
	defaultCacheTime = time.Hour * 24
	oneHour          = 60 * 60
)

// BigData will query our locally cached results of big data first, then the bigquery database
// if we should need updated values looking for sub domains of a etld.
type BigData struct {
	st           state.Stater
	dc           *dnsclient.Client
	bdClient     am.BigDataService
	bigQuerier   bq.BigQuerier
	certListener certstream.Listener

	// for closing subscriptions to listen for group updates
	exitContext context.Context
	cancel      context.CancelFunc
	// concurrent safe cache of scan groups updated via Subscribe callbacks
	groupCache *cache.ScanGroupSubscriber
}

// New big data analysis module
func New(dc *dnsclient.Client, st state.Stater, bdClient am.BigDataService, bqQuerier bq.BigQuerier, certListener certstream.Listener) *BigData {
	ctx, cancel := context.WithCancel(context.Background())
	b := &BigData{exitContext: ctx, cancel: cancel}
	b.st = st
	b.dc = dc
	b.bdClient = bdClient
	b.bigQuerier = bqQuerier
	b.certListener = certListener
	// start cache subscriber and listen for updates
	b.groupCache = cache.NewScanGroupSubscriber(ctx, st)
	return b
}

// Init the bigdata service with bigquery details
func (b *BigData) Init(config []byte) error {

	return nil
}

// shouldAnalyze determines if we should analyze the specific address or not.
func (b *BigData) shouldAnalyze(ctx context.Context, address *am.ScanGroupAddress) bool {
	if address.HostAddress == "" || address.IsHostedService {
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
		log.Ctx(ctx).Info().Float32("confidence", address.ConfidenceScore).Msg("score too low")
		return false
	}

	return true
}

// Analyze will attempt to find additional domains by looking in various public data sets we've collected. We only
// do this for ETLDs.
func (b *BigData) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {
	nsCfg := module.DefaultNSConfig()
	ctx = module.DefaultLogger(ctx, userContext, address)

	bigDataRecords := make(map[string]*am.ScanGroupAddress, 0)

	if !b.shouldAnalyze(ctx, address) {
		log.Ctx(ctx).Info().Msg("not analyzing")
		return address, bigDataRecords, nil
	}

	if group, err := b.groupCache.GetGroupByIDs(address.OrgID, address.GroupID); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("unable to find group id in cache, using default settings")
	} else {
		nsCfg = group.ModuleConfigurations.NSModule
	}

	etld, err := parsers.GetETLD(address.HostAddress)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("unable to get etld, not running bigdata tests")
		return address, bigDataRecords, nil
	}

	// we should already have it, but just in case
	// NOTE: if this is the first time it is seen, only one instance
	// will have this until records are added to the database that other instances will be able to read via GetETLDs
	b.certListener.AddETLD(etld)

	records, err := b.doCTSubdomainAnalysis(ctx, userContext, nsCfg, address, etld)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to do certificate transparency analysis")
	}

	for k, v := range records {
		bigDataRecords[k] = v
	}
	return address, bigDataRecords, nil
}

func (b *BigData) doCTSubdomainAnalysis(ctx context.Context, userContext am.UserContext, nsCfg *am.NSModuleConfig, address *am.ScanGroupAddress, etld string) (map[string]*am.ScanGroupAddress, error) {
	var queryTime time.Time
	subdomains := make(map[string]*am.CTSubdomain, 0)
	records := make(map[string]*am.ScanGroupAddress, 0)

	log.Ctx(ctx).Info().Str("etld", etld).Int("GroupID", address.GroupID).Msg("checking state for etld")
	shouldCT, err := b.st.DoCTDomain(ctx, address.OrgID, address.GroupID, oneHour, etld)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("unable to check do ct domain")
		return records, err
	}

	if !shouldCT {
		log.Ctx(ctx).Info().Msg("not analyzing etld, as it is already complete")
		return records, nil
	}

	retryErr := retrier.Retry(func() error {
		queryTime, subdomains, err = b.bdClient.GetCTSubdomains(ctx, userContext, etld)
		return err
	})

	// if we can't check the database reliably, we don't want to hammer bigquery (costs) so just fail closed here.
	if retryErr != nil {
		log.Ctx(ctx).Warn().Err(retryErr).Msg("unable to get CT records from database, returning")
		return records, nil
	}

	// we've never looked up this etld before
	if subdomains == nil || len(subdomains) == 0 {
		log.Ctx(ctx).Info().Str("etld", etld).Msg("first time etld seen, searching big query")
		queryTime = time.Date(2018, time.May, 0, 0, 0, 0, 0, time.Local)
	} else {
		// we already have results
		log.Ctx(ctx).Info().Str("etld", etld).Msg("DEV MODE: already searched bigquery and data is stored in database")
		// TODO: For Dev mode we are only going to use an exported, reduced, data set
		return b.processCTSubdomainRecords(ctx, nsCfg, address, subdomains, etld)
	}

	// check if we should add new CTSubDomains and update the ctRecords with new records found from bigquery
	allRecords := b.addNewCTSubDomainRecords(ctx, userContext, subdomains, queryTime, etld)
	newAddresses, err := b.processCTSubdomainRecords(ctx, nsCfg, address, allRecords, etld)
	return newAddresses, err
}

// addNewCTSubDomainRecords queries bigquery for subdomains for the etld. If the initial subdomains list was empty/nil
// (because we didn't have any in the db) we set it to the records returned from bigquery. If we did have results,
// we iterate over the new results from bigquery and put them into the subdomains map. Finally, if we had big query
// results, we add the newly identified subdomains to the database via AddCTSubdomains.
func (b *BigData) addNewCTSubDomainRecords(ctx context.Context, userContext am.UserContext, subdomains map[string]*am.CTSubdomain, queryTime time.Time, etld string) map[string]*am.CTSubdomain {
	now := time.Now()
	if now.Sub(queryTime) < defaultCacheTime {
		log.Ctx(ctx).Info().TimeDiff("query_diff", queryTime, time.Now()).Msg("< cacheTime not querying bigquery")
		return subdomains
	}

	bqRecords, err := b.bigQuerier.QuerySubdomains(ctx, queryTime, etld)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("unable to query big query")
		return subdomains
	}

	if subdomains == nil {
		// database results were empty so whatever we got from bigQuerier, is now our result set.
		subdomains = bqRecords
	} else {
		// wasn't nil so we have records we should append to.
		for subdomain := range bqRecords {
			subdomains[subdomain] = &am.CTSubdomain{ETLD: etld, InsertedTime: now.UnixNano(), Subdomain: subdomain}
		}
	}

	if len(bqRecords) == 0 {
		log.Ctx(ctx).Info().Msg("no new records found in bigquery")
		return subdomains
	}

	// if we still have any left, we want to add new ones, and update the last query time.
	devModeQueryTime := time.Date(2019, time.February, 13, 0, 0, 0, 0, time.Local)
	// TODO: change devModeQueryTime to just time.Now()
	// if we still have any left, we want to add new ones, and update the last query time.
	if err := b.bdClient.AddCTSubdomains(ctx, userContext, etld, devModeQueryTime, bqRecords); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to add new ct records and update query time")
	}

	return subdomains
}

func (b *BigData) processCTSubdomainRecords(ctx context.Context, nsCfg *am.NSModuleConfig, address *am.ScanGroupAddress, records map[string]*am.CTSubdomain, etld string) (map[string]*am.ScanGroupAddress, error) {
	newAddresses := make(map[string]*am.ScanGroupAddress, 0)
	newHosts := make(map[string]struct{})

	for record := range records {
		newHosts[record] = struct{}{}
	}

	log.Ctx(ctx).Info().Int("new_hosts", len(newHosts)).Msg("resolving with ResolveNewAddresses")
	newAddresses = module.ResolveNewAddresses(ctx, b.dc, &module.ResolverData{
		Address:           address,
		RequestsPerSecond: int(nsCfg.RequestsPerSecond),
		NewAddresses:      newHosts,
		DiscoveryMethod:   am.DiscoveryBigDataCT,
		Cache:             b.groupCache,
	})
	log.Ctx(ctx).Info().Int("new_addresses", len(newAddresses)).Msg("returning from ResolveNewAddresses")
	return newAddresses, nil
}
