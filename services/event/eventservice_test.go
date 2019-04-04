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
	"github.com/linkai-io/am/services/webdata"
)

var env string
var dbstring string
var webDBString string

const serviceKey = "eventservice"
const webServiceKey = "webdataservice"

func init() {
	var err error
	flag.StringVar(&env, "env", "local", "environment we are running tests in")
	flag.Parse()
	sec := secrets.NewSecretsCache(env, "")
	dbstring, err = sec.DBString(serviceKey)
	if err != nil {
		panic("error getting dbstring secret")
	}

	webDBString, err = sec.DBString(webServiceKey)
	if err != nil {
		panic("error getting dbstring secret for webdata")
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
		Data:           []string{"completed run"},
	}
	events[1] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         am.EventNewHost,
		EventTimestamp: now.UnixNano(),
		Data:           []string{"example.com", "test.example.com", "something.example.com"},
	}
	events[2] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         am.EventNewWebsite,
		EventTimestamp: now.UnixNano(),
		Data:           []string{"https://blah.example.com"},
	}
	events[3] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         am.EventCertExpiring,
		EventTimestamp: now.UnixNano(),
		Data:           []string{"test.example.com", "443", "1111111111111"},
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
				TypeID:              am.EventInitialGroupComplete,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventNewHost,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventNewWebsite,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
		},
	}

	if err := service.UpdateSettings(ctx, userContext, settings); err != nil {
		t.Fatalf("error updating user settings: %v\n", err)
	}

	retSettings, err := service.GetSettings(ctx, userContext)
	if err != nil {
		t.Fatalf("error getting settings: %v\n", err)
	}
	testCompareSettings(settings, retSettings, t)

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

func TestNotifyComplete(t *testing.T) {
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

	webService := webdata.New(auth)
	if err := webService.Init([]byte(webDBString)); err != nil {
		t.Fatalf("error initalizing webdata service: %s\n", err)
	}

	amtest.CreateOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	defer amtest.DeleteOrg(db, orgName, t)
	userID := amtest.GetUserId(db, orgID, orgName, t)

	groupID := amtest.CreateScanGroup(db, orgName, groupName, t)
	userContext := amtest.CreateUserContext(orgID, userID)
	now := time.Now()

	// populate with fake data
	addr := amtest.CreateScanGroupAddress(db, orgID, groupID, t)
	webData := amtest.CreateMultiWebData(addr, addr.HostAddress, addr.IPAddress)
	for _, web := range webData {
		if _, err := webService.Add(ctx, userContext, web); err != nil {
			t.Fatalf("error adding webdata for notify complete")
		}
	}

	settings := &am.UserEventSettings{
		WeeklyReportSendDay: 0,
		ShouldWeeklyEmail:   true,
		DailyReportSendHour: 0,
		ShouldDailyEmail:    true,
		UserTimezone:        "Asia/Tokyo",
		Subscriptions: []*am.EventSubscriptions{
			&am.EventSubscriptions{
				TypeID:              am.EventInitialGroupComplete,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventNewHost,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventNewWebsite,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventCertExpiring,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
		},
	}

	if err := service.UpdateSettings(ctx, userContext, settings); err != nil {
		t.Fatalf("error updating user settings: %v\n", err)
	}

	err := service.NotifyComplete(ctx, userContext, now.UnixNano(), groupID)
	if err != nil {
		t.Fatalf("error notifying complete")
	}

	returned, err := service.Get(ctx, userContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: &am.FilterType{}})
	if err != nil {
		t.Fatalf("error getting events: %v\n", err)
	}

	// should only get 1 since those are what we subscribed to
	if len(returned) != 3 {
		t.Fatalf("expected 1 results got %v\n", len(returned))
	}
	for _, e := range returned {
		t.Logf("%#v\n", e)
	}

	notificationIDs := make([]int64, len(returned)-1)
	for i, ret := range returned {
		if i == len(returned)-1 { // leave 1 notification left so we can make sure it works properly
			break
		}
		notificationIDs[i] = ret.NotificationID
	}

	if err := service.MarkRead(ctx, userContext, notificationIDs); err != nil {
		t.Fatalf("error marking notifications as read: %v\n", err)
	}

	returned, err = service.Get(ctx, userContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: &am.FilterType{}})
	if err != nil {
		t.Fatalf("error getting events after mark read: %v\n", err)
	}

	if len(returned) != 1 {
		t.Fatalf("error events returned after mark read should be 1 got: %d\n", len(returned))
	}

}

func testCompareSettings(e, r *am.UserEventSettings, t *testing.T) {
	if e.DailyReportSendHour != r.DailyReportSendHour {
		t.Fatalf("DailyReportSendHour did not match expected: %v %v\n", e.DailyReportSendHour, r.DailyReportSendHour)
	}

	if e.ShouldDailyEmail != r.ShouldDailyEmail {
		t.Fatalf("ShouldDailyEmail did not match expected: %v %v\n", e.ShouldDailyEmail, r.ShouldDailyEmail)
	}

	if e.ShouldWeeklyEmail != r.ShouldWeeklyEmail {
		t.Fatalf("ShouldWeeklyEmail did not match expected: %v %v\n", e.ShouldWeeklyEmail, r.ShouldWeeklyEmail)
	}

	if e.UserTimezone != r.UserTimezone {
		t.Fatalf("UserTimezone did not match expected: %v %v\n", e.UserTimezone, r.UserTimezone)
	}

	if e.WeeklyReportSendDay != r.WeeklyReportSendDay {
		t.Fatalf("WeeklyReportSendDay did not match expected: %v %v\n", e.WeeklyReportSendDay, r.WeeklyReportSendDay)
	}

	if len(e.Subscriptions) != len(r.Subscriptions) {
		t.Fatalf("Subscriptions %d did not match len %d\n", len(e.Subscriptions), len(r.Subscriptions))
	}
}