package brute

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/linkai-io/am/services/module"

	"github.com/linkai-io/am/pkg/bq"
	"github.com/linkai-io/am/pkg/retrier"

	"github.com/linkai-io/am/pkg/parsers"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/bigdata/state"
	"github.com/miekg/dns"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	defaultCacheTime = time.Hour * 4
	oneHour          = 60 * 60
)

// BigData will query our locally cached results of big data first, then the bigquery database
// if we should need updated values looking for sub domains of a etld.
type BigData struct {
	st         state.Stater
	dc         *dnsclient.Client
	bdClient   am.BigDataService
	bigQuerier bq.BigQuerier
	subdomains []string

	// for closing subscriptions to listen for group updates
	exitContext context.Context
	cancel      context.CancelFunc
	// concurrent safe cache of scan groups updated via Subscribe callbacks
	groupCache *cache.ScanGroupSubscriber
}

// New big data analysis module
func New(dc *dnsclient.Client, st state.Stater, bdClient am.BigDataService, bqQuerier bq.BigQuerier) *BigData {
	ctx, cancel := context.WithCancel(context.Background())
	b := &BigData{exitContext: ctx, cancel: cancel}
	b.st = st
	b.dc = dc
	b.bdClient = bdClient
	b.bigQuerier = bqQuerier
	// start cache subscriber and listen for updates
	b.groupCache = cache.NewScanGroupSubscriber(ctx, st)
	return b
}

// Init the bigdata service with bigquery details
func (b *BigData) Init() error {

	return nil
}

func (b *BigData) defaultModuleConfig() *am.NSModuleConfig {
	return &am.NSModuleConfig{
		RequestsPerSecond: 10,
	}
}

// Analyze will attempt to find additional domains by looking in various public data sets we've collected. We only
// do this for ETLDs.
func (b *BigData) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {
	nsCfg := b.defaultModuleConfig()
	logger := log.With().
		Int("OrgID", userContext.GetOrgID()).
		Int("UserID", userContext.GetUserID()).
		Str("TraceID", userContext.GetTraceID()).
		Str("IPAddress", address.IPAddress).
		Str("HostAddress", address.HostAddress).
		Int64("AddressID", address.AddressID).
		Str("AddressHash", address.AddressHash).Logger()

	bigDataRecords := make(map[string]*am.ScanGroupAddress, 0)

	if !b.shouldAnalyze(ctx, logger, address) {
		logger.Info().Msg("not analyzing")
		return address, bigDataRecords, nil
	}

	if group, err := b.groupCache.GetGroupByIDs(address.OrgID, address.GroupID); err != nil {
		logger.Warn().Err(err).Msg("unable to find group id in cache, using default settings")
	} else {
		nsCfg = group.ModuleConfigurations.NSModule
	}

	etld, err := parsers.GetETLD(address.HostAddress)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to get etld, not running bigdata tests")
		return address, bigDataRecords, nil
	}

	records, err := b.doCTAnalysis(ctx, userContext, logger, nsCfg, address, etld)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to do certificate transparency analysis")
	}

	for k, v := range records {
		bigDataRecords[k] = v
	}
	return address, bigDataRecords, nil
}

func (b *BigData) doCTAnalysis(ctx context.Context, userContext am.UserContext, logger zerolog.Logger, nsCfg *am.NSModuleConfig, address *am.ScanGroupAddress, etld string) (map[string]*am.ScanGroupAddress, error) {
	ctRecords := make(map[string]*am.CTRecord, 0)

	records := make(map[string]*am.ScanGroupAddress, 0)

	shouldCT, err := b.st.DoCTDomain(ctx, address.OrgID, address.GroupID, oneHour, etld)
	if err != nil {
		return records, err
	}

	if !shouldCT {
		logger.Info().Msg("not analyzing etld, as it is already complete")
		return records, nil
	}

	var queryTime time.Time

	retryErr := retrier.Retry(func() error {
		queryTime, ctRecords, err = b.bdClient.GetCT(ctx, userContext, etld)
		return err
	})

	// if we can't check the database reliably, we don't want to hammer bigquery (costs) so just fail closed here.
	if retryErr != nil {
		logger.Warn().Msg("unable to get CT records from database, returning")
		return records, nil
	}

	// we've never looked up this etld before
	if len(ctRecords) == 0 {
		queryTime = time.Date(2018, time.May, 0, 0, 0, 0, 0, nil)
	}

	// check if we should add new CTRecords and update the ctRecords with new records found from bigquery
	b.addNewCTRecords(ctx, userContext, logger, ctRecords, queryTime, etld)
	newAddresses, err := b.processCTRecords(ctx, nsCfg, logger, address, ctRecords, etld)
	return newAddresses, err
}

// addNewCTRecords queries BigQuery to see if we have any new records for this etld, provided that now - queryTime is > cachetime (default 4 hours).
// Note, this *DOES* modify ctRecords by adding the bigquery results to it.
func (b *BigData) addNewCTRecords(ctx context.Context, userContext am.UserContext, logger zerolog.Logger, ctRecords map[string]*am.CTRecord, queryTime time.Time, etld string) {

	if time.Now().Sub(queryTime) < defaultCacheTime {
		logger.Info().TimeDiff("query_diff", queryTime, time.Now()).Msg("< cacheTime not querying bigquery")
		return
	}

	bqRecords, err := b.bigQuerier.QueryETLD(ctx, queryTime, etld)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to query big query")
		return
	}

	if ctRecords != nil {
		for hash, record := range bqRecords {
			if _, ok := ctRecords[hash]; ok {
				// already exists, remove it from our bigquery records map
				delete(bqRecords, hash)
				continue
			}
			// add the new record to our ct records map
			ctRecords[hash] = record
		}
	}

	if len(bqRecords) == 0 {
		log.Info().Msg("no new records found in bigquery")
		return
	}

	// if we still have any left, we want to add new ones, and update the last query time.
	if err := b.bdClient.AddCT(ctx, userContext, etld, time.Now(), bqRecords); err != nil {
		logger.Error().Err(err).Msg("failed to add new ct records and update query time")
	}
}

func (b *BigData) processCTRecords(ctx context.Context, nsCfg *am.NSModuleConfig, logger zerolog.Logger, address *am.ScanGroupAddress, records map[string]*am.CTRecord, etld string) (map[string]*am.ScanGroupAddress, error) {
	newAddresses := make(map[string]*am.ScanGroupAddress, 0)

	newHosts := make(map[string]struct{})

	needle, err := regexp.Compile("(?i)" + etld)
	if err != nil {
		return newAddresses, err
	}

	needles := make([]*regexp.Regexp, 1)
	needles[0] = needle
	for _, record := range records {
		allHosts := strings.Join([]string{record.CommonName, record.VerifiedDNSNames, record.UnverifiedDNSNames}, " ")
		recordHosts := parsers.ExtractHostsFromResponse(needles, allHosts)
		for k, v := range recordHosts {
			newHosts[k] = v
		}
	}

	newAddresses = module.ResolveNewAddresses(ctx, logger, b.dc, &module.ResolverData{
		Address:           address,
		RequestsPerSecond: int(nsCfg.RequestsPerSecond),
		NewAddresses:      newHosts,
		DiscoveryMethod:   am.DiscoveryBigDataCT,
	})
	return newAddresses, nil
}

// shouldAnalyze determines if we should analyze the specific address or not.
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
