package redis_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/pkg/state/redis"

	"github.com/linkai-io/am/am"
)

func TestPut(t *testing.T) {
	r := redis.New()
	if err := r.Init([]byte("{\"rc_addr\":\"0.0.0.0:6379\",\"rc_pass\":\"test132\"}")); err != nil {
		t.Fatalf("error connecting to redis: %s\n", err)
	}
	now := time.Now().UnixNano()
	sg := &am.ScanGroup{
		OrgID:                1,
		GroupID:              1,
		GroupName:            "testredis",
		CreationTime:         now,
		ModifiedTime:         now,
		ModuleConfigurations: amtest.CreateModuleConfig(),
	}
	userContext := amtest.CreateUserContext(1, 1)
	ctx := context.Background()
	if err := r.Put(ctx, userContext, sg); err != nil {
		t.Fatalf("error putting sg: %s\n", err)
	}

	if err := r.Delete(ctx, userContext, sg); err != nil {
		t.Fatalf("error deleting all keys: %s\n", err)
	}
}

func TestAddresses(t *testing.T) {
	r := redis.New()
	if err := r.Init([]byte("{\"rc_addr\":\"0.0.0.0:6379\",\"rc_pass\":\"test132\"}")); err != nil {
		t.Fatalf("error connecting to redis: %s\n", err)
	}
	now := time.Now().UnixNano()
	sg := &am.ScanGroup{
		OrgID:                1,
		GroupID:              1,
		GroupName:            "testredis",
		CreationTime:         now,
		ModifiedTime:         now,
		ModuleConfigurations: amtest.CreateModuleConfig(),
	}
	userContext := amtest.CreateUserContext(1, 1)
	ctx := context.Background()
	if err := r.Put(ctx, userContext, sg); err != nil {
		t.Fatalf("error putting sg: %s\n", err)
	}

	defer r.Delete(ctx, userContext, sg)

	addrs := amtest.GenerateAddrs(userContext.GetOrgID(), sg.GroupID, 100)
	if err := r.PutAddresses(ctx, userContext, sg.GroupID, addrs); err != nil {
		t.Fatalf("error pushing addresses: %s\n", err)
	}

	returned, err := r.GetAddresses(ctx, userContext, sg.GroupID, 100)
	if err != nil {
		t.Fatalf("error getting addresses: %s\n", err)
	}

	expected := make(map[int64]*am.ScanGroupAddress, 0)
	for i := 0; i < len(addrs); i++ {
		expected[addrs[i].AddressID] = addrs[i]
	}
	amtest.TestCompareAddresses(expected, returned, t)

	exists, err := r.Exists(ctx, userContext.GetOrgID(), sg.GroupID, addrs[0].HostAddress, addrs[0].IPAddress)
	if err != nil {
		t.Fatalf("error testing if member host: %s ip: %s exists: %s\n", addrs[0].HostAddress, addrs[0].IPAddress, err)
	}

	if !exists {
		t.Fatalf("error member host: %s ip: %s did not exist", addrs[0].HostAddress, addrs[0].IPAddress)
	}

	host := "notexist"
	ip := "notexist"
	exists, err = r.Exists(ctx, userContext.GetOrgID(), sg.GroupID, host, ip)
	if err != nil {
		t.Fatalf("error testing if member host: %s ip: %s exists: %s\n", host, ip, err)
	}

	if exists {
		t.Fatalf("error member host: %s ip: %s should not exist", host, ip)
	}
}

func TestGetGroup(t *testing.T) {
	r := redis.New()
	if err := r.Init([]byte("{\"rc_addr\":\"0.0.0.0:6379\",\"rc_pass\":\"test132\"}")); err != nil {
		t.Fatalf("error connecting to redis: %s\n", err)
	}
	now := time.Now().UnixNano()
	sg := &am.ScanGroup{
		OrgID:                1,
		GroupID:              1,
		GroupName:            "testredis",
		CreationTime:         now,
		ModifiedTime:         now,
		ModuleConfigurations: amtest.CreateModuleConfig(),
	}
	userContext := amtest.CreateUserContext(1, 1)
	ctx := context.Background()
	if err := r.Put(ctx, userContext, sg); err != nil {
		t.Fatalf("error putting sg: %s\n", err)
	}

	wantModules := true
	returned, err := r.GetGroup(ctx, 1, 1, wantModules)
	if err != nil {
		t.Fatalf("error getting group: %s\n", err)
	}

	if returned.ModuleConfigurations == nil {
		t.Fatalf("error module configurations was nil\n")
	}

	amtest.TestCompareScanGroup(sg, returned, t)
	amtest.TestCompareGroupModules(sg.ModuleConfigurations, returned.ModuleConfigurations, t)

	if err := r.Delete(ctx, userContext, sg); err != nil {
		t.Fatalf("error deleting all keys: %s\n", err)
	}
}

func TestGroupStatus(t *testing.T) {
	oid := 2
	gid := 2
	r := redis.New()
	if err := r.Init([]byte("{\"rc_addr\":\"0.0.0.0:6379\",\"rc_pass\":\"test132\"}")); err != nil {
		t.Fatalf("error connecting to redis: %s\n", err)
	}
	now := time.Now().UnixNano()
	sg := &am.ScanGroup{
		OrgID:                oid,
		GroupID:              gid,
		GroupName:            "testredis",
		CreationTime:         now,
		ModifiedTime:         now,
		ModuleConfigurations: amtest.CreateModuleConfig(),
	}
	userContext := amtest.CreateUserContext(oid, gid)
	ctx := context.Background()

	// test empty
	exists, status, err := r.GroupStatus(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error getting non-existent group status: %s\n", err)
	}

	if exists {
		t.Fatalf("group should not have existed\n")
	}

	if err := r.Put(ctx, userContext, sg); err != nil {
		t.Fatalf("error putting sg: %s\n", err)
	}

	defer func() {
		if err := r.Delete(ctx, userContext, sg); err != nil {
			t.Fatalf("error deleting all keys: %s\n", err)
		}
	}()

	exists, status, err = r.GroupStatus(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error getting status: %s\n", err)
	}

	if !exists {
		t.Fatalf("error status for gid should have existed\n")
	}

	if status != am.GroupStopped {
		t.Fatalf("expected group stopped got: %v\n", am.GroupStatusMap[status])
	}
}

func BenchmarkPut(b *testing.B) {
	r := redis.New()
	if err := r.Init([]byte("{\"rc_addr\":\"0.0.0.0:6379\",\"rc_pass\":\"test132\"}")); err != nil {
		b.Fatalf("error connecting to redis: %s\n", err)
	}
	now := time.Now().UnixNano()
	sg := &am.ScanGroup{
		OrgID:                1,
		GroupID:              1,
		GroupName:            "testredis",
		CreationTime:         now,
		ModifiedTime:         now,
		ModuleConfigurations: amtest.CreateModuleConfig(),
	}
	userContext := amtest.CreateUserContext(1, 1)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := r.Put(ctx, userContext, sg); err != nil {
			b.Fatalf("error putting sg: %s\n", err)
		}

		if err := r.Delete(ctx, userContext, sg); err != nil {
			b.Fatalf("error deleting all keys: %s\n", err)
		}
	}

}

func TestState_DoNSRecords(t *testing.T) {
	r := redis.New()
	if err := r.Init([]byte("{\"rc_addr\":\"0.0.0.0:6379\",\"rc_pass\":\"test132\"}")); err != nil {
		t.Fatalf("error connecting to redis: %s\n", err)
	}
	ctx := context.Background()
	orgID := 1
	groupID := 1
	testSeconds := 1
	ok, err := r.DoNSRecords(ctx, orgID, groupID, testSeconds, "test.org")
	if err != nil {
		t.Fatalf("got error setting ns records: %s\n", err)
	}

	if !ok {
		t.Fatalf("error should have been OK to test records for new zone\n")
	}

	ok, err = r.DoNSRecords(ctx, orgID, groupID, testSeconds, "test.org")
	if err != nil {
		t.Fatalf("got error setting ns records: %s\n", err)
	}

	if ok {
		t.Fatalf("error should have NOT been ok to test records for new zone\n")
	}

	time.Sleep(time.Second * 2)
	ok, err = r.DoNSRecords(ctx, orgID, groupID, testSeconds, "test.org")
	if err != nil {
		t.Fatalf("got error setting ns records: %s\n", err)
	}

	if !ok {
		t.Fatalf("error should have been OK to test records for new zone after expiration\n")
	}
}

func TestState_Subscribe(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(2)
	onStartFn := func() error {
		t.Logf("started!\n")
		wg.Done()
		return nil
	}

	onMsgFn := func(channel string, data []byte) error {
		t.Logf("On channel %s got %v\n", channel, string(data))
		wg.Done()
		return nil
	}

	r := redis.New()
	if err := r.Init([]byte("{\"rc_addr\":\"0.0.0.0:6379\",\"rc_pass\":\"test132\"}")); err != nil {
		t.Fatalf("error connecting to redis: %s\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go r.Subscribe(ctx, onStartFn, onMsgFn, "test")

	conn := r.TestGetConn()
	time.Sleep(time.Second)
	conn.Do("PUBLISH", "test", "why hello there")
	time.Sleep(time.Second * 3)
	cancel()

	wg.Wait()
}
