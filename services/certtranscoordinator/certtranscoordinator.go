package certtranscoordinator

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/rs/zerolog/log"
)

var (
	ErrNoAvailableServers = errors.New("no certificate transparency servers available")
)

// Service for interfacing with postgresql/rds
type Service struct {
	pool   *pgx.ConnPool
	config *pgx.ConnPoolConfig
	// our server states
	servers        *CTServers
	updateDuration int64

	workerLock  sync.RWMutex
	workers     map[int64]chan struct{}
	stoppedCh   chan struct{}
	workerCount int32
	maxWorkers  int32
	status      int32
}

// New returns an empty Service
func New() *Service {
	return &Service{
		servers:        NewCTServers(),
		workers:        make(map[int64]chan struct{}),
		stoppedCh:      make(chan struct{}),
		maxWorkers:     5,
		updateDuration: int64(time.Minute),
	}
}

// Init by parsing the config and initializing the database pool
func (s *Service) Init(config []byte) error {
	var err error

	s.config, err = s.parseConfig(config)
	if err != nil {
		return err
	}

	if s.pool, err = pgx.NewConnPool(*s.config); err != nil {
		return err
	}

	return nil
}

// parseConfig parses the configuration options and validates they are sane.
func (s *Service) parseConfig(config []byte) (*pgx.ConnPoolConfig, error) {
	dbstring := string(config)
	if dbstring == "" {
		return nil, am.ErrEmptyDBConfig
	}

	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		return nil, am.ErrInvalidDBString
	}

	return &pgx.ConnPoolConfig{
		ConnConfig:     conf,
		MaxConnections: 50,
		AfterConnect:   s.afterConnect,
	}, nil
}

// afterConnect will iterate over prepared statements with keywords
func (s *Service) afterConnect(conn *pgx.Conn) error {
	for k, v := range queryMap {
		if _, err := conn.Prepare(k, v); err != nil {
			return err
		}
	}
	return nil
}

// initServers queries the database looking for the list of certificate servers.
// if empty, we call updateservers and insert each one into the db
// once we have a list of servers, we return them to the ctserver component to
// allow clients to aqcuire/return them.
func (s *Service) initServers() {
	servers := make(map[string]*am.CTServer, 0)

	rows, err := s.pool.Query("getServers")
	if err != nil {
		log.Warn().Err(err).Msg("unable to query servers from database")
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		server := &am.CTServer{}

		if err := rows.Scan(&server.ID, &server.URL, &server.Index, &server.IndexUpdated, &server.Step, &server.TreeSize, &server.TreeSizeUpdated); err != nil {
			log.Warn().Err(err).Msg("unable to read server state")
			continue
		}
		servers[server.URL] = server
	}

	if len(servers) == 0 {
		servers = s.servers.UpdateServers(context.Background())
		for _, server := range servers {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			id, err := s.updateServer(ctx, server)
			if err != nil {
				log.Error().Err(err).Msg("unable to insert server")
				continue
			}
			server.ID = id
		}
	}

	for _, server := range servers {
		s.servers.Return(server)
	}
}

func (s *Service) updateServer(ctx context.Context, server *am.CTServer) (int, error) {
	var id int
	err := s.pool.QueryRowEx(ctx, "insertServer", &pgx.QueryExOptions{}, server.URL, server.Index, server.IndexUpdated, server.Step, server.TreeSize, server.TreeSizeUpdated).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetServer acquires a server, checks if it should be updated, and returns it to caller
func (s *Service) GetStatus(ctx context.Context) (am.CTCoordinatorStatus, int32, error) {
	return am.CTCoordinatorStatus(atomic.LoadInt32(&s.status)), atomic.LoadInt32(&s.workerCount), nil
}

// SetStatus if status is the same, return. Otherwise store the new status, and if stopped
// range over all workers and close their channel and delete it. If set to started, then
// we need to re-make our stoppedCh and re-call Run()
func (s *Service) SetStatus(ctx context.Context, status am.CTCoordinatorStatus) (int32, error) {
	if atomic.LoadInt32(&s.status) == int32(status) {
		return 0, nil
	}

	atomic.StoreInt32(&s.status, int32(status))
	if status == am.Stopped {
		s.workerLock.Lock()
		for t, ch := range s.workers {
			delete(s.workers, t)
			close(ch)
		}
		s.workerLock.Unlock()
		return 0, nil
	}

	if status == am.Started {
		s.stoppedCh = make(chan struct{})
		s.Run()
	}
	return atomic.LoadInt32(&s.workerCount), nil
}

// Run executes s.maxWorkers in their own go routines, with a close channel
// which we can use to control it's lifecycle.
func (s *Service) Run() {
	max := int(atomic.LoadInt32(&s.maxWorkers))

	for i := 0; i < max; i++ {
		s.workerLock.Lock()
		closeCh := make(chan struct{})
		s.workers[time.Now().UnixNano()] = closeCh
		s.workerLock.Unlock()
		go s.processClient(closeCh)
		atomic.AddInt32(&s.workerCount, 1)
	}
}

// addWorker adds a new worker and starts a new go routine.
func (s *Service) addWorker() {
	closeCh := make(chan struct{})
	atomic.AddInt32(&s.maxWorkers, 1)
	s.workerLock.Lock()
	s.workers[time.Now().UnixNano()] = closeCh
	s.workerLock.Unlock()

	go s.processClient(closeCh)
	atomic.AddInt32(&s.workerCount, 1)
}

// removeWorker extracts a random worker and closes it's channel and deletes it
// from our map of workers.
func (s *Service) removeWorker() {
	atomic.AddInt32(&s.maxWorkers, -1)
	s.workerLock.Lock()
	for t, ch := range s.workers {
		close(ch)
		delete(s.workers, t)
		break
	}
	s.workerLock.Unlock()
	atomic.AddInt32(&s.workerCount, -1)
}

func (s *Service) processClient(closeCh chan struct{}) {
	for {
		select {
		case <-closeCh:
			log.Info().Msg("worker closed")
			return
		default:
		}

		ctx := context.Background()
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		server := s.servers.Acquire(timeoutCtx)
		cancel()

		if server == nil {
			continue
		}

		if s.shouldUpdateTreeSize(server) {
			s.updateTreeSize(ctx, server)
		}
		/*
			s.client.Send
			// HANDLE CLIENT STUFF HERE
			id, err := s.updateServer(ctx, server)
			if err != nil {
				log.Warn().Str("certificate_server", server.URL).Err(err).Msg("unable to update server in database")
			} else {
				server.ID = id
			}
			s.servers.Return(server)
		*/
	}
}

func (s *Service) shouldUpdateTreeSize(server *am.CTServer) bool {
	since := time.Since(time.Unix(0, server.TreeSizeUpdated))
	duration := atomic.LoadInt64(&s.updateDuration)
	if time.Duration(duration) > since {
		return true
	}
	return false
}

func (s *Service) updateTreeSize(ctx context.Context, server *am.CTServer) {
	s.servers.update(ctx, server)
}

// AddWorker adds workers depending on number of workerCount. Should not be > 10
func (s *Service) AddWorker(ctx context.Context, workerCount int) error {
	if atomic.LoadInt32(&s.status) == int32(am.Stopped) {
		return errors.New("coordinator is currently stopped")
	}

	if workerCount <= 0 || workerCount > 10 {
		return nil
	}

	workers := atomic.LoadInt32(&s.workerCount)
	if workers+int32(workerCount) > 100 {
		return errors.New("invalid number of workers")
	}

	for i := 0; i < workerCount; i++ {
		s.addWorker()
	}

	if atomic.LoadInt32(&s.workerCount) == 0 {
		atomic.StoreInt32(&s.status, int32(am.Stopped))
	}

	return nil
}

// RemoveWorker removes workers depending on number of workerCount. Should not be > 10
func (s *Service) RemoveWorker(ctx context.Context, workerCount int) error {
	if atomic.LoadInt32(&s.status) == int32(am.Stopped) {
		return errors.New("coordinator is currently stopped")
	}

	if workerCount <= 0 || workerCount > 10 {
		return nil
	}

	workers := atomic.LoadInt32(&s.workerCount)
	if workers-int32(workerCount) < 0 {
		workerCount = int(workers)
	}

	for i := 0; i < workerCount; i++ {
		s.removeWorker()
	}

	if atomic.LoadInt32(&s.workerCount) == 0 {
		atomic.StoreInt32(&s.status, int32(am.Stopped))
	}

	return nil
}

// UpdateDuration changes how often we should update the tree size for servers
func (s *Service) UpdateDuration(ctx context.Context, newDuration int64) error {
	if newDuration < int64(time.Minute) {
		return errors.New("invalid duration, too small, must be greater than 1 minute")
	}
	atomic.StoreInt64(&s.updateDuration, newDuration)
	return nil
}
