package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/linkai-io/am/amtest"

	"github.com/linkai-io/am/am"

	"github.com/linkai-io/am/services/coordinator/state/redis"
)

var testQueueMap = map[string]string{"queue_name": "queue_url"}

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
	if err := r.Put(ctx, userContext, sg, testQueueMap); err != nil {
		t.Fatalf("error putting sg: %s\n", err)
	}

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
	exists, status, lastModified, err := r.GroupStatus(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error getting non-existent group status: %s\n", err)
	}

	if exists {
		t.Fatalf("group should not have existed\n")
	}

	if err := r.Put(ctx, userContext, sg, testQueueMap); err != nil {
		t.Fatalf("error putting sg: %s\n", err)
	}

	defer func() {
		if err := r.Delete(ctx, userContext, sg); err != nil {
			t.Fatalf("error deleting all keys: %s\n", err)
		}
	}()

	exists, status, lastModified, err = r.GroupStatus(ctx, userContext, gid)
	if err != nil {
		t.Fatalf("error getting status: %s\n", err)
	}

	if !exists {
		t.Fatalf("error status for gid should have existed\n")
	}

	if now != lastModified {
		t.Fatalf("error last modified expected %v got %v\n", now, lastModified)
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
		if err := r.Put(ctx, userContext, sg, testQueueMap); err != nil {
			b.Fatalf("error putting sg: %s\n", err)
		}

		if err := r.Delete(ctx, userContext, sg); err != nil {
			b.Fatalf("error deleting all keys: %s\n", err)
		}
	}

}
