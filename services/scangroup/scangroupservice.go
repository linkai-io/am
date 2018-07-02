package scangroup

import (
	"context"

	"gopkg.linkai.io/v1/repos/am/am"

	"gopkg.linkai.io/v1/repos/am/services/scangroup/store"
)

// Service implements the logic of the ScanGroupService. It manages adding new scan groups
// and versions for different configuration possibilities.
type Service struct {
	store store.Storer
}

// New returns a new ScanGroupService backed by a datastore
func New(store store.Storer) *Service {
	return &Service{store: store}
}

// Init the service
func (s *Service) Init(config []byte) error {
	return nil
}

// Get returns a scan group identified by scangroup id
func (s *Service) Get(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, group *am.ScanGroup, err error) {
	return oid, group, err
}

// Create a new scan group, returning orgID and groupID on success, error otherwise
func (s *Service) Create(ctx context.Context, orgID, requesterUserID int32, newGroup *am.ScanGroup) (oid int32, gid int32, err error) {
	return oid, gid, err
}

// Delete a scan group, returning orgID and groupID on success, error otherwise
func (s *Service) Delete(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, gid int32, err error) {
	return oid, gid, err
}

// GetVersion returns the configuration of the requested version.
func (s *Service) GetVersion(ctx context.Context, orgID, requesterUserID, groupID, groupVersionID int32) (oid int32, groupVersion *am.ScanGroupVersion, err error) {
	return oid, groupVersion, err
}

// CreateVersion for a scan group, allowing modification of module configurations
func (s *Service) CreateVersion(ctx context.Context, orgID, requesterUserID int32, scanGroupVersion *am.ScanGroupVersion) (oid int32, gid int32, gvid int32, err error) {
	return oid, gid, gvid, err
}

// DeleteVersion requires orgID, groupVersionID and one of groupID or versionName. returning orgID, groupID and groupVersionID if success
func (s *Service) DeleteVersion(ctx context.Context, orgID, requesterUserID, groupID, groupVersionID int32, versionName string) (oid int32, gid int32, gvid int32, err error) {
	return oid, gid, gvid, err
}

// Groups returns all groups for an organization.
func (s *Service) Groups(ctx context.Context, orgID int32) (oid int32, groups []*am.ScanGroup, err error) {
	return oid, groups, err
}

// Addresses returns all addresses for a scan group
func (s *Service) Addresses(ctx context.Context, orgID, requesterUserID, groupID int32) (oid int32, addresses []*am.ScanGroupAddress, err error) {
	return oid, addresses, err
}

// AddAddresses adds new addresses to the scan_group address table
func (s *Service) AddAddresses(ctx context.Context, orgID, requesterUserID int32, addresses []*am.ScanGroupAddress) (oid int32, failed []*am.FailedAddress, err error) {
	return oid, failed, err
}

// UpdateAddresses updates addresses with new configuration settings to the scan_group address table
func (s *Service) UpdateAddresses(ctx context.Context, orgID, requesterUserID int32, addresses []*am.ScanGroupAddress) (oid int32, failed []*am.FailedAddress, err error) {
	return oid, failed, err
}
