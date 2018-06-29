package pg

import (
	"encoding/json"
	"errors"

	"github.com/jackc/pgx"
)

var (
	ErrEmptyDBAddr = errors.New("empty db_addr field")
	ErrEmptyDBUser = errors.New("empty db_user field")
	ErrEmptyDBPass = errors.New("empty db_pass field")
	ErrEmptyDBName = errors.New("empty db_name field")
)

// Config represents this modules configuration data to be passed in on
// initialization.
type Config struct {
	Addr           string `json:"db_addr"`
	User           string `json:"db_user"`
	Pass           string `json:"db_pass"`
	Database       string `json:"db_name"`
	MaxConnections int    `json:"db_max_conn"`
}

// Store for interfacing with postgresql/rds
type Store struct {
	pool   *pgx.ConnPool
	config *pgx.ConnPoolConfig
}

// New returns an empty store
func New() *Store {
	return &Store{}
}

// Init by parsing the config and initializing the database pool
func (s *Store) Init(config []byte) error {
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
func (s *Store) parseConfig(config []byte) (*pgx.ConnPoolConfig, error) {
	var v *Config
	if err := json.Unmarshal(config, v); err != nil {
		return nil, err
	}

	if v.Addr == "" {
		return nil, ErrEmptyDBAddr
	}

	if v.Database == "" {
		return nil, ErrEmptyDBName
	}

	if v.User == "" {
		return nil, ErrEmptyDBUser
	}

	if v.Pass == "" {
		return nil, ErrEmptyDBPass
	}

	if v.MaxConnections == 0 {
		v.MaxConnections = 50
	}

	return &pgx.ConnPoolConfig{ConnConfig: pgx.ConnConfig{
		Host:     v.Addr,
		User:     v.User,
		Password: v.Pass,
		Database: v.Database,
	},
		MaxConnections: v.MaxConnections,
		AfterConnect:   s.afterConnect,
	}, nil
}

// afterConnect will iterate over prepared statements with keywords
func (s *Store) afterConnect(conn *pgx.Conn) error {
	return nil
}
