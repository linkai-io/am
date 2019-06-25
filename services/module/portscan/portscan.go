package portscan

import (
	"context"

	"github.com/linkai-io/am/pkg/cache"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
)

const (
	oneHour     = 60 * 60
	threeHours  = oneHour * 3
	fiveMinutes = 60 * 5
)

// PortScanner will port scan all 100% confidence hosts (and if ip only, IPs)
type PortScanner struct {
	dc          *dnsclient.Client
	groupCache  *cache.ScanGroupCache
	exitContext context.Context
	cancel      context.CancelFunc
}

// New port scanner module
func New(dc *dnsclient.Client) *PortScanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &PortScanner{exitContext: ctx, cancel: cancel, dc: dc, groupCache: cache.NewScanGroupCache()}
}

// Init the port scanner
func (p *PortScanner) Init(config []byte) error {

	return nil
}

// AddGroup on start of a group analysis (before any addresses come in)
func (p *PortScanner) AddGroup(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) error {
	key := p.groupCache.MakeGroupKey(group.OrgID, group.GroupID)
	p.groupCache.Put(key, group)
	return nil
}

// RemoveGroup on end of group analysis
func (p *PortScanner) RemoveGroup(ctx context.Context, userContext am.UserContext, orgID, groupID int) error {
	key := p.groupCache.MakeGroupKey(orgID, groupID)
	p.groupCache.Clear(key)
	return nil
}

// Analyze will attempt port scan
func (p *PortScanner) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, *am.PortResults, error) {
	ctx = module.DefaultLogger(ctx, userContext, address)
	var group *am.ScanGroup

	if group = p.groupCache.GetByIDs(address.OrgID, address.GroupID); group == nil {
		log.Ctx(ctx).Warn().Err(am.ErrScanGroupNotExists).Msg("unable to find group id in cache, returning")
		return nil, nil, am.ErrScanGroupNotExists
	}
	return address, nil, nil
}

// shouldAnalyze determines if we should analyze the specific address or not. Updates address.IsWildcardZone
// if tested.
func (p *PortScanner) shouldAnalyze(ctx context.Context, address *am.ScanGroupAddress) bool {
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
		log.Ctx(ctx).Info().Float32("confidence", address.ConfidenceScore).Msg("score too low")
		return false
	}

	return true
}
