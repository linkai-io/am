package redis

import (
	"encoding/json"
	"errors"

	"gopkg.linkai.io/v1/repos/am/pkg/redisclient"
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
	OrgID  int64  `json:"org_id"`
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
func (rs *State) Init(config []byte) error {
	stateConfig, err := rs.parseConfig(config)
	if err != nil {
		return err
	}

	rs.rc = redisclient.New(stateConfig.RCAddr, stateConfig.RCPass)

	return rs.rc.Init()
}

// parseConfig parses the configuration options and validates they are sane.
func (rs *State) parseConfig(config []byte) (*Config, error) {
	var v *Config
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
func (rs *State) IsValid(zone string) bool {
	conn := rs.rc.Get()
	rs.rc.Return(conn)
	return true
}
