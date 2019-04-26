package webflow

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/linkai-io/am/pkg/webflowclient"

	"github.com/linkai-io/am/am"
	"github.com/rs/zerolog/log"
)

type WebFlowRequester interface {
	Do(ctx context.Context, request webflowclient.RequestEvent) (*webflowclient.Results, error)
}

type Executor interface {
	Init() error
	Start(ctx context.Context, webFlowConfig *am.CustomWebFlowConfig) error
}

type WebFlowExecutor struct {
	requester       WebFlowRequester
	userContext     am.UserContext
	webFlowService  am.CustomWebFlowService
	addressClient   am.AddressService
	scanGroupClient am.ScanGroupService
	shouldStop      int32
	closeCh         chan struct{}
	stopCh          chan struct{}
	total           int32
	inProgress      int32
	completed       int32

	webFlowLock   *sync.RWMutex
	webFlowConfig *am.CustomWebFlowConfig
}

func NewWebFlowExecutor(userContext am.UserContext, webFlowService am.CustomWebFlowService, addressClient am.AddressService, scanGroupClient am.ScanGroupService, requester WebFlowRequester) *WebFlowExecutor {
	return &WebFlowExecutor{
		userContext:     userContext,
		requester:       requester,
		webFlowService:  webFlowService,
		addressClient:   addressClient,
		scanGroupClient: scanGroupClient,
		webFlowLock:     &sync.RWMutex{},
		closeCh:         make(chan struct{}),
		stopCh:          make(chan struct{}),
	}
}

func (e *WebFlowExecutor) Init() error {
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
				atomic.AddInt32(&e.shouldStop, am.WebFlowStatusStopped)
				return
			}
			total := atomic.LoadInt32(&e.total)
			inProgress := atomic.LoadInt32(&e.inProgress)
			completed := atomic.LoadInt32(&e.completed)
			timeoutCtx, cancel = context.WithTimeout(context.Background(), time.Second*20)
			if err := e.webFlowService.UpdateStatus(timeoutCtx, e.userContext, total, inProgress, completed, e.webFlowConfig.WebFlowID); err != nil {
				log.Warn().Err(err).Int32("web_flow_id", e.webFlowConfig.WebFlowID).Msg("failed to update status for web flow")
			}
			cancel()
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
	if e.webFlowConfig == nil {
		return am.ErrEmptyCustomWebFlowConfig
	}
	go e.monitorFlows()
	go e.run(group)
	return nil
}

func (e *WebFlowExecutor) run(group *am.ScanGroup) {
	ctx := context.Background()

	groupLog := log.With().
		Int("UserID", e.userContext.GetUserID()).
		Int("GroupID", group.GroupID).
		Int32("WebFlowID", e.webFlowConfig.WebFlowID).
		Int("OrgID", e.userContext.GetOrgID()).Logger()

	ctx = groupLog.WithContext(ctx)

	ports := make([]int32, 2)
	ports[0] = int32(80)
	ports[1] = int32(443)

	if e.webFlowConfig.Configuration.OnlyPort == 0 && group.ModuleConfigurations.PortModule.CustomPorts != nil {
		ports = append(ports, group.ModuleConfigurations.PortModule.CustomPorts...)
	}

	filter := &am.ScanGroupAddressFilter{
		OrgID:   e.userContext.GetOrgID(),
		GroupID: group.GroupID,
		Start:   0,
		Limit:   1000,
		Filters: &am.FilterType{},
	}

	lastHost := ""
	for {
		if atomic.LoadInt32(&e.shouldStop) == am.WebFlowStatusStopped {
			break
		}
		_, hosts, err := e.addressClient.GetHostList(ctx, e.userContext, filter)
		if err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("error getting hosts from client")
			break
		}

		log.Ctx(ctx).Info().Int("hosts", len(hosts)).Msg("got hosts from client")
		if hosts == nil || len(hosts) == 0 {
			break
		}

		lastHost = hosts[len(hosts)-1].HostAddress
		filter.Filters.AddString("start_host", lastHost)
		log.Ctx(ctx).Info().Msg("pooling host requests")
		e.poolRequests(ctx, group, hosts, ports)
	}

	if _, err := e.webFlowService.Stop(ctx, e.userContext, e.webFlowConfig.WebFlowID); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to stop custom web flow")
	}
}

func (e *WebFlowExecutor) poolRequests(ctx context.Context, group *am.ScanGroup, hosts []*am.ScanGroupHostList, ports []int32) {
	rps := int(group.ModuleConfigurations.NSModule.RequestsPerSecond)
	numHosts := len(hosts)
	atomic.AddInt32(&e.total, int32(numHosts))

	if len(hosts) < rps {
		rps = numHosts
	}
	pool := workerpool.New(rps)

	out := make(chan *webflowclient.Results, numHosts)
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(30*numHosts))
	defer cancel()

	log.Ctx(ctx).Info().Msg("iterating hosts")
	for _, host := range hosts {
		if atomic.LoadInt32(&e.shouldStop) == am.WebFlowStatusStopped {
			return
		}

		hostName := host.HostAddress
		if hostName == "" {
			continue
		}

		log.Ctx(ctx).Info().Msgf("queueing %s", hostName)
		req := &webflowclient.RequestEvent{UserContext: e.userContext, Host: hostName, Ports: ports, Config: e.webFlowConfig.Configuration}
		atomic.AddInt32(&e.inProgress, 1)
		pool.Submit(e.executeRequest(timeoutCtx, req, out))
	}
	log.Ctx(ctx).Info().Msg("queued all hosts")
	pool.StopWait()
	close(out)

	for res := range out {
		if err := e.webFlowService.AddResults(ctx, e.userContext, res.Results); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("failed to add results")
		}
	}
}

func (e *WebFlowExecutor) executeRequest(ctx context.Context, req *webflowclient.RequestEvent, out chan *webflowclient.Results) func() {
	return func() {
		log.Ctx(ctx).Info().Msgf("doing request: %s", req.Host)
		result, err := e.requester.Do(ctx, *req)
		atomic.AddInt32(&e.inProgress, -1)
		atomic.AddInt32(&e.completed, 1)
		if err != nil {
			return
		}
		select {
		case <-e.stopCh:
			return
		case <-ctx.Done():
			return
		case out <- result:
			return
		}
	}
}
