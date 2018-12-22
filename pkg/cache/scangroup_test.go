package cache_test

import (
	"testing"
	"time"

	"github.com/linkai-io/am/amtest"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/redisclient"
)

func TestScanGroupCacheEmpty(t *testing.T) {

	sc := cache.NewScanGroupCache()
	if s := sc.Get("test"); s != nil {
		t.Fatalf("error empty key should return nil")
	}

	if s := sc.GetByIDs(0, 0); s != nil {
		t.Fatalf("error empty uid/gid should return nil")
	}
}

func TestScanGroupCachePut(t *testing.T) {
	orgID := 1
	groupID := 1
	sc := cache.NewScanGroupCache()
	sg := &am.ScanGroup{
		OrgID:                orgID,
		GroupID:              groupID,
		GroupName:            "test",
		CreationTime:         time.Now().UnixNano(),
		ModifiedBy:           "user@email.com",
		ModifiedByID:         1,
		CreatedBy:            "user@email.com",
		CreatedByID:          1,
		ModifiedTime:         time.Now().UnixNano(),
		OriginalInputS3URL:   "s3://test",
		Paused:               true,
		Deleted:              true,
		ModuleConfigurations: amtest.CreateModuleConfig(),
	}
	key := redisclient.NewRedisKeys(orgID, groupID).Config()
	sc.Put(key, sg)

	returned := sc.Get(key)
	amtest.TestCompareScanGroup(sg, returned, t)
	amtest.TestCompareGroupModules(sg.ModuleConfigurations, returned.ModuleConfigurations, t)

	returned = sc.GetByIDs(orgID, groupID)
	amtest.TestCompareScanGroup(sg, returned, t)
	amtest.TestCompareGroupModules(sg.ModuleConfigurations, returned.ModuleConfigurations, t)

	sc.Clear(key)

	if returned := sc.Get(key); returned != nil {
		t.Fatalf("error retrieved data that should be cleared\n")
	}
}
