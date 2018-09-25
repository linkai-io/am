package ns

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/parsers"
	"github.com/miekg/dns"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/ns/state"
)

const (
	// how long until we remove the zone record saying we've already done ns lookups
	nsExpire = 14400
)

var (
	// ErrEmptyDNSServer missing dns server
	ErrEmptyDNSServer = errors.New("dns_server was empty or invalid")
)

// NS module for extracting NS related information for a scan group.
type NS struct {
	st state.Stater
	dc *dnsclient.Client
	// for closing subscriptions to listen for group updates
	exitContext context.Context
	cancel      context.CancelFunc
	// concurrent safe cache of scan groups updated via Subscribe callbacks
	groupCache *cache.ScanGroupSubscriber
}

// New creates a new NS module for identifying zone information via DNS
// and storing the results in Redis.
func New(st state.Stater) *NS {
	ctx, cancel := context.WithCancel(context.Background())
	ns := &NS{st: st, exitContext: ctx, cancel: cancel}
	// start cache subscriber and listen for updates
	ns.groupCache = cache.NewScanGroupSubscriber(ctx, st)
	return ns
}

// Init the redisclient and dns client.
func (ns *NS) Init(config []byte) error {
	ns.dc = dnsclient.New([]string{"unbound:53"}, 3)
	// populate cache
	return nil
}

// Stop this module from running, and close down subscriptions
func (ns *NS) Stop(ctx context.Context) {
	ns.cancel()
}

// Name returns the module name
func (ns *NS) Name() string {
	return "NS"
}

// Analyze an address, extracts NS, MX, A, AAAA, CNAME records
// TODO: add error if shutting down so dispatcher can retry
func (ns *NS) Analyze(ctx context.Context, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {
	nsRecords := make(map[string]*am.ScanGroupAddress, 0)

	if !ns.shouldAnalyze(address) {
		log.Printf("will not analyze %s %s\n", address.IPAddress, address.HostAddress)
		return address, nsRecords, nil
	}

	log.Printf("analyzing: %s %s\n", address.IPAddress, address.HostAddress)
	resolvedHosts := ns.analyzeHost(ctx, address)
	resolvedIPs := ns.analyzeIP(ctx, address)

	addAddressToMap(nsRecords, resolvedHosts)
	addAddressToMap(nsRecords, resolvedIPs)

	if address.HostAddress == "" {
		return address, nsRecords, nil
	}

	etld, err := parsers.GetETLD(address.HostAddress)
	if err != nil || etld == "" {
		// push nsRecords
		return address, nsRecords, nil
	}

	ok, err := ns.st.DoNSRecords(ctx, address.OrgID, address.GroupID, nsExpire, etld)
	if err != nil {
		log.Printf("unable to do ns records for %s, %s\n", etld, err)
	}

	if ok {
		zoneRecords := ns.analyzeZone(ctx, etld, address)
		log.Printf("got %d zone records for %s\n", len(zoneRecords), etld)
		addAddressToMap(nsRecords, zoneRecords)
	}

	// push nsRecords
	return address, nsRecords, nil
}

// shouldAnalyze determines if we should analyze the specific address or not
func (ns *NS) shouldAnalyze(address *am.ScanGroupAddress) bool {
	if address.IsHostedService || IsHostedDomain(address.HostAddress) {
		return false
	}

	switch uint16(address.NSRecord) {
	case dns.TypeMX, dns.TypeNS, dns.TypeSRV:
		return false
	}
	return true
}

// recordFromAddress creates a new address from this address, copying over the necessary details.
func (ns *NS) newAddress(address *am.ScanGroupAddress, ip, host, discoveredBy string, recordType uint) *am.ScanGroupAddress {
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

// analyzeZone looks up various supporting records for a zone (mx/ns/axfr)
func (ns *NS) analyzeZone(ctx context.Context, zone string, address *am.ScanGroupAddress) []*am.ScanGroupAddress {
	nsData := make([]*am.ScanGroupAddress, 0)

	r, err := ns.dc.LookupMX(zone)
	if err == nil {
		for _, host := range r.Hosts {
			newAddress := ns.newAddress(address, "", host, am.DiscoveryNSQueryOther, uint(r.RecordType))
			nsData = append(nsData, newAddress)
		}
	}

	r, err = ns.dc.LookupNS(zone)
	if err == nil {
		for _, host := range r.Hosts {
			newAddress := ns.newAddress(address, "", host, am.DiscoveryNSQueryOther, uint(r.RecordType))
			nsData = append(nsData, newAddress)
		}
	}

	axfr, err := ns.dc.DoAXFR(zone)
	if err != nil {
		return nsData
	}

	for _, result := range axfr {
		// TODO report axfr ns servers as a finding
		for _, r := range result {
			if len(r.Hosts) == 1 {
				for _, ip := range r.IPs {
					newAddress := ns.newAddress(address, ip, r.Hosts[0], am.DiscoveryNSAXFR, uint(r.RecordType))
					nsData = append(nsData, newAddress)
				}
			} else if len(r.IPs) == 1 {
				for _, host := range r.Hosts {
					newAddress := ns.newAddress(address, r.IPs[0], host, am.DiscoveryNSAXFR, uint(r.RecordType))
					nsData = append(nsData, newAddress)
				}
			}
		}
	}
	return nsData
}

// analyzeIP for this address, finding potentially new hostnames
// if ip == same host, update last seen and scanned time
// if host == empty, add host to the first record
// if ip == different host, create new address
// if host address from address was not returned in any records, try to do a look up??
func (ns *NS) analyzeIP(ctx context.Context, address *am.ScanGroupAddress) []*am.ScanGroupAddress {
	nsData := make([]*am.ScanGroupAddress, 0)

	if address.IPAddress == "" {
		return nsData
	}

	r, err := ns.dc.ResolveIP(address.IPAddress)
	if err != nil || r == nil {
		address.LastScannedTime = time.Now().UnixNano()
		log.Printf("unable to resolve ip: %s\n", err)
		return nsData
	}

	foundOriginal := false

	// we may get multiple hosts back, so check if we've ever found it before?
	for _, host := range r.Hosts {
		// we've seen this same host before *or* never resolved this ip before
		if host == address.HostAddress || address.HostAddress == "" && !foundOriginal {
			foundOriginal = true
			address.HostAddress = host
			if !address.IsHostedService && address.HostAddress != "" {
				address.IsHostedService = IsHostedDomain(address.HostAddress)
			}
			address.LastSeenTime = time.Now().UnixNano()
			address.LastScannedTime = time.Now().UnixNano()
			address.NSRecord = int32(r.RecordType)
			// update the hash address now that we have a proper host for it
			address.AddressHash = convert.HashAddress(address.IPAddress, host)
			continue
		}
		// or we got a new hostname when attempting to resolve this ip.
		// Copy details from original address into the new address
		newAddress := ns.newAddress(address, address.IPAddress, host, am.DiscoveryNSQueryIPToName, uint(r.RecordType))
		nsData = append(nsData, newAddress)
	}

	return nsData
}

// analyzeHost resolves the host address to ips
func (ns *NS) analyzeHost(ctx context.Context, address *am.ScanGroupAddress) []*am.ScanGroupAddress {

	nsData := make([]*am.ScanGroupAddress, 0)
	if address.HostAddress == "" {
		return nsData
	}

	r, err := ns.dc.ResolveName(address.HostAddress)
	if err != nil {
		log.Printf("unable to resolve ip: %s\n", err)
	}

	// we don't need to test for original here because we just take the first one
	// (i == 0 && j == 0)
	for i, rr := range r {
		for j, ip := range rr.IPs {
			if i == 0 && j == 0 {
				if !address.IsHostedService && address.HostAddress != "" {
					address.IsHostedService = IsHostedDomain(address.HostAddress)
				}
				address.IPAddress = ip
				address.LastSeenTime = time.Now().UnixNano()
				address.LastScannedTime = time.Now().UnixNano()
				address.NSRecord = int32(rr.RecordType)
				address.AddressHash = convert.HashAddress(ip, address.HostAddress)
				continue
			}

			newAddress := ns.newAddress(address, ip, address.HostAddress, am.DiscoveryNSQueryNameToIP, uint(rr.RecordType))
			nsData = append(nsData, newAddress)
		}
	}
	return nsData
}

func addAddressToMap(addressMap map[string]*am.ScanGroupAddress, addresses []*am.ScanGroupAddress) {
	for _, addr := range addresses {
		addressMap[addr.AddressHash] = addr
	}
}
