package portscan

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/linkai-io/am/pkg/parsers"

	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/portscanner"

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
	scanner     portscanner.Executor
}

// New port scanner module
func New(scanner portscanner.Executor, dc *dnsclient.Client) *PortScanner {
	ctx, cancel := context.WithCancel(context.Background())
	return &PortScanner{scanner: scanner, exitContext: ctx, cancel: cancel, dc: dc, groupCache: cache.NewScanGroupCache()}
}

// Init the port scanner
func (p *PortScanner) Init(config []byte) error {
	return p.scanner.Init(nil)
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
	var err error
	var group *am.ScanGroup

	ctx = module.DefaultLogger(ctx, userContext, address)

	if group = p.groupCache.GetByIDs(address.OrgID, address.GroupID); group == nil {
		return nil, nil, am.ErrScanGroupNotExists
	}
	if group.ModuleConfigurations == nil || group.ModuleConfigurations.PortModule == nil {
		return nil, nil, am.ErrEmptyModuleConfig
	}

	cfg := group.ModuleConfigurations.PortModule

	targetIP := address.IPAddress
	hostAddress := address.HostAddress

	if hostAddress == "" {
		hostAddress = address.IPAddress
	} else {
		if address, err = p.getTargetIPv4(ctx, address); err != nil {
			return nil, nil, err
		}
		targetIP = address.IPAddress
	}

	if targetIP == "" {
		return nil, nil, am.ErrEmptyIP
	}

	if parsers.IsBannedIP(targetIP) {
		return nil, nil, am.ErrBannedIP
	}

	log.Ctx(ctx).Info().Str("ip_address", targetIP).Msg("scanning now")
	start := time.Now()
	results, err := p.scanner.PortScan(ctx, targetIP, int(cfg.RequestsPerSecond), cfg.TCPPorts)
	if err != nil {
		return nil, nil, err
	}

	log.Ctx(ctx).Info().TimeDiff("scan_time", time.Now(), start).Msg("scan completed")
	portResults := &am.PortResults{
		OrgID:       address.OrgID,
		GroupID:     address.GroupID,
		HostAddress: hostAddress,
		Ports: &am.Ports{
			Current: &am.PortData{
				IPAddress: targetIP,
				TCPPorts:  results.Open,
			},
		},
		ScannedTimestamp:         start.UnixNano(),
		PreviousScannedTimestamp: 0,
	}
	return address, portResults, nil
}

// getTargetIPv4 grabs the first valid ipv4 address
func (p *PortScanner) getTargetIPv4(ctx context.Context, address *am.ScanGroupAddress) (*am.ScanGroupAddress, error) {
	log.Ctx(ctx).Info().Str("host_address", address.HostAddress).Msg("resolving")
	results, err := p.dc.ResolveName(ctx, address.HostAddress)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, errors.New("unable to resolve address")
	}

	log.Ctx(ctx).Info().Msgf("Results: %#v %d", results, len(results))

	for _, result := range results {
		log.Ctx(ctx).Info().Msgf("got result %#v", result)
		// ipv4 for now :/
		if result.RecordType == dns.TypeAAAA {
			continue
		}

		// iterate once to see if the returned IP matches what our scangroup address is.
		for _, ip := range result.IPs {
			log.Ctx(ctx).Info().Msgf("IPS: %v", ip)
			// if the IP isn't empty and matches the original scangroup address, just return that.
			if ip != "" && ip == address.IPAddress {
				log.Ctx(ctx).Info().Msg("ip address returned matches original scangroupaddress")
				return address, nil
			}
		}

		// all results are different ips, time to make a new address and return that instead
		for _, ip := range result.IPs {

			ipAddr := net.ParseIP(ip)
			if ipAddr == nil {
				continue
			}
			if ipAddr.To4() != nil {
				if parsers.IsBannedIP(ip) {
					continue
				}
				newAddress := module.NewAddressFromDNS(address, ip, address.HostAddress, am.DiscoveryNSQueryNameToIP, uint(result.RecordType))
				newAddress.ConfidenceScore = module.CalculateConfidence(ctx, address, newAddress)
				log.Ctx(ctx).Info().Msg("new address returned and created")
				return newAddress, nil
			}
		}
	}
	return nil, errors.New("no IPv4 addresses returned")
}
