package am

import (
	"sync"
	"sync/atomic"
	"time"
)

// ScanGroupsStats stats of scan groups
type ScanGroupsStats struct {
	groups    map[int]*GroupStats
	groupLock sync.RWMutex
}

// NewScanGroupsStats for holding statistics of our active scan groups
func NewScanGroupsStats() *ScanGroupsStats {
	return &ScanGroupsStats{groups: make(map[int]*GroupStats)}
}

// AddGroup of addresses to have statistics collected for
func (s *ScanGroupsStats) AddGroup(userContext UserContext, orgID, groupID int) {
	s.groupLock.Lock()
	s.groups[groupID] = NewGroupStats(userContext, orgID, groupID)
	s.groupLock.Unlock()
}

// IncActive of how many addresses are being analyzed
func (s *ScanGroupsStats) IncActive(groupID int, count int32) {
	s.groupLock.RLock()
	group, ok := s.groups[groupID]
	if ok {
		group.IncActive(count)
	}
	s.groupLock.RUnlock()
}

// SetBatchSize of how many addresses will be analyzed for this group
func (s *ScanGroupsStats) SetBatchSize(groupID int, count int32) {
	s.groupLock.RLock()
	group, ok := s.groups[groupID]
	if ok {
		group.SetBatchSize(count)
	}
	s.groupLock.RUnlock()
}

func (s *ScanGroupsStats) SetComplete(groupID int) {
	s.groupLock.RLock()
	group, ok := s.groups[groupID]
	if ok {
		group.SetEndTime()
	}
	s.groupLock.RUnlock()
}

// GetActive addresses being analyzed for this group
func (s *ScanGroupsStats) GetActive(groupID int) int32 {
	var active int32

	s.groupLock.RLock()
	defer s.groupLock.RUnlock()
	group, ok := s.groups[groupID]
	if ok {
		active = group.GetActive()
	}
	return active
}

// GetGroup returns a copy of the group
func (s *ScanGroupsStats) GetGroup(groupID int) *GroupStats {
	s.groupLock.RLock()
	defer s.groupLock.RUnlock()

	if stats, ok := s.groups[groupID]; ok {
		return &GroupStats{
			UserContext:     stats.UserContext, // this is probably not safe *shrug* TODO: deep copy
			OrgID:           stats.OrgID,
			GroupID:         stats.GroupID,
			ActiveAddresses: stats.ActiveAddresses,
			BatchSize:       stats.BatchSize,
			BatchStart:      stats.BatchStart,
			BatchEnd:        stats.BatchEnd,
		}
	}
	return nil
}

// Groups returns a list of all groups
func (s *ScanGroupsStats) Groups() []*GroupStats {
	s.groupLock.RLock()
	defer s.groupLock.RUnlock()
	groups := make([]*GroupStats, 0)
	for _, group := range s.groups {
		groups = append(groups, s.GetGroup(group.GroupID))
	}
	return groups
}

// DeleteGroup from the stats container
func (s *ScanGroupsStats) DeleteGroup(groupID int) {
	s.groupLock.Lock()
	if _, ok := s.groups[groupID]; ok {
		delete(s.groups, groupID)
	}
	s.groupLock.Unlock()
}

// GroupStats holds basic information on active groups running
type GroupStats struct {
	UserContext     UserContext `json:"-"`
	OrgID           int         `json:"org_id"`
	GroupID         int         `json:"group_id"`
	ActiveAddresses int32       `json:"active_addresses"`
	BatchSize       int32       `json:"batch_size"`
	LastUpdated     int64       `json:"last_updated"` // only comes back from DB
	BatchStart      int64       `json:"batch_start"`
	BatchEnd        int64       `json:"batch_end"`
}

// NewGroupStats initializes with org/group ids
func NewGroupStats(userContext UserContext, orgID, groupID int) *GroupStats {
	return &GroupStats{UserContext: userContext, OrgID: orgID, GroupID: groupID, BatchStart: time.Now().UnixNano()}
}

// IncActive addresses by count ( can be negative to decrease)
func (g *GroupStats) IncActive(count int32) {
	atomic.AddInt32(&g.ActiveAddresses, count)
}

// SetBatchSize of how many addresses we analyzed this batch
func (g *GroupStats) SetBatchSize(count int32) {
	atomic.AddInt32(&g.BatchSize, count)
}

// GetActive count of addresses
func (g *GroupStats) GetActive() int32 {
	return atomic.LoadInt32(&g.ActiveAddresses)
}

// SetEndTime for this batch
func (g *GroupStats) SetEndTime() {
	atomic.StoreInt64(&g.BatchEnd, time.Now().UnixNano())
}
