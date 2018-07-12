package am

import (
	"context"
)

const (
	RNScanGroupGroups    = "lrn:service:scangroup:feature:groups"
	RNScanGroupAddresses = "lrn:service:scangroup:feature:addresses"
)

// ModuleConfiguration contains all the module configurations
type ModuleConfiguration struct {
	NSModule    *NSModuleConfig    `json:"ns_module"`
	BruteModule *BruteModuleConfig `json:"brute_module"`
	PortModule  *PortModuleConfig  `json:"port_module"`
	WebModule   *WebModuleConfig   `json:"web_module"`
}

// ScanGroup is an initial scan grouping configuration that has the original
// input file along with owner related information
type ScanGroup struct {
	OrgID                int                  `json:"org_id"`
	GroupID              int                  `json:"group_id"`
	GroupName            string               `json:"group_name"`
	CreationTime         int64                `json:"creation_time"`
	CreatedBy            int                  `json:"created_by"`
	ModifiedBy           int                  `json:"modified_by"`
	ModifiedTime         int64                `json:"modified_time"`
	OriginalInput        []byte               `json:"original_input"`
	ModuleConfigurations *ModuleConfiguration `json:"module_configurations"`
	Deleted              bool                 `json:"deleted"`
}

// ScanGroupAddress contains details on addresses belonging to the scan group
// for scanning.
type ScanGroupAddress struct {
	AddressID       int64  `json:"address_id"`
	OrgID           int    `json:"org_id"`
	GroupID         int    `json:"group_id"`
	Address         string `json:"address"`
	ConfigurationID int    `json:"configuration_id"`
	AddedTime       int64  `json:"added_time"`
	AddedBy         string `json:"added_by"`
	Ignored         bool   `json:"ignored"`
}

// FailedAddress is used when we are unable to add or update addresses
type FailedAddress struct {
	Address      string `json:"address"`
	FailedReason string `json:"failed_reason"`
}

// ScanGroupService manages input lists and configurations for an organization and group. OrgIDs should
// always be returned for ensuring data integrity for requesters
type ScanGroupService interface {
	Init(config []byte) error
	IsAuthorized(ctx context.Context, userContext UserContext, resource, action string) bool
	Get(ctx context.Context, userContext UserContext, groupID int) (oid int, group *ScanGroup, err error)
	GetByName(ctx context.Context, userContext UserContext, groupName string) (oid int, group *ScanGroup, err error)
	Groups(ctx context.Context, userContext UserContext) (oid int, groups []*ScanGroup, err error)
	Create(ctx context.Context, userContext UserContext, newGroup *ScanGroup) (oid int, gid int, err error)
	Update(ctx context.Context, userContext UserContext, group *ScanGroup) (oid int, gid int, err error)
	Delete(ctx context.Context, userContext UserContext, groupID int) (oid int, gid int, err error)
	Addresses(ctx context.Context, userContext UserContext, groupID int) (oid int, addresses []*ScanGroupAddress, err error)
	AddAddresses(ctx context.Context, userContext UserContext, addresses []*ScanGroupAddress) (oid int, failed []*FailedAddress, err error)
	UpdateAddresses(ctx context.Context, userContext UserContext, addresses []*ScanGroupAddress) (oid int, failed []*FailedAddress, err error)
}
