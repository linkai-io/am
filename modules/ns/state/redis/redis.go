package redis

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/linkai-io/am/pkg/redisclient"
)

var (
	//ErrEmptyRCAddress missing redis address
	ErrEmptyRCAddress = errors.New("rc_addr was empty or invalid")
	// ErrEmptyRCPassword missing redis password
	ErrEmptyRCPassword = errors.New("rc_pass was empty or invalid")
)

// Config represents this modules configuration data to be passed in on
// initialization.
type Config struct {
	RCAddr string `json:"rc_addr"`
	RCPass string `json:"rc_pass"`
}

// State manager
type State struct {
	rc *redisclient.Client
}

// New redis backed state
func New() *State {
	return &State{}
}

// Init by parsing the config and initializing the redis client
func (s *State) Init(config []byte) error {
	stateConfig, err := s.parseConfig(config)
	if err != nil {
		return err
	}

	s.rc = redisclient.New(stateConfig.RCAddr, stateConfig.RCPass)

	return s.rc.Init()
}

// parseConfig parses the configuration options and validates they are sane.
func (rs *State) parseConfig(config []byte) (*Config, error) {
	v := &Config{}
	if err := json.Unmarshal(config, v); err != nil {
		return nil, err
	}

	if v.RCAddr == "" {
		return nil, ErrEmptyRCAddress
	}

	if v.RCPass == "" {
		return nil, ErrEmptyRCPassword
	}
	return v, nil
}

// IsValid checks if the zone is valid for testing (not done before)
// and not an ignored zone (aws, etc).
func (s *State) IsValid(zone string) bool {
	conn := s.rc.Get()
	s.rc.Return(conn)
	return true
}

// DoNSRecords org:group:ns:zone:<zonename> recorded bool EXPIRE 4 H?
func (s *State) DoNSRecords(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return false, err
	}
	defer s.rc.Return(conn)

	// create redis keys for this org/group
	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
	ret, err := redis.String(conn.Do("SET", keys.NSZone(zone), time.Now().UnixNano(), "NX", "PX", expireSeconds))
	if err != nil {
		// redis will return ErrNil if value is already set.
		if err == redis.ErrNil {
			return false, nil
		}
		return false, err
	}
	log.Printf("%#v\n", ret)
	return ret == "OK", nil
}
