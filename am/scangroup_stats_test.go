package am_test

import (
	"testing"

	"github.com/linkai-io/am/am"
)

func TestGroupStats(t *testing.T) {
	g := am.NewScanGroupsStats()
	userContext := &am.UserContextData{
		OrgID:  1,
		UserID: 1,
	}
	g.AddGroup(userContext, userContext.OrgID, 1)

	stats := g.GetGroup(1)
	if stats.BatchStart == 0 {
		t.Fatalf("error batch time not set on creation")
	}
	if stats.GroupID == 0 {
		t.Fatalf("error groupid not set on creation")
	}

	g.IncActive(1, 10)
	if g.GetActive(1) != 10 {
		t.Fatalf("active should be 10")
	}

	g.IncActive(1, -10)
	if g.GetActive(1) != 0 {
		t.Fatalf("active should be 0")
	}

	g.SetBatchSize(1, 10)
	stats = g.GetGroup(1)
	if stats.BatchSize != 10 {
		t.Fatalf("batch size should be 10 got: %d\n", stats.BatchSize)
	}

	g.AddGroup(userContext, userContext.OrgID, 2)
	if len(g.Groups()) != 2 {
		t.Fatalf("should have two groups, got %d\n", len(g.Groups()))
	}

	g.DeleteGroup(1)
	if len(g.Groups()) != 1 {
		t.Fatalf("should have one groups, got %d\n", len(g.Groups()))
	}

	// make sure we don't panic on invalid group
	g.DeleteGroup(100)

}
