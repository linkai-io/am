package am

import (
	"context"
)

type ScanGroupAction int

// List of possible action types for the scan group service for authorization purposes
const (
	GetGroup ScanGroupAction = iota
	GetGroups
	CreateGroup
	DeleteGroup
	GetVersion
	CreateVersion
	DeleteVersion
	GetAddresses
	AddAddresses
	UpdatedAddresses
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
	OrgID         int32  `json:"org_id"`
	GroupID       int32  `json:"group_id"`
	GroupName     string `json:"group_name"`
	CreationTime  int64  `json:"creation_time"`
	CreatedBy     int32  `json:"created_by"`
	OriginalInput []byte `json:"original_input"`
	Deleted       bool   `json:"deleted"`
}

// ScanGroupVersion tracks versions of scan group configurations to support
// adding and removing hosts, and changing module configurations
type ScanGroupVersion struct {
	OrgID                int32                `json:"org_id"`
	GroupID              int32                `json:"group_id"`
	GroupVersionID       int32                `json:"group_version_id"`
	VersionName          string               `json:"version_name"`
	CreationTime         int64                `json:"creation_time"`
	CreatedBy            int32                `json:"created_by"`
	ModuleConfigurations *ModuleConfiguration `json:"module_configurations"`
	Deleted              bool                 `json:"deleted"`
}

// ScanGroupAddress contains details on addresses belonging to the scan group
// for scanning.
type ScanGroupAddress struct {
	AddressID int64                `json:"address_id"`
	OrgID     int32                `json:"org_id"`
	GroupID   int32                `json:"group_id"`
	Address   string               `json:"address"`
	Settings  *ModuleConfiguration `json:"settings"`
	AddedTime int64                `json:"added_time"`
	AddedBy   string               `json:"added_by"`
	Ignored   bool                 `json:"ignored"`
}

// FailedAddress is used when we are unable to add or update addresses
type FailedAddress struct {
	Address      string `json:"address"`
	FailedReason string `json:"failed_reason"`
}

// Input represents a parsed and validated input
type Input map[string]interface{}

// ScanGroupService manages input lists and configurations for an organization and group. OrgIDs should
// always be returned for ensuring data integrity for requesters
type ScanGroupService interface {
	Init(config []byte) error
	Get(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, group *ScanGroup, err error)
	Create(ctx context.Context, orgID, requesterUserID int32, newGroup *ScanGroup, newVersion *ScanGroupVersion) (oid int32, gid int32, err error)
	Delete(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, gid int32, err error)
	GetVersion(ctx context.Context, orgID, requesterUserID, groupID, groupVersionID int32) (oid int32, groupVersion *ScanGroupVersion, err error)
	CreateVersion(ctx context.Context, orgID, requesterUserID int32, scanGroupVersion *ScanGroupVersion) (oid int32, gid int32, gvid int32, err error)
	DeleteVersion(ctx context.Context, orgID, requesterUserID, groupID, groupVersionID int32, versionName string) (oid int32, gid int32, gvid int32, err error)
	Groups(ctx context.Context, orgID int32) (oid int32, groups []*ScanGroup, err error)
	Addresses(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, addresses []*ScanGroupAddress, err error)
	AddAddresses(ctx context.Context, orgID, requesterUserID int32, addresses []*ScanGroupAddress) (oid int32, failed []*FailedAddress, err error)
	UpdateAddresses(ctx context.Context, orgID, requesterUserID int32, addresses []*ScanGroupAddress) (oid int32, failed []*FailedAddress, err error)
}

// ScanGroupReaderService read only implementation acquiring input lists and scan configs
type ScanGroupReaderService interface {
	Init(config []byte) error
	Get(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, group *ScanGroup, err error)
}
