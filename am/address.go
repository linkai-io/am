package am

import "context"

const (
	RNAddressAddresses = "lrn:service:address:feature:addresses"
)

// ScanGroupAddress contains details on addresses belonging to the scan group
// for scanning.
type ScanGroupAddress struct {
	AddressID       int64  `json:"address_id"`
	OrgID           int    `json:"org_id"`
	GroupID         int    `json:"group_id"`
	HostAddress     string `json:"host_address"`
	IPAddress       string `json:"ip_address"`
	DiscoveryTime   int64  `json:"discovery_time"`
	DiscoveredBy    string `json:"discovered_by"`
	LastJobID       int64  `json:"last_job_id"`
	LastSeenTime    int64  `json:"last_seen_time"`
	IsSOA           bool   `json:"is_soa"`
	IsWildcardZone  bool   `json:"is_wildcard_zone"`
	IsHostedService bool   `json:"is_hosted_service"`
	Ignored         bool   `json:"ignored"`
}

// ScanGroupAddressFilter filters the results of an Addresses search
type ScanGroupAddressFilter struct {
	GroupID      int  `json:"group_id"`
	WithIgnored  bool `json:"with_ignored"`
	IgnoredValue bool `json:"ignored_value"`
	Start        int  `json:"start"`
	Limit        int  `json:"limit"`
}

type AddressService interface {
	Addresses(ctx context.Context, userContext UserContext, filter *ScanGroupAddressFilter) (oid int, addresses []*ScanGroupAddress, err error)
	AddressCount(ctx context.Context, userContext UserContext, groupID int) (oid int, count int, err error)
	Update(ctx context.Context, userContext UserContext, addresses []*ScanGroupAddress) (oid int, count int, err error)
	Delete(ctx context.Context, userContext UserContext, groupID int, addressIDs []int64) (oid int, err error)
}
