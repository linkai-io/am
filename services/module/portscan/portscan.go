package portscan

import (
	"context"

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
	exitContext context.Context
	cancel      context.CancelFunc
}

// New brute force module
func New(dc *dnsclient.Client) *PortScanner {
	ctx, cancel := context.WithCancel(context.Background())
	b := &PortScanner{exitContext: ctx, cancel: cancel}
	b.dc = dc
	return b
}

// Init the brute forcer with the initial input subdomain list
func (p *PortScanner) Init(config []byte) error {

	return nil
}

// Analyze will attempt port scan
func (p *PortScanner) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, error) {
	ctx = module.DefaultLogger(ctx, userContext, address)

	return address, nil
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
