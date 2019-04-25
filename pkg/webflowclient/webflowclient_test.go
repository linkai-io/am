package webflowclient_test

import (
	"context"
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/filestorage"
	"github.com/linkai-io/am/pkg/webflowclient"
)

func TestDo(t *testing.T) {
	s := filestorage.NewLocalStorage()
	if err := s.Init(); err != nil {
		t.Fatalf("error initing local storage")
	}

	c := webflowclient.New(s)

	cfg := &am.CustomRequestConfig{
		Method: "GET",
		URI:    "/",
		Headers: map[string]string{
			"": "",
		},
		Body: "",
		Match: map[int32]string{
			am.CustomMatchStatusCode: "401",
			am.CustomMatchString:     "error",
			am.CustomMatchRegex:      "(?i)(amazonS3.*)|(.*Cloudfront.*)",
		},
		OnlyPort:   0,
		OnlyScheme: "",
	}
	event := &webflowclient.RequestEvent{
		Host:   "dev.console.linkai.io",
		Config: cfg,
		Ports:  []int32{80, 443},
		UserContext: &am.UserContextData{
			OrgID:   1,
			UserID:  1,
			OrgCID:  "/tmp",
			TraceID: "1234-test",
		},
	}
	r, err := c.Do(context.Background(), event)
	if err != nil {
		t.Fatalf("error Do: %#v\n", err)
	}
	for _, res := range r.Results {
		t.Logf("%#v\n", res)
		if res.URL == "https://dev.console.linkai.io:443/" && len(res.Result) != 2 {
			t.Fatalf("expected 2 matches, got: %d\n", len(res.Result))
		}
		for _, m := range res.Result {
			t.Logf("\tmatch: %#v", m)
		}
	}

}

func TestDoBannedIP(t *testing.T) {
	s := filestorage.NewLocalStorage()
	if err := s.Init(); err != nil {
		t.Fatalf("error initing local storage")
	}

	c := webflowclient.New(s)

	cfg := &am.CustomRequestConfig{
		Method: "GET",
		URI:    "/",
		Headers: map[string]string{
			"": "",
		},
		Body: "",
		Match: map[int32]string{
			am.CustomMatchStatusCode: "401",
			am.CustomMatchString:     "error",
			am.CustomMatchRegex:      "(?i)(amazonS3.*)|(.*Cloudfront.*)",
		},
		OnlyPort:   0,
		OnlyScheme: "",
	}
	event := &webflowclient.RequestEvent{
		Host:   "localhost",
		Config: cfg,
		Ports:  []int32{80, 443},
		UserContext: &am.UserContextData{
			OrgID:   1,
			UserID:  1,
			OrgCID:  "/tmp",
			TraceID: "1234-test",
		},
	}

	r, err := c.Do(context.Background(), event)
	if err != nil {
		t.Fatalf("error Do: %#v\n", err)
	}

	for _, res := range r.Results {
		t.Logf("%#v\n", res)
		if res.URL == "https://dev.console.linkai.io:443/" && len(res.Result) != 2 {
			t.Fatalf("expected 2 matches, got: %d\n", len(res.Result))
		}
		for _, m := range res.Result {
			t.Logf("\tmatch: %#v", m)
		}
	}

}
