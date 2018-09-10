package ns

import (
	"errors"
	"log"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/modules/ns/state"
	"github.com/linkai-io/am/pkg/dnsclient"
)

var (
	// ErrEmptyDNSServer missing dns server
	ErrEmptyDNSServer = errors.New("dns_server was empty or invalid")
)

// NS module for extracting NS related information for an input list.
type NS struct {
	st state.Stater
	dc *dnsclient.Client
}

// New creates a new NS module for identifying zone information via DNS
// and storing the results in Redis.
func New(st state.Stater) *NS {
	return &NS{st: st}
}

// Init the redisclient and dns client.
func (ns *NS) Init(config []byte) error {
	ns.dc = dnsclient.New([]string{"0.0.0.0:2053"}, 2)
	return nil
}

// Name returns the module name
func (ns *NS) Name() string {
	return "NS"
}

// Analyze a domain zone, extracts NS, MX, A, AAAA, CNAME records
func (ns *NS) Analyze(address *am.ScanGroupAddress) {
	ns.analyzeHost(address)
	ns.analyzeIP(address)

	return
}

// analyzeIP for this address, finding potentially new hostnames
// if ip == same host, update last seen and scanned time
// if host == empty, add host to the first record
// if ip == different host, create new address
// if host address from address was not returned in any records, try to do a look up??
func (ns *NS) analyzeIP(address *am.ScanGroupAddress) []*am.NSData {
	r, err := ns.dc.ResolveIP(address.IPAddress)
	if err != nil {
		address.LastScannedTime = time.Now().UnixNano()
		log.Printf("unable to resolve ip: %s\n", err)
	}

	nsData := make([]*am.NSData, 0)

	foundOriginal := false
	for _, host := range r.Hosts {
		// we've seen this same host before *or* never resolved this ip before
		if host == address.HostAddress || address.HostAddress == "" {
			foundOriginal = true
			address.HostAddress = host
			address.LastSeenTime = time.Now().UnixNano()
			address.LastScannedTime = time.Now().UnixNano()
			nsRecord := &am.NSData{
				ScanGroupAddress: *address,
				NSRecordType:     uint(r.RecordType),
			}
			nsData = append(nsData, nsRecord)
			continue
		}
		// or we got a new hostname when attempting to resolve this ip.
		// Copy details from original address into the new address
		newAddress := &am.ScanGroupAddress{
			OrgID:         address.OrgID,
			GroupID:       address.GroupID,
			DiscoveryTime: time.Now().UnixNano(),
			DiscoveredBy:  am.DiscoveryNSQueryIPToName,
			LastSeenTime:  time.Now().UnixNano(),
			IPAddress:     address.IPAddress,
			HostAddress:   host,
		}

		nsRecord := &am.NSData{
			ScanGroupAddress: *newAddress,
			NSRecordType:     uint(r.RecordType),
		}

		nsData = append(nsData, nsRecord)
		log.Printf("AnalyzeIP found new address: %v from: %d\n", newAddress, address.AddressID)
	}

	if !foundOriginal {
		ns.analyzeHost(address)
	}

	return nsData
}

func (ns *NS) analyzeHost(address *am.ScanGroupAddress) {
	r, err := ns.dc.ResolveName(address.HostAddress)
	if err != nil {
		log.Printf("unable to resolve ip: %s\n", err)
	}
	for i, rr := range r {
		for j, ip := range rr.IPs {
			if i == 0 && j == 0 {
				address.IPAddress = ip
				address.LastSeenTime = time.Now().UnixNano()
				address.LastScannedTime = time.Now().UnixNano()
				continue
			}
			newAddress := &am.ScanGroupAddress{
				OrgID:         address.OrgID,
				GroupID:       address.GroupID,
				DiscoveryTime: time.Now().UnixNano(),
				DiscoveredBy:  am.DiscoveryNSQueryNameToIP,
				LastSeenTime:  time.Now().UnixNano(),
				IPAddress:     ip,
				HostAddress:   address.HostAddress,
			}
			log.Printf("AnalyzeHost found new address: %v from: %d\n", newAddress, address.AddressID)
		}
	}
}
