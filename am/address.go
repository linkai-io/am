package am

import (
	"context"
	"time"
)

const (
	RNAddressAddresses = "lrn:service:address:feature:addresses"
	AddressServiceKey  = "addressservice"
)

const (
	FilterIgnored                = "ignored"
	FilterWildcard               = "wildcard"
	FilterHosted                 = "hosted"
	FilterAfterScannedTime       = "after_scanned_time"
	FilterBeforeScannedTime      = "before_scanned_time"
	FilterAfterSeenTime          = "after_seen_time"
	FilterBeforeSeenTime         = "before_seen_time"
	FilterAfterDiscoveredTime    = "after_discovered_time"
	FilterBeforeDiscoveredTime   = "before_discovered_time"
	FilterAboveConfidence        = "above_confidence"
	FilterBelowConfidence        = "below_confidence"
	FilterEqualsConfidence       = "equals_confidence"
	FilterAboveUserConfidence    = "above_user_confidence"
	FilterBelowUserConfidence    = "below_user_confidence"
	FilterEqualsUserConfidence   = "equals_user_confidence"
	FilterEqualsNSRecord         = "ns_record"
	FilterNotNSRecord            = "not_ns_record"
	FilterIPAddress              = "ip_address"
	FilterNotIPAddress           = "not_ip_address"
	FilterHostAddress            = "host_address"
	FilterNotHostAddress         = "not_host_address"
	FilterEndsHostAddress        = "ends_host_address"
	FilterNotEndsHostAddress     = "not_ends_host_address"
	FilterStartsHostAddress      = "starts_host_address"
	FilterNotStartsHostAddress   = "not_starts_host_address"
	FilterContainsHostAddress    = "contains_host_address"
	FilterNotContainsHostAddress = "not_contains_host_address"
)

/*
(1, 'input_list'),
    (2, 'manual'),
    (3, 'other'),
    -- ns analyzer module 100-200
    (100, 'ns_query_other'),
    (101, 'ns_query_ip_to_name'),
	(102, 'ns_query_name_to_ip'),
	(103, 'dns_axfr'),
    -- dns brute module 200-300
    (200, 'dns_brute_forcer'),
    (201, 'dns_mutator'),
    -- web modules 300 - 999
    (300, 'web_crawler'),
	-- other, feature modules
	(400, 'bigdata'),
	(401, 'bigdata_certificate_transparency'),
	(1000, 'git_hooks');
*/
const (
	DiscoveryNSInputList     = "input_list"
	DiscoveryNSManual        = "manual"
	DiscoveryNSQueryOther    = "ns_query_other"
	DiscoveryNSQueryIPToName = "ns_query_ip_to_name"
	DiscoveryNSQueryNameToIP = "ns_query_name_to_ip"
	DiscoveryNSAXFR          = "ns_query_axfr"
	DiscoveryNSSECWalk       = "ns_query_nsec_walk"
	DiscoveryBruteSubDomain  = "dns_brute_forcer"
	DiscoveryBruteMutator    = "dns_mutator"
	DiscoveryWebCrawler      = "web_crawler"
	DiscoveryGitHooks        = "git_hooks"
	DiscoveryBigData         = "bigdata"
	DiscoveryBigDataCT       = "bigdata_certificate_transparency"
)

// ScanGroupAddress contains details on addresses belonging to the scan group
// for scanning.
type ScanGroupAddress struct {
	AddressID           int64   `json:"address_id"`
	OrgID               int     `json:"org_id"`
	GroupID             int     `json:"group_id"`
	HostAddress         string  `json:"host_address"`
	IPAddress           string  `json:"ip_address"`
	DiscoveryTime       int64   `json:"discovery_time"`
	DiscoveredBy        string  `json:"discovered_by"`
	LastScannedTime     int64   `json:"last_scanned_time"`
	LastSeenTime        int64   `json:"last_seen_time"`
	ConfidenceScore     float32 `json:"confidence_score"`
	UserConfidenceScore float32 `json:"user_confidence_score"`
	IsSOA               bool    `json:"is_soa"`
	IsWildcardZone      bool    `json:"is_wildcard_zone"`
	IsHostedService     bool    `json:"is_hosted_service"`
	Ignored             bool    `json:"ignored"`
	FoundFrom           string  `json:"found_from"` // address hash it was discovered from
	NSRecord            int32   `json:"ns_record"`
	AddressHash         string  `json:"address_hash"`
	Deleted             bool    `json:"deleted"`
}

type ScanGroupHostList struct {
	OrgID       int          `json:"org_id"`
	GroupID     int          `json:"group_id"`
	ETLD        string       `json:"etld"`
	HostAddress string       `json:"host_address"` // or ip address if no hostname
	AddressIDs  []int64      `json:"address_ids"`
	IPAddresses []string     `json:"ip_addresses"`
	Ports       *PortResults `json:"ports,omitempty"`
}

// ScanGroupAddressFilter filters the results of an Addresses search
type ScanGroupAddressFilter struct {
	OrgID   int         `json:"org_id"`
	GroupID int         `json:"group_id"`
	Start   int64       `json:"start"`
	Limit   int         `json:"limit"`
	Filters *FilterType `json:"filters"`
}

type ScanGroupAggregates struct {
	Time  []int64 `json:"time"`
	Count []int32 `json:"count"`
}

// ScanGroupAddressStats general statistics for scan group addresses
type ScanGroupAddressStats struct {
	OrgID             int                             `json:"org_id"`
	GroupID           int                             `json:"group_id"`
	DiscoveredBy      []string                        `json:"discovered_by"`
	DiscoveredByCount []int32                         `json:"discovered_by_count"`
	Aggregates        map[string]*ScanGroupAggregates `json:"aggregates"`
	Total             int32                           `json:"total"`
	ConfidentTotal    int32                           `json:"confident_total"`
}

// AddressService manages all asset data
type AddressService interface {
	Init(config []byte) error
	Get(ctx context.Context, userContext UserContext, filter *ScanGroupAddressFilter) (oid int, addresses []*ScanGroupAddress, err error)
	OrgStats(ctx context.Context, userContext UserContext) (oid int, orgStats []*ScanGroupAddressStats, err error)
	GroupStats(ctx context.Context, userContext UserContext, groupID int) (oid int, groupStats *ScanGroupAddressStats, err error)
	GetHostList(ctx context.Context, userContext UserContext, filter *ScanGroupAddressFilter) (oid int, hostList []*ScanGroupHostList, err error)
	Count(ctx context.Context, userContext UserContext, groupID int) (oid int, count int, err error)
	Update(ctx context.Context, userContext UserContext, addresses map[string]*ScanGroupAddress) (oid int, count int, err error)
	UpdateHostPorts(ctx context.Context, userContext UserContext, address *ScanGroupAddress, portResults *PortResults) (oid int, err error)
	Delete(ctx context.Context, userContext UserContext, groupID int, addressIDs []int64) (oid int, err error)
	Ignore(ctx context.Context, userContext UserContext, groupID int, addressIDs []int64, ignoreValue bool) (oid int, err error)
	Archive(ctx context.Context, userContext UserContext, group *ScanGroup, archiveTime time.Time) (int, int, error)
}
