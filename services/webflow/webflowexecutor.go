package webflow

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/rs/zerolog/log"
)

type Executor interface {
	Init() error
	Start(ctx context.Context, webFlowConfig *am.CustomWebFlowConfig) error
}

type WebFlowExecutor struct {
	userContext     am.UserContext
	webFlowService  am.CustomWebFlowService
	addressClient   am.AddressService
	scanGroupClient am.ScanGroupService
	closeCh         chan struct{}
	stopCh          chan struct{}

	webFlowLock   *sync.RWMutex
	webFlowConfig *am.CustomWebFlowConfig
}

func NewWebFlowExecutor(userContext am.UserContext, webFlowService am.CustomWebFlowService, addressClient am.AddressService, scanGroupClient am.ScanGroupService) *WebFlowExecutor {
	return &WebFlowExecutor{
		userContext:     userContext,
		webFlowService:  webFlowService,
		addressClient:   addressClient,
		scanGroupClient: scanGroupClient,
		webFlowLock:     &sync.RWMutex{},
		closeCh:         make(chan struct{}),
		stopCh:          make(chan struct{}),
	}
}

func (e *WebFlowExecutor) Init() error {
	go e.monitorFlows()
	return nil
}

func (e *WebFlowExecutor) monitorFlows() {
	t := time.NewTicker(time.Second * 30)
	stackTicker := time.NewTicker(time.Minute * 15)
	defer t.Stop()
	defer stackTicker.Stop()

	for {
		select {
		case <-e.closeCh:
			return
		case <-stackTicker.C:
			buf := make([]byte, 1<<20)
			stacklen := runtime.Stack(buf, true)
			log.Printf("*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
		case <-t.C:

			timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*20)
			_, status, err := e.webFlowService.GetStatus(timeoutCtx, e.userContext, e.webFlowConfig.WebFlowID)
			cancel()
			if err != nil {
				log.Warn().Err(err).Msg("error getting group status")
				continue
			}

			if status.WebFlowStatus == am.WebFlowStatusStopped {
				close(e.stopCh)
			}

		}
	}
}

func (e *WebFlowExecutor) Start(ctx context.Context, webFlowConfig *am.CustomWebFlowConfig) error {

	oid, group, err := e.scanGroupClient.Get(ctx, e.userContext, webFlowConfig.GroupID)
	if err != nil {
		return err
	}

	if oid != e.userContext.GetOrgID() {
		return am.ErrOrgIDMismatch
	}
	e.webFlowConfig = webFlowConfig
	go e.run(group)
	return nil
}

func (e *WebFlowExecutor) run(group *am.ScanGroup) {
	ctx := context.Background()
	groupLog := log.With().
		Int("UserID", e.userContext.GetUserID()).
		Int("GroupID", group.GroupID).
		Int32("WebFlowID", e.webFlowConfig.WebFlowID).
		Int("OrgID", e.userContext.GetOrgID()).
		Str("TraceID", e.userContext.GetTraceID()).Logger()

	ctx = groupLog.WithContext(ctx)

	for {
		filter := &am.ScanGroupAddressFilter{
			OrgID:   e.userContext.GetOrgID(),
			GroupID: group.GroupID,
			Start:   0,
			Limit:   1000,
			Filters: &am.FilterType{},
		}

		_, hosts, err := e.addressClient.GetHostList(ctx, e.userContext, filter)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("error getting hosts from client")
			return
		}

		if hosts == nil || len(hosts) == 0 {
			break
		}
	}

}
