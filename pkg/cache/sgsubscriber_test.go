package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/linkai-io/am/amtest"

	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/state"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/redisclient"
)

func TestScanGroupSubscriber(t *testing.T) {
	orgID := 1
	groupID := 1
	cacher := &mock.CacheStater{}
	expected := &am.ScanGroup{
		OrgID:                orgID,
		GroupID:              groupID,
		ModifiedBy:           "user@email.com",
		ModifiedByID:         groupID,
		CreatedBy:            "user@email.com",
		CreatedByID:          groupID,
		ModifiedTime:         time.Now().UnixNano(),
		CreationTime:         time.Now().UnixNano(),
		GroupName:            "test",
		OriginalInputS3URL:   "s3://url",
		ModuleConfigurations: amtest.CreateModuleConfig(),
	}

	expectedUpdate := &am.ScanGroup{
		OrgID:                orgID,
		GroupID:              groupID,
		ModifiedBy:           "user@email.com",
		ModifiedByID:         groupID,
		CreatedBy:            "user@email.com",
		CreatedByID:          groupID,
		ModifiedTime:         time.Now().UnixNano(),
		CreationTime:         time.Now().UnixNano(),
		GroupName:            "updated",
		OriginalInputS3URL:   "s3://url",
		ModuleConfigurations: amtest.CreateModuleConfig(),
	}

	cacher.GetGroupFn = func(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
		return expected, nil
	}

	cacher.SubscribeFn = func(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
		return nil
	}

	ctx := context.Background()
	sub := cache.NewScanGroupSubscriber(ctx, cacher)

	returned, err := sub.GetGroupByIDs(orgID, groupID)
	if err != nil {
		t.Fatalf("error getting group: %s\n", err)
	}

	if cacher.GetGroupInvoked == false {
		t.Fatalf("did not invoke cache state GetGroup")
	}

	amtest.TestCompareScanGroup(expected, returned, t)
	amtest.TestCompareGroupModules(expected.ModuleConfigurations, returned.ModuleConfigurations, t)

	// test update
	cacher.GetGroupFn = func(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
		return expectedUpdate, nil
	}
	key := redisclient.NewRedisKeys(orgID, groupID).Config()
	sub.ChannelOnMessage(am.RNScanGroupGroups, []byte(key))

	returned, err = sub.GetGroupByIDs(orgID, groupID)
	if err != nil {
		t.Fatalf("error getting group after update: %s\n", err)
	}

	amtest.TestCompareScanGroup(expectedUpdate, returned, t)
	amtest.TestCompareGroupModules(expectedUpdate.ModuleConfigurations, returned.ModuleConfigurations, t)

}
