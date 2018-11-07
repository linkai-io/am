package module

import (
	"context"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/pkg/parsers"
	"github.com/rs/zerolog/log"
)

// NewAddress creates a new address from this address, copying over the necessary details.
func NewAddressFromDNS(address *am.ScanGroupAddress, ip, host, discoveredBy string, recordType uint) *am.ScanGroupAddress {
	newAddress := &am.ScanGroupAddress{
		OrgID:           address.OrgID,
		GroupID:         address.GroupID,
		DiscoveryTime:   time.Now().UnixNano(),
		DiscoveredBy:    discoveredBy,
		LastSeenTime:    time.Now().UnixNano(),
		IPAddress:       ip,
		HostAddress:     host,
		IsHostedService: address.IsHostedService,
		NSRecord:        int32(recordType),
		AddressHash:     convert.HashAddress(ip, host),
		FoundFrom:       address.AddressHash,
	}

	if !address.IsHostedService && address.HostAddress != "" {
		newAddress.IsHostedService = IsHostedDomain(newAddress.HostAddress)
	}
	return newAddress
}

// AddAddressToMap from slice
func AddAddressToMap(addressMap map[string]*am.ScanGroupAddress, addresses []*am.ScanGroupAddress) {
	for _, addr := range addresses {
		addressMap[addr.AddressHash] = addr
	}
}

// CalculateConfidence of the new addresses
func CalculateConfidence(ctx context.Context, address, newAddress *am.ScanGroupAddress) float32 {
	origTLD, err := parsers.GetETLD(address.HostAddress)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("unable to get tld of original address")
		return 0
	}

	newTLD, err := parsers.GetETLD(newAddress.HostAddress)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("unable to get tld of new address")
		return 0
	}

	if origTLD == newTLD {
		return address.ConfidenceScore
	}
	return 0
}

type results struct {
	R        []*dnsclient.Results
	Hostname string
	Err      error
}

type ResolverData struct {
	Address           *am.ScanGroupAddress
	RequestsPerSecond int
	NewAddresses      map[string]struct{}
	DiscoveryMethod   string
}

// ResolveNewAddresses is a generic resolver function for looking up hostnames to ip addresses and collecting them as a map to return
// to caller
func ResolveNewAddresses(ctx context.Context, dns *dnsclient.Client, data *ResolverData) map[string]*am.ScanGroupAddress {
	newRecords := make(map[string]*am.ScanGroupAddress, 0)

	numHosts := len(data.NewAddresses)
	rps := data.RequestsPerSecond
	if numHosts < rps {
		rps = numHosts
	}
	pool := workerpool.New(rps)

	out := make(chan *results, numHosts)

	// submit all hosts to our worker pool
	for newHost := range data.NewAddresses {
		task := func(ctx context.Context, host string, out chan<- *results) func() {
			return func() {
				r, err := dns.ResolveName(ctx, host)
				out <- &results{Hostname: host, R: r, Err: err}
			}
		}
		h := newHost
		pool.Submit(task(ctx, h, out))
	}

	pool.StopWait()
	close(out)

	log.Ctx(ctx).Info().Msg("all tasks completed")

	for result := range out {
		if result.Err != nil {
			continue
		}

		for _, rr := range result.R {
			for _, ip := range rr.IPs {
				log.Ctx(ctx).Info().Str("hostname", result.Hostname).Str("ip_address", ip).Msg("found new record")
				newAddress := NewAddressFromDNS(data.Address, ip, result.Hostname, data.DiscoveryMethod, uint(rr.RecordType))
				newAddress.ConfidenceScore = CalculateConfidence(ctx, data.Address, newAddress)
				newRecords[newAddress.AddressHash] = newAddress
			}
		}
	}

	return newRecords
}
