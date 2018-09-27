package cache

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/linkai-io/am/pkg/redisclient"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/am"
)

type ScanGroupCache struct {
	mu     sync.RWMutex
	groups map[string]*am.ScanGroup
}

func NewScanGroupCache() *ScanGroupCache {
	s := &ScanGroupCache{}
	s.groups = make(map[string]*am.ScanGroup, 0)
	return s
}

// Put adds or replaces a module configuration for a key
func (s *ScanGroupCache) Put(key string, group *am.ScanGroup) {

	keys := strings.Split(key, ":")
	if len(keys) < 2 {
		log.Warn().Str("key", key).Msg("failed to put group, invalid key")
		return
	}

	s.mu.Lock()
	s.groups[key] = group
	s.mu.Unlock()
}

// Get a module configuration via the key
func (s *ScanGroupCache) Get(key string) *am.ScanGroup {
	groupCopy := &am.ScanGroup{}

	s.mu.RLock()
	defer s.mu.RUnlock()

	group, ok := s.groups[key]
	if !ok {
		return nil
	}

	if err := Copy(group, groupCopy); err != nil {
		return nil
	}

	return groupCopy
}

func (s *ScanGroupCache) Clear(key string) {
	s.mu.Lock()
	delete(s.groups, key)
	s.mu.Unlock()
}

// GetByIDs returns the scangroup for the specified org/group
func (s *ScanGroupCache) GetByIDs(orgID, groupID int) *am.ScanGroup {
	key := s.MakeGroupKey(orgID, groupID)
	return s.Get(key)
}

// MakeGroupKey from org/group id
func (s *ScanGroupCache) MakeGroupKey(orgID, groupID int) string {
	return redisclient.NewRedisKeys(orgID, groupID).Config()
}

// Copy *yawn* yeah whatever
func Copy(src interface{}, dst interface{}) error {
	if dst == nil {
		return fmt.Errorf("dst cannot be nil")
	}
	if src == nil {
		return fmt.Errorf("src cannot be nil")
	}

	bytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("Unable to marshal src: %s", err)
	}
	err = json.Unmarshal(bytes, dst)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal into dst: %s", err)
	}
	return nil
}
