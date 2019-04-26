package webflow_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/webflowclient"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/services/webflow"
)

type testRequester struct {
	wg *sync.WaitGroup
	t  *testing.T
}

func (e *testRequester) Do(ctx context.Context, event webflowclient.RequestEvent) (*webflowclient.Results, error) {
	log.Printf("AHA")
	e.t.Logf("do called %#v %#v\n", event, event.Config)
	e.wg.Done()
	return &webflowclient.Results{}, nil
}

func TestExecutor(t *testing.T) {
	addrLen := 5
	addrs := amtest.GenerateAddrs(1, 1, addrLen)
	for i, a := range addrs {
		a.HostAddress = fmt.Sprintf("%d.example.com", i)
	}
	wg := &sync.WaitGroup{}
	wg.Add(addrLen)

	addrClient := amtest.MockAddressService(1, addrs)
	userContext := amtest.CreateUserContext(1, 1)
	sgClient := amtest.MockScanGroupService(1, 1)
	requester := &testRequester{t: t, wg: wg}

	webFlowClient := webflow.New(nil, nil, nil, requester)
	webFlowClient.Init(nil)

	executor := webflow.NewWebFlowExecutor(userContext, webFlowClient, addrClient, sgClient, requester)
	cfg := &am.CustomWebFlowConfig{
		OrgID:        1,
		GroupID:      1,
		WebFlowID:    1,
		WebFlowName:  "test",
		CreationTime: time.Now().UnixNano(),
		ModifiedTime: time.Now().UnixNano(),
		Deleted:      false,
		Configuration: &am.CustomRequestConfig{
			Method: "GET",
			URI:    "/admin",
			Headers: map[string]string{
				"blah": "xyz",
			},
			Body: "",
			Match: map[int32]string{
				0: "",
			},
			OnlyPort:   0,
			OnlyScheme: "",
		},
	}
	ctx := context.Background()
	if err := executor.Start(ctx, cfg); err != nil {
		t.Fatalf("error starting")
	}
	wg.Wait()
}
