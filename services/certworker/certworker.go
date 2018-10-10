package certworker

import (
	"context"
	"sync/atomic"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Service for interfacing with postgresql/rds
type Service struct {
	pool          *pgx.ConnPool
	config        *pgx.ConnPoolConfig
	numExtractors int32
}

// New returns an empty Service
func New() *Service {
	return &Service{}
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

// GetCTCertificates starts a new extractor and iterates numExtractors times to parse and upload
// certificates
func (s *Service) GetCTCertificates(ctx context.Context, ctServer *am.CTServer) (*am.CTServer, error) {
	extractor := NewExtractor(nil, ctServer, int(atomic.LoadInt32(&s.numExtractors)))
	return extractor.Run(ctx)
}

// SetExtractors allows modifying the number of extractors to use per call of GetCTCertificates. Note
// it will not dynamically update if a call is already in progress.
func (s *Service) SetExtractors(ctx context.Context, numExtractors int32) error {
	if numExtractors <= 0 || numExtractors > 200 {
		return errors.New("invalid number of extractors")
	}

	atomic.StoreInt32(&s.numExtractors, numExtractors)
	log.Info().Int32("num", numExtractors).Msg("updated extractor count")
	return nil
}
