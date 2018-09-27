package cache

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/linkai-io/am/am"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ScanGroupSubscriber manage scan group cache and state
type ScanGroupSubscriber struct {
	// concurrent safe cache of scan groups updated via Subscribe callbacks
	groupCache  *ScanGroupCache
	st          CacheStater
	exitContext context.Context
}

// NewScanGroupSubscriber returns a scan group cache state subscriber
// for listening to scan group updates and handling the caching. This
// will be re-used in most (if not all) modules.
func NewScanGroupSubscriber(exitContext context.Context, state CacheStater) *ScanGroupSubscriber {
	s := &ScanGroupSubscriber{}
	s.exitContext = exitContext
	s.groupCache = NewScanGroupCache()
	s.st = state
	go s.st.Subscribe(s.exitContext, s.ChannelOnStart, s.ChannelOnMessage, am.RNScanGroupGroups)
	return s
}

// ChannelOnStart when we are subscribed to listen for group/other state updates
func (s *ScanGroupSubscriber) ChannelOnStart() error {
	return nil
}

// ChannelOnMessage when we receieve updates to scan groups/other state.
func (s *ScanGroupSubscriber) ChannelOnMessage(channel string, data []byte) error {
	switch channel {
	case am.RNScanGroupGroups:
		key := string(data)
		if err := s.updateGroupFromState(key); err != nil {
			log.Error().Err(err).Msg("updating group from state failed")
		}
	}
	return nil
}

// updateGroupFromState calls out to our state system to grab the scan group
// and put it in our group cache.
func (s *ScanGroupSubscriber) updateGroupFromState(key string) error {
	ctx := context.Background()
	orgID, groupID, err := s.splitKey(key)
	if err != nil {
		return err
	}

	wantModules := true
	group, err := s.st.GetGroup(ctx, orgID, groupID, wantModules)
	if err != nil {
		return errors.Wrap(err, "unable to get group from cache")
	}

	s.groupCache.Put(key, group)

	return nil
}

// GetGroupByIDs using the orgID/groupID first check if it is in our cache,
// if not grab the scan group from our state.
func (s *ScanGroupSubscriber) GetGroupByIDs(orgID, groupID int) (*am.ScanGroup, error) {
	var err error

	key := s.groupCache.MakeGroupKey(orgID, groupID)
	group := s.groupCache.Get(key)
	if group == nil {
		ctx := context.Background()
		wantModules := true
		group, err = s.st.GetGroup(ctx, orgID, groupID, wantModules)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get group from cache")
		}
		s.groupCache.Put(key, group)
	}
	return group, nil
}

// splitKey into org/group
func (s *ScanGroupSubscriber) splitKey(key string) (orgID int, groupID int, err error) {
	keys := strings.Split(key, ":")
	if len(keys) < 2 {
		return 0, 0, fmt.Errorf("failed to parse group key for cache, invalid key: %s", key)
	}
	orgID, err = strconv.Atoi(keys[0])
	if err != nil {
		return 0, 0, err
	}
	groupID, err = strconv.Atoi(keys[1])
	if err != nil {
		return 0, 0, err
	}
	return orgID, groupID, nil
}
