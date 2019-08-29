package webhooks_test

import (
	"context"
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/webhooks"
)

func TestSendLocal(t *testing.T) {
	env := "local"
	c := webhooks.New(env, "")
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
	resp, err := c.Send(ctx, evt)
	if err != nil {
		t.Fatalf("failed to send webhook event")
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200 response code got %v\n", resp.StatusCode)
	}
}
