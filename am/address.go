package am

import "context"

const (
	RNAddressAddresses = "lrn:service:address:feature:addresses"
)

// ScanGroupAddress contains details on addresses belonging to the scan group
// for scanning.
type ScanGroupAddress struct {
	AddressID       int64   `json:"address_id"`
	OrgID           int     `json:"org_id"`
	GroupID         int     `json:"group_id"`
	HostAddress     string  `json:"host_address"`
	IPAddress       string  `json:"ip_address"`
	DiscoveryTime   int64   `json:"discovery_time"`
	DiscoveredBy    string  `json:"discovered_by"`
	LastScannedTime int64   `json:"last_scanned_time"`
	LastSeenTime    int64   `json:"last_seen_time"`
	ConfidenceScore float32 `json:"confidence_score"`
	IsSOA           bool    `json:"is_soa"`
	IsWildcardZone  bool    `json:"is_wildcard_zone"`
	IsHostedService bool    `json:"is_hosted_service"`
	Ignored         bool    `json:"ignored"`
}

// ScanGroupAddressFilter filters the results of an Addresses search
type ScanGroupAddressFilter struct {
	OrgID               int   `json:"org_id"`
	GroupID             int   `json:"group_id"`
	WithIgnored         bool  `json:"with_ignored"`
	IgnoredValue        bool  `json:"ignored_value"`
	WithLastScannedTime bool  `json:"with_scanned_time"`
	SinceScannedTime    int64 `json:"since_scanned_time"`
	Start               int64 `json:"start"`
	Limit               int   `json:"limit"`
}

type AddressService interface {
	Get(ctx context.Context, userContext UserContext, filter *ScanGroupAddressFilter) (oid int, addresses []*ScanGroupAddress, err error)
	Count(ctx context.Context, userContext UserContext, groupID int) (oid int, count int, err error)
	Update(ctx context.Context, userContext UserContext, addresses []*ScanGroupAddress) (oid int, count int, err error)
	Delete(ctx context.Context, userContext UserContext, groupID int, addressIDs []int64) (oid int, err error)
}
