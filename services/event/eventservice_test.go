package event_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/services/address"
	"github.com/linkai-io/am/services/event"
	"github.com/linkai-io/am/services/webdata"
)

var env string
var dbstring string
var webDBString string
var addrDBString string

const serviceKey = "eventservice"
const webServiceKey = "webdataservice"
const addressKey = "addressservice"

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

	addrDBString, err = sec.DBString(addressKey)
	if err != nil {
		panic("error getting dbstring secret for address")
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

	orgOne := "eventaddget1"
	groupOne := "eventaddgetgroup1"
	orgTwo := "eventaddget2"
	groupTwo := "eventaddgetgroup2"
	groupThree := "eventaddgetgroup3"

	auth := amtest.MockEmptyAuthorizer()

	service := event.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing event service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
	defer db.Close()

	amtest.CreateOrg(db, orgOne, t)
	orgOneID := amtest.GetOrgID(db, orgOne, t)
	defer amtest.DeleteOrg(db, orgOne, t)

	amtest.CreateOrg(db, orgTwo, t)
	orgTwoID := amtest.GetOrgID(db, orgTwo, t)
	defer amtest.DeleteOrg(db, orgTwo, t)

	userOneID := amtest.GetUserId(db, orgOneID, orgOne, t)
	userTwoID := amtest.GetUserId(db, orgTwoID, orgTwo, t)

	groupOneID := amtest.CreateScanGroup(db, orgOne, groupOne, t)
	groupTwoID := amtest.CreateScanGroup(db, orgTwo, groupTwo, t)
	groupThreeID := amtest.CreateScanGroup(db, orgTwo, groupThree, t)

	userOneContext := amtest.CreateUserContext(orgOneID, userOneID)
	userTwoContext := amtest.CreateUserContext(orgTwoID, userTwoID)

	now := time.Now()

	eventsOne := makeEvents(orgOneID, groupOneID, now)
	if err := service.Add(ctx, userOneContext, eventsOne); err != nil {
		t.Fatalf("error adding events 1: %v\n", err)
	}

	eventsTwo := makeEvents(orgTwoID, groupTwoID, now)
	if err := service.Add(ctx, userTwoContext, eventsTwo); err != nil {
		t.Fatalf("error adding events 2: %v\n", err)
	}

	settings := &am.UserEventSettings{
		WeeklyReportSendDay: 0,
		ShouldWeeklyEmail:   false,
		DailyReportSendHour: 0,
		ShouldDailyEmail:    false,
		UserTimezone:        "Asia/Tokyo",
		Subscriptions: []*am.EventSubscriptions{
			&am.EventSubscriptions{
				TypeID:              am.EventInitialGroupCompleteID,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventNewHostID,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventNewWebsiteID,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
		},
	}

	if err := service.UpdateSettings(ctx, userOneContext, settings); err != nil {
		t.Fatalf("error updating user settings: %v\n", err)
	}

	retSettings, err := service.GetSettings(ctx, userOneContext)
	if err != nil {
		t.Fatalf("error getting settings: %v\n", err)
	}
	testCompareSettings(settings, retSettings, t)

	returned, err := service.Get(ctx, userOneContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: &am.FilterType{}})
	if err != nil {
		t.Fatalf("error getting events: %#v\n", err)
	}

	// should only get 3 since those are what we subscribed to
	if len(returned) != 3 {
		t.Fatalf("expected 3 results got %v\n", len(returned))
	}
	for _, e := range returned {
		t.Logf("%#v\n", e)
	}

	// Test group two
	if err := service.UpdateSettings(ctx, userTwoContext, settings); err != nil {
		t.Fatalf("error updating user settings: %v\n", err)
	}

	retSettings, err = service.GetSettings(ctx, userTwoContext)
	if err != nil {
		t.Fatalf("error getting settings: %v\n", err)
	}
	testCompareSettings(settings, retSettings, t)

	returned, err = service.Get(ctx, userTwoContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: &am.FilterType{}})
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

	eventsThree := makeEvents(orgTwoID, groupThreeID, now)
	if err := service.Add(ctx, userTwoContext, eventsThree); err != nil {
		t.Fatalf("error adding events new 2: %v\n", err)
	}
	returned, err = service.Get(ctx, userTwoContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: &am.FilterType{}})
	if err != nil {
		t.Fatalf("error getting events: %v\n", err)
	}
	// should get 6 since we didn't filter on group
	if len(returned) != 6 {
		t.Fatalf("expected 6 results got %v\n", len(returned))
	}
	for _, e := range returned {
		t.Logf("%#v\n", e)
	}
	groupFilter := &am.FilterType{}
	groupFilter.AddInt32(am.FilterEventGroupID, int32(groupThreeID))
	returned, err = service.Get(ctx, userTwoContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: groupFilter})
	if err != nil {
		t.Fatalf("error getting events: %v\n", err)
	}
	// should get 3 since we filtered on group
	if len(returned) != 3 {
		t.Fatalf("expected 3 results got %v\n", len(returned))
	}
	for _, e := range returned {
		t.Logf("%#v\n", e)
	}
}

func makeEvents(orgID, groupID int, now time.Time) []*am.Event {
	events := make([]*am.Event, 4)

	m, _ := json.Marshal(&am.EventInitialGroupComplete{Message: "completed run"})
	events[0] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         am.EventInitialGroupCompleteID,
		EventTimestamp: now.UnixNano(),
		JSONData:       string(m),
	}

	m, _ = json.Marshal([]*am.EventNewHost{
		&am.EventNewHost{Host: "example.com"},
		&am.EventNewHost{Host: "test.example.com"},
		&am.EventNewHost{Host: "something.example.com"},
	})
	events[1] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         am.EventNewHostID,
		EventTimestamp: now.UnixNano(),
		JSONData:       string(m),
	}

	m, _ = json.Marshal([]*am.EventNewWebsite{
		&am.EventNewWebsite{LoadURL: "https://blah.example.com", URL: "https://blah.example.com/", Port: 443},
	})
	events[2] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         am.EventNewWebsiteID,
		EventTimestamp: now.UnixNano(),
		JSONData:       string(m),
	}

	m, _ = json.Marshal([]*am.EventCertExpiring{
		&am.EventCertExpiring{SubjectName: "test.example.com", Port: 443, ValidTo: int64(1111111111111)},
	})
	events[3] = &am.Event{
		OrgID:          orgID,
		GroupID:        groupID,
		TypeID:         am.EventCertExpiringID,
		EventTimestamp: now.UnixNano(),
		JSONData:       string(m),
	}
	return events
}

func TestNotifyComplete(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "eventnotifycomplete"
	groupName := "eventnotifycomplete"

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

	// only this host should be found, since multiwebdata will create hosts that will exist < end of start scan time.
	newWebHost := amtest.CreateWebData(addr, "new.website.com", "1.1.1.1")
	newWebHost.DetectedTech = map[string]*am.WebTech{"AngularJS": &am.WebTech{
		Matched:  "1.5.3",
		Version:  "1.5.3",
		Location: "script",
	}}
	if _, err := webService.Add(ctx, userContext, newWebHost); err != nil {
		t.Fatalf("error adding single new host webdata for notify complete")
	}

	settings := &am.UserEventSettings{
		WeeklyReportSendDay: 0,
		ShouldWeeklyEmail:   true,
		DailyReportSendHour: 0,
		ShouldDailyEmail:    true,
		UserTimezone:        "Asia/Tokyo",
		Subscriptions: []*am.EventSubscriptions{
			&am.EventSubscriptions{
				TypeID:              am.EventInitialGroupCompleteID,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventNewHostID,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventNewWebsiteID,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventCertExpiringID,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventNewWebTechID,
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
		t.Fatalf("error notifying complete %#v", err)
	}

	returned, err := service.Get(ctx, userContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: &am.FilterType{}})
	if err != nil {
		t.Fatalf("error getting events: %v\n", err)
	}

	if len(returned) != 4 {
		t.Fatalf("expected 4 results got %v\n", len(returned))
	}

	newSiteFound := false
	newTechFound := false
	for _, e := range returned {
		t.Logf("%#v\n", e)
		if e.TypeID == am.EventNewWebsiteID {
			newSiteFound = true
			var newWeb []*am.EventNewWebsite
			json.Unmarshal([]byte(e.JSONData), &newWeb)
			if newWeb[0].LoadURL != "http://new.website.com/" && newWeb[0].Port != 80 {
				t.Fatalf("expected data to equal our new website, got %#v\n", e.JSONData)
			}
		}
		if e.TypeID == am.EventNewWebTechID {
			newTechFound = true
			var newTech []*am.EventNewWebTech
			err := json.Unmarshal([]byte(e.JSONData), &newTech)
			if err != nil {
				t.Fatalf("%#v\n", err)
			}
			t.Logf("%s\n", e.JSONData)
			t.Logf("%#v\n", newTech)
			if newTech[0].TechName != "AngularJS" {
				t.Fatalf("expected data to equal our new AngularJS, got %#v\n", e.Data)
			}
		}
	}

	if !newTechFound {
		t.Fatalf("failed to find new tech event")
	}

	if !newSiteFound {
		t.Fatalf("failed to find new site event")
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

	if err := service.MarkRead(ctx, userContext, []int64{returned[0].NotificationID}); err != nil {
		t.Fatalf("error marking notifications as read: %v\n", err)
	}

	returned, err = service.Get(ctx, userContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: &am.FilterType{}})
	if err != nil {
		t.Fatalf("error getting events after mark read: %v\n", err)
	}

	if len(returned) != 0 {
		t.Fatalf("error events returned after mark read should be 0 got: %d\n", len(returned))
	}

	// Test re-adding the same 'new' host from before, changing nothing so we should no longer have any new events
	now = time.Now() // update time
	newWebHost = amtest.CreateWebData(addr, "new.website.com", "1.1.1.1")
	newWebHost.DetectedTech = map[string]*am.WebTech{"AngularJS": &am.WebTech{
		Matched:  "1.5.3",
		Version:  "1.5.3",
		Location: "script",
	}}
	if _, err := webService.Add(ctx, userContext, newWebHost); err != nil {
		t.Fatalf("error adding single new host webdata for notify complete")
	}

	err = service.NotifyComplete(ctx, userContext, now.UnixNano(), groupID)
	if err != nil {
		t.Fatalf("error notifying complete %#v", err)
	}

	certonly, err := service.Get(ctx, userContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: &am.FilterType{}})
	if err != nil {
		t.Fatalf("error getting events: %v\n", err)
	}

	if len(certonly) != 1 {
		t.Fatalf("expected 1 results got %v\n", len(certonly))
	}

	if certonly[0].TypeID != am.EventCertExpiringID {
		t.Fatalf("expecting 1 event of cert expiring (%d), got %d\n", am.EventCertExpiredID, certonly[0].TypeID)
	}

}

func TestNotifyCompletePorts(t *testing.T) {
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()

	orgName := "eventnotifycompleteports"
	groupName := "eventnotifycompleteports"

	auth := amtest.MockEmptyAuthorizer()

	service := event.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing event service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
	defer db.Close()

	addrService := address.New(auth)
	if err := addrService.Init([]byte(addrDBString)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}

	amtest.CreateOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	defer amtest.DeleteOrg(db, orgName, t)
	userID := amtest.GetUserId(db, orgID, orgName, t)

	groupID := amtest.CreateScanGroup(db, orgName, groupName, t)
	userContext := amtest.CreateUserContext(orgID, userID)
	now := time.Now()

	portResults := &am.PortResults{
		OrgID:       orgID,
		GroupID:     groupID,
		HostAddress: "example.com",
		Ports: &am.Ports{
			Current: &am.PortData{
				IPAddress:  "1.1.1.2",
				TCPPorts:   []int32{443, 8080},
				UDPPorts:   nil,
				TCPBanners: nil,
				UDPBanners: nil,
			},
		},
		ScannedTimestamp:         time.Now().Add(time.Hour * -1).UnixNano(),
		PreviousScannedTimestamp: 0,
	}
	if _, err := addrService.UpdateHostPorts(ctx, userContext, nil, portResults); err != nil {
		t.Fatalf("error adding ports %#v\n", err)
	}

	// update again for changes
	portResults.Ports.Current = &am.PortData{
		IPAddress:  "1.1.1.1",
		TCPPorts:   []int32{80, 443, 9090},
		UDPPorts:   nil,
		TCPBanners: nil,
		UDPBanners: nil,
	}
	portResults.ScannedTimestamp = time.Now().UnixNano()

	if _, err := addrService.UpdateHostPorts(ctx, userContext, nil, portResults); err != nil {
		t.Fatalf("error adding ports again %#v\n", err)
	}

	settings := &am.UserEventSettings{
		WeeklyReportSendDay: 0,
		ShouldWeeklyEmail:   true,
		DailyReportSendHour: 0,
		ShouldDailyEmail:    true,
		UserTimezone:        "Asia/Tokyo",
		Subscriptions: []*am.EventSubscriptions{
			&am.EventSubscriptions{
				TypeID:              am.EventNewOpenPortID,
				Subscribed:          true,
				SubscribedTimestamp: now.UnixNano(),
			},
			&am.EventSubscriptions{
				TypeID:              am.EventClosedPortID,
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
		t.Fatalf("error notifying complete %#v", err)
	}

	returned, err := service.Get(ctx, userContext, &am.EventFilter{Start: 0, Limit: 1000, Filters: &am.FilterType{}})
	if err != nil {
		t.Fatalf("error getting events: %v\n", err)
	}

	if len(returned) != 2 {
		t.Fatalf("expected 2 results got %v\n", len(returned))
	}

	t.Logf("%#v", returned[0])
	t.Logf("%#v", returned[1])

}

func TestDeletePopulated(t *testing.T) {
	t.Skip("disabled for testing mailreports")
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	auth := amtest.MockEmptyAuthorizer()
	service := event.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing event service: %s\n", err)
	}

	db := amtest.InitDB(env, t)

	for i := 0; i < 5; i++ {
		orgName := fmt.Sprintf("eventpopulate%d", i)
		amtest.DeleteOrg(db, orgName, t)
	}
}

func TestPopulate(t *testing.T) {
	t.Skip("disabled for testing mailreports")
	if os.Getenv("INFRA_TESTS") == "" {
		t.Skip("skipping infrastructure tests")
	}

	ctx := context.Background()
	auth := amtest.MockEmptyAuthorizer()

	service := event.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing event service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
	defer db.Close()

	for i := 0; i < 5; i++ {

		orgName := fmt.Sprintf("eventpopulate%d", i)
		groupName := fmt.Sprintf("eventpopulate%d", i)

		webService := webdata.New(auth)
		if err := webService.Init([]byte(webDBString)); err != nil {
			t.Fatalf("error initalizing webdata service: %s\n", err)
		}

		addrService := address.New(auth)
		if err := addrService.Init([]byte(addrDBString)); err != nil {
			t.Fatalf("error initalizing address service: %s\n", err)
		}

		amtest.CreateOrg(db, orgName, t)
		orgID := amtest.GetOrgID(db, orgName, t)
		//defer amtest.DeleteOrg(db, orgName, t)
		userID := amtest.GetUserId(db, orgID, orgName, t)

		groupID := amtest.CreateScanGroup(db, orgName, groupName, t)
		userContext := amtest.CreateUserContext(orgID, userID)
		now := time.Now()

		// populate with fake data
		addr := amtest.CreateScanGroupAddress(db, orgID, groupID, t)
		addr.HostAddress = fmt.Sprintf("%s.eample.com", orgName)
		webData := amtest.CreateMultiWebData(addr, addr.HostAddress, addr.IPAddress)
		for _, web := range webData {
			if _, err := webService.Add(ctx, userContext, web); err != nil {
				t.Fatalf("error adding webdata for notify complete")
			}
		}

		// only this host should be found, since multiwebdata will create hosts that will exist < end of start scan time.
		newWebHost := amtest.CreateWebData(addr, "new.website.com", "1.1.1.1")
		newWebHost.DetectedTech = map[string]*am.WebTech{"AngularJS": &am.WebTech{
			Matched:  "1.5.3",
			Version:  "1.5.3",
			Location: "script",
		}}
		if _, err := webService.Add(ctx, userContext, newWebHost); err != nil {
			t.Fatalf("error adding single new host webdata for notify complete")
		}

		portResults := &am.PortResults{
			OrgID:       orgID,
			GroupID:     groupID,
			HostAddress: "example.com",
			Ports: &am.Ports{
				Current: &am.PortData{
					IPAddress:  "1.1.1.2",
					TCPPorts:   []int32{443, 8080},
					UDPPorts:   nil,
					TCPBanners: nil,
					UDPBanners: nil,
				},
			},
			ScannedTimestamp:         time.Now().Add(time.Hour * -1).UnixNano(),
			PreviousScannedTimestamp: 0,
		}

		if _, err := addrService.UpdateHostPorts(ctx, userContext, nil, portResults); err != nil {
			t.Fatalf("error adding ports %#v\n", err)
		}

		// update again for changes
		portResults.Ports.Current = &am.PortData{
			IPAddress:  "1.1.1.1",
			TCPPorts:   []int32{80, 443, 9090},
			UDPPorts:   nil,
			TCPBanners: nil,
			UDPBanners: nil,
		}
		portResults.ScannedTimestamp = time.Now().UnixNano()

		if _, err := addrService.UpdateHostPorts(ctx, userContext, nil, portResults); err != nil {
			t.Fatalf("error adding ports again %#v\n", err)
		}

		portResults2 := &am.PortResults{
			OrgID:       orgID,
			GroupID:     groupID,
			HostAddress: "example1.com",
			Ports: &am.Ports{
				Current: &am.PortData{
					IPAddress:  "1.1.1.1",
					TCPPorts:   []int32{443, 8080},
					UDPPorts:   nil,
					TCPBanners: nil,
					UDPBanners: nil,
				},
			},
			ScannedTimestamp:         time.Now().Add(time.Hour * -1).UnixNano(),
			PreviousScannedTimestamp: 0,
		}
		if _, err := addrService.UpdateHostPorts(ctx, userContext, nil, portResults2); err != nil {
			t.Fatalf("error adding ports %#v\n", err)
		}

		// update again for changes
		portResults2.Ports.Current = &am.PortData{
			IPAddress:  "1.1.1.1",
			TCPPorts:   []int32{80, 443, 9090},
			UDPPorts:   nil,
			TCPBanners: nil,
			UDPBanners: nil,
		}
		portResults2.ScannedTimestamp = time.Now().UnixNano()

		if _, err := addrService.UpdateHostPorts(ctx, userContext, nil, portResults2); err != nil {
			t.Fatalf("error adding ports again %#v\n", err)
		}

		settings := &am.UserEventSettings{
			WeeklyReportSendDay: 0,
			ShouldWeeklyEmail:   true,
			DailyReportSendHour: 0,
			ShouldDailyEmail:    true,
			UserTimezone:        "Asia/Tokyo",
			Subscriptions: []*am.EventSubscriptions{
				&am.EventSubscriptions{
					TypeID:              am.EventInitialGroupCompleteID,
					Subscribed:          true,
					SubscribedTimestamp: now.UnixNano(),
				},
				&am.EventSubscriptions{
					TypeID:              am.EventNewHostID,
					Subscribed:          true,
					SubscribedTimestamp: now.UnixNano(),
				},
				&am.EventSubscriptions{
					TypeID:              am.EventNewWebsiteID,
					Subscribed:          true,
					SubscribedTimestamp: now.UnixNano(),
				},
				&am.EventSubscriptions{
					TypeID:              am.EventCertExpiringID,
					Subscribed:          true,
					SubscribedTimestamp: now.UnixNano(),
				},
				&am.EventSubscriptions{
					TypeID:              am.EventNewWebTechID,
					Subscribed:          true,
					SubscribedTimestamp: now.UnixNano(),
				},
				&am.EventSubscriptions{
					TypeID:              am.EventNewOpenPortID,
					Subscribed:          true,
					SubscribedTimestamp: now.UnixNano(),
				},
				&am.EventSubscriptions{
					TypeID:              am.EventClosedPortID,
					Subscribed:          true,
					SubscribedTimestamp: now.UnixNano(),
				},
			},
		}
		// skip 1 org to make sure they don't get emails
		if i != 4 {
			t.Logf("updating settings for %s\n", orgName)
			if err := service.UpdateSettings(ctx, userContext, settings); err != nil {
				t.Fatalf("error updating user settings: %v\n", err)
			}
		}

		err := service.NotifyComplete(ctx, userContext, now.UnixNano(), groupID)
		if err != nil {
			t.Fatalf("error notifying complete")
		}
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
