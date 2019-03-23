package event_test

import (
	"context"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/services/event"
)

var env string
var dbstring string

const serviceKey = "eventservice"

func init() {
	var err error
	flag.StringVar(&env, "env", "local", "environment we are running tests in")
	flag.Parse()
	sec := secrets.NewSecretsCache(env, "")
	dbstring, err = sec.DBString(serviceKey)
	if err != nil {
		panic("error getting dbstring secret")
	}
}

func TestNew(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	service := event.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}
}

func TestAddGet(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "eventaddget"
	groupName := "eventaddgetgroup"

	auth := amtest.MockEmptyAuthorizer()

	service := event.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing event service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
	defer db.Close()

	amtest.CreateOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	defer amtest.DeleteOrg(db, orgName, t)
	//userID := amtest.GetUserId(db, orgID, orgName, t)

	groupID := amtest.CreateScanGroup(db, orgName, groupName, t)
	userContext := amtest.CreateUserContext(orgID, 1)

	events := make([]*am.Event, 4)
	now := time.Now()
	events[0] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         1,
		EventTimestamp: now.UnixNano(),
		Data: map[string][]string{
			am.EventTypes[1]: []string{"completed run"},
		},
	}
	events[1] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         10,
		EventTimestamp: now.UnixNano(),
		Data: map[string][]string{
			am.EventTypes[10]: []string{"example.com", "test.example.com", "something.example.com"},
		},
	}
	events[2] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         100,
		EventTimestamp: now.UnixNano(),
		Data: map[string][]string{
			am.EventTypes[100]: []string{"https://blah.example.com"},
		},
	}
	events[3] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         150,
		EventTimestamp: now.UnixNano(),
		Data: map[string][]string{
			am.EventTypes[150]: []string{"test.example.com", "443", "1111111111111"},
		},
	}
	if err := service.Add(ctx, userContext, events); err != nil {
		t.Fatalf("error adding events: %v\n", err)
	}

	settings := &am.UserEventSettings{
		WeeklyReportSendDay: 0,
		ShouldWeeklyEmail:   false,
		DailyReportSendHour: 0,
		ShouldDailyEmail:    false,
		UserTimezone:        "Asia/Tokyo",
		Subscriptions: []*am.EventSubscriptions{
			&am.EventSubscriptions{
				TypeID:              1,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              10,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              100,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
		},
	}

	if err := service.UpdateSettings(ctx, userContext, settings); err != nil {
		t.Fatalf("error updating user settings: %v\n", err)
	}

	returned, err := service.Get(ctx, userContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: &am.FilterType{}})
	if err != nil {
		t.Fatalf("error getting events: %v\n", err)
	}

	// should only get 3 since those are what we subscribed to
	if len(returned) != 3 {
		t.Fatalf("expected 3 results got %v\n", len(returned))
	}
	for _, e := range returned {
		t.Logf("%#v\n", e)
	}
}
