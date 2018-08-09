package am

import (
	"context"
)

const (
	RNScanGroupGroups = "lrn:service:scangroup:feature:groups"
)

// ModuleConfiguration contains all the module configurations
type ModuleConfiguration struct {
	NSModule    *NSModuleConfig    `json:"ns_module"`
	BruteModule *BruteModuleConfig `json:"brute_module"`
	PortModule  *PortModuleConfig  `json:"port_module"`
	WebModule   *WebModuleConfig   `json:"web_module"`
}

// ScanGroup is a grouping configuration that has owner related information
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
}
