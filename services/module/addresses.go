package module

import (
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/parsers"
	"github.com/rs/zerolog"
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
func CalculateConfidence(logger zerolog.Logger, address, newAddress *am.ScanGroupAddress) float32 {
	origTLD, err := parsers.GetETLD(address.HostAddress)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to get tld of original address")
		return 0
	}

	newTLD, err := parsers.GetETLD(newAddress.HostAddress)
	if err != nil {
		logger.Warn().Err(err).Msg("unable to get tld of new address")
		return 0
	}

	if origTLD == newTLD {
		return address.ConfidenceScore
	}
	return 0
}
