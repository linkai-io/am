package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/linkai-io/am/modules/ns/state/redis"
)

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
