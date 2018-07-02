package am

import (
	"context"
)

// ModuleConfiguration contains all the module configurations
type ModuleConfiguration struct {
	NSModule    *NSModuleConfig
	BruteModule *BruteModuleConfig
	PortModule  *PortModuleConfig
	WebModule   *WebModuleConfig
}

// ScanGroup is an initial scan grouping configuration that has the original
// input file along with owner related information
type ScanGroup struct {
	OrgID         int32
	GroupID       int32
	GroupName     string
	CreationTime  int64
	CreatedBy     int32
	OriginalInput []byte
	Deleted       bool
}

// ScanGroupVersion tracks versions of scan group configurations to support
// adding and removing hosts, and changing module configurations
type ScanGroupVersion struct {
	OrgID                int32
	GroupID              int32
	GroupVersionID       int32
	VersionName          string
	CreationTime         int64
	CreatedBy            int32
	ModuleConfigurations *ModuleConfiguration
	Deleted              bool
}

// ScanGroupAddress contains details on addresses belonging to the scan group
// for scanning.
type ScanGroupAddress struct {
	AddressID int64
	OrgID     int32
	GroupID   int32
	Address   string
	Settings  *ModuleConfiguration
	AddedTime int64
	AddedBy   string
	Ignored   bool
}

// FailedAddress is used when we are unable to add or update addresses
type FailedAddress struct {
	Address      string
	FailedReason string
}

// Input represents a parsed and validated input
type Input map[string]interface{}

// ScanGroupService manages input lists and configurations for an organization and group. OrgIDs should
// always be returned for ensuring data integrity for requesters
type ScanGroupService interface {
	Init(config []byte) error
	Get(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, group *ScanGroup, err error)
	Create(ctx context.Context, orgID, requesterUserID int32, newGroup *ScanGroup) (oid int32, gid int32, err error)
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
