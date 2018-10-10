package certtranscoordinator

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/google/certificate-transparency-go/client"
	"github.com/google/certificate-transparency-go/jsonclient"
	"github.com/google/certificate-transparency-go/loglist"
	"github.com/google/certificate-transparency-go/x509util"
	"github.com/linkai-io/am/am"
	"github.com/rs/zerolog/log"
)

var (
	ErrMaxAcquireReached = errors.New("max number of acquired servers reached")
)

// CTServers holds a pool of ct servers
type CTServers struct {
	servers     chan *am.CTServer
	httpClient  *http.Client
	serverCount int32
}

// NewCTServers creates a pool for managing access to servers
func NewCTServers() *CTServers {
	s := &CTServers{}
	s.servers = make(chan *am.CTServer, 100)
	s.httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   50 * time.Second,
			DisableKeepAlives:     false,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
			ExpectContinueTimeout: 5 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
	}
	return s
}

// Acquire the server from the pool
func (s *CTServers) Acquire(ctx context.Context) *am.CTServer {
	select {
	case <-ctx.Done():
		return nil
	case serv := <-s.servers:
		atomic.AddInt32(&s.serverCount, -1)
		return serv
	}
}

// Return the server to the pool
func (s *CTServers) Return(server *am.CTServer) {
	s.servers <- server
	atomic.AddInt32(&s.serverCount, 1)
}

// Drain our list of servers so no one else can access
func (s *CTServers) Drain(ctx context.Context) map[string]*am.CTServer {
	servers := make(map[string]*am.CTServer, 0)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	for serv := s.Acquire(ctx); serv != nil; {
		servers[serv.URL] = serv
	}
	return servers
}

// Len returns the length of the channel holding our servers
func (s *CTServers) Len() int {
	return len(s.servers)
}

// UpdateServers gets the latest list of servers, drains our pool of already known servers
// then creates new entries and calls the update method. It returns all the drained servers
// so caller MUST return them via Return after processing them.
func (s *CTServers) UpdateServers(ctx context.Context) map[string]*am.CTServer {
	serverList, err := getLatestCTLogList()
	if err != nil {
		log.Warn().Err(err).Msg("unable to get latest certificate transparency server list")
	}

	ctServers := s.Drain(context.Background())

	for _, serverURL := range serverList {
		server := &am.CTServer{}
		ok := false
		if server, ok = ctServers[serverURL]; !ok {
			server = &am.CTServer{URL: serverURL, Step: 64}
			ctServers[serverURL] = server
		}
	}
	s.updateTreeSize(ctx, ctServers)

	return ctServers
}

// updateTreeSize creates a worker pool and submits 15 servers at a time to be updated
func (s *CTServers) updateTreeSize(ctx context.Context, servers map[string]*am.CTServer) {
	pool := workerpool.New(20)
	log.Info().Msg("updating tree size for servers")
	for _, server := range servers {
		pool.Submit(func() {
			s.update(ctx, server)
		})
	}
	pool.StopWait()
}

// update returns a func to update the treesize and treesize update time
func (s *CTServers) update(ctx context.Context, server *am.CTServer) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	lc, err := client.New("https://"+server.URL, s.httpClient, jsonclient.Options{})
	if err != nil {
		log.Warn().Err(err).Str("certificate_server", server.URL).Msg("unable to create client for server")
		return
	}

	head, err := lc.GetSTH(timeoutCtx)
	if err != nil {
		log.Warn().Err(err).Str("certificate_server", server.URL).Msg("unable to get tree index")
		return
	}
	server.TreeSize = int64(head.TreeSize)
	server.TreeSizeUpdated = time.Now().UnixNano()
	log.Info().Str("certificate_server", server.URL).Msgf("updated %v\n", server)
}

func getLatestCTLogList() ([]string, error) {
	hc := &http.Client{}

	llData, err := x509util.ReadFileOrURL(loglist.AllLogListURL, hc)
	if err != nil {
		return nil, err
	}

	ll, err := loglist.NewFromJSON(llData)
	if err != nil {
		return nil, err
	}

	list := make([]string, len(ll.Logs))
	for i := 0; i < len(ll.Logs); i++ {
		list[i] = ll.Logs[i].URL
	}
	return list, nil
}
