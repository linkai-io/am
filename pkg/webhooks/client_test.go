package webhooks_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/webhooks"
)

func TestSendEvent(t *testing.T) {

	c := webhooks.NewClient()
	eventData := makeEvents()

	evt := &webhooks.Data{
		Settings: &am.WebhookEventSettings{
			URL:           "https://hooks.slack.com/services/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",
			Version:       "v1",
			Type:          "slack",
			ScanGroupName: "test",
		},
		Event: eventData,
	}
	ctx := context.Background()
	respCode, err := c.SendEvent(ctx, evt)
	if err != nil {
		t.Fatalf("failed to send webhook event")
	}
	if respCode != 200 {
		t.Fatalf("expected 200 response code got %v\n", respCode)
	}
}

func makeEvents() []*am.Event {
	events := make([]*am.Event, 7)
	m, _ := json.Marshal([]*am.EventAXFR{&am.EventAXFR{Servers: []string{"ns3.example.com", "ns4.example.com"}}})
	events[0] = &am.Event{
		NotificationID: 7,
		OrgID:          0,
		GroupID:        0,
		TypeID:         am.EventAXFRID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
		Read:           false,
	}

	m, _ = json.Marshal([]*am.EventNewHost{
		&am.EventNewHost{Host: "json.example.com"},
		&am.EventNewHost{Host: "json.test.example.com"},
		&am.EventNewHost{Host: "json.something.example.com"},
	})
	events[1] = &am.Event{
		NotificationID: 8,
		OrgID:          0,
		GroupID:        0,
		TypeID:         am.EventNewHostID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
		Read:           false,
	}

	m, _ = json.Marshal([]*am.EventNewWebsite{
		&am.EventNewWebsite{LoadURL: "https://json.example.com", URL: "https://json.example.com/", Port: 443},
		&am.EventNewWebsite{LoadURL: "http://json.redirect.example.com", URL: "https://json.redirect.example.com:443/", Port: 443},
	})

	events[2] = &am.Event{
		NotificationID: 9,
		OrgID:          0,
		GroupID:        0,
		TypeID:         am.EventNewWebsiteID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
		Read:           false,
	}

	m, _ = json.Marshal([]*am.EventCertExpiring{
		&am.EventCertExpiring{SubjectName: "json.example.com", ValidTo: time.Now().Add(24 * time.Hour).Unix(), Port: 443},
	})
	events[3] = &am.Event{
		NotificationID: 10,
		OrgID:          0,
		GroupID:        0,
		TypeID:         am.EventCertExpiringID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
		Read:           false,
	}

	m, _ = json.Marshal([]*am.EventNewWebTech{
		&am.EventNewWebTech{LoadURL: "https://json.example.com", URL: "https://json.example.com/", Port: 443, TechName: "jQuery", Version: "1.2.3"},
		&am.EventNewWebTech{LoadURL: "http://json.example.com", URL: "https://json.example.com/", Port: 443, TechName: "jQuery", Version: "1.2.3"},
	})
	events[4] = &am.Event{
		NotificationID: 11,
		OrgID:          0,
		GroupID:        0,
		TypeID:         am.EventNewWebTechID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
		Read:           false,
	}
	m, _ = json.Marshal([]*am.EventNewOpenPort{
		&am.EventNewOpenPort{Host: "json.example.com", CurrentIP: "1.1.1.1", PreviousIP: "1.1.1.2", OpenPorts: []int32{8080, 9000}},
		&am.EventNewOpenPort{Host: "json1.example.com", CurrentIP: "1.1.1.1", PreviousIP: "1.1.1.1", OpenPorts: []int32{23, 22}},
	})
	events[5] = &am.Event{
		NotificationID: 12,
		OrgID:          0,
		GroupID:        0,
		TypeID:         am.EventNewOpenPortID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
		Read:           false,
	}

	m, _ = json.Marshal([]*am.EventClosedPort{
		&am.EventClosedPort{Host: "json.example.com", CurrentIP: "1.1.1.1", PreviousIP: "1.1.1.2", ClosedPorts: []int32{12, 23}},
		&am.EventClosedPort{Host: "json1.example.com", CurrentIP: "1.1.1.1", PreviousIP: "1.1.1.1", ClosedPorts: []int32{2222}},
	})

	events[6] = &am.Event{
		NotificationID: 13,
		OrgID:          0,
		GroupID:        0,
		TypeID:         am.EventClosedPortID,
		EventTimestamp: time.Now().UnixNano(),
		JSONData:       string(m),
		Read:           false,
	}
	return events
}
