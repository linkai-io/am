package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"
	"gopkg.linkai.io/v1/repos/am/am"

	"gopkg.linkai.io/v1/repos/am/pkg/redisclient"
)

const (
	// OrgID:GroupID
	ConfigFmt      = "%d:%d:configuration"
	AddrFmt        = "%d:%d:address:" // orgid:groupid:address:md5(ip,host)
	NSConfigFmt    = ":module:nsconfig"
	BruteConfigFmt = ":module:dnsbruteconfig"
	PortConfigFmt  = ":module:portconfig"
	WebConfigFmt   = ":module:webconfig"
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
func (s *State) parseConfig(config []byte) (*Config, error) {
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

func (s *State) Start(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return err
	}
	defer s.rc.Return(conn)

	configKey := fmt.Sprintf(ConfigFmt, userContext.GetOrgID(), scanGroupID)
	_, err = conn.Do("HSET", configKey, "status", am.GroupStarted)

	return err
}

func (s *State) Stop(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return err
	}
	defer s.rc.Return(conn)

	configKey := fmt.Sprintf(ConfigFmt, userContext.GetOrgID(), scanGroupID)
	_, err = conn.Do("HSET", configKey, "status", am.GroupStopped)

	return err
}

func (s *State) Put(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) error {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return err
	}
	defer s.rc.Return(conn)

	// start transaction
	if err := conn.Send("MULTI"); err != nil {
		return err
	}

	// create redis keys for this org/group
	configKey := fmt.Sprintf(ConfigFmt, group.OrgID, group.GroupID)
	nsKey := configKey + NSConfigFmt
	bruteKey := configKey + BruteConfigFmt
	bruteKeyHosts := bruteKey + ":custom_hosts"
	portKey := configKey + PortConfigFmt
	portKeyPorts := portKey + ":custom_ports"
	webKey := configKey + WebConfigFmt

	// create primary configuration
	if err := conn.Send("HMSET", configKey, "modified_time", group.ModifiedTime, "status", am.GroupStopped); err != nil {
		return err
	}

	ns := group.ModuleConfigurations.NSModule

	if err := conn.Send("HMSET", nsKey, "requests_per_second", ns.RequestsPerSecond); err != nil {
		return err
	}

	// put dns brute config
	brute := group.ModuleConfigurations.BruteModule
	conn.Send("HMSET", bruteKey, "max_depth", brute.MaxDepth, "requests_per_second", brute.RequestsPerSecond)

	args := make([]interface{}, len(brute.CustomSubNames)+1)
	args[0] = bruteKeyHosts
	for i := 1; i < len(args); i++ {
		args[i] = brute.CustomSubNames[i-1]
	}

	if err := conn.Send("LPUSH", args...); err != nil {
		return err
	}

	// put port config
	port := group.ModuleConfigurations.PortModule
	if err := conn.Send("HMSET", portKey, "requests_per_second", port.RequestsPerSecond); err != nil {
		return err
	}

	portArgs := make([]interface{}, len(port.CustomPorts)+1)
	portArgs[0] = portKeyPorts
	for i := 1; i < len(portArgs); i++ {
		portArgs[i] = port.CustomPorts[i-1]
	}

	if err := conn.Send("LPUSH", portArgs...); err != nil {
		return err
	}

	// put web config
	web := group.ModuleConfigurations.WebModule
	if err := conn.Send("HMSET", webKey, "extract_js", web.ExtractJS, "fingerprint_frameworks", web.FingerprintFrameworks, "max_links", web.MaxLinks, "take_screenshots", web.TakeScreenShots); err != nil {
		return err
	}

	_, err = conn.Do("EXEC")
	return err
}

// GroupStatus returns the status of this group in redis (exists, status, and last modified time)
func (s *State) GroupStatus(ctx context.Context, userContext am.UserContext, scanGroupID int) (bool, am.GroupStatus, int64, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return false, 0, 0, err
	}
	defer s.rc.Return(conn)

	configKey := fmt.Sprintf(ConfigFmt, userContext.GetOrgID(), scanGroupID)
	r, err := redis.StringMap(conn.Do("HGETALL", configKey))
	if err != nil {
		return false, 0, 0, err
	}

	if len(r) == 0 {
		return false, 0, 0, nil
	}

	status, err := strconv.ParseInt(r["status"], 10, 32)
	if err != nil {
		return false, 0, 0, err
	}

	modifiedTime, err := strconv.ParseInt(r["modified_time"], 10, 64)
	if err != nil {
		return false, 0, 0, err
	}
	return true, am.GroupStatus(status), modifiedTime, nil
}

// Delete all keys for this scan group
// TODO replace keys with scan
func (s *State) Delete(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) error {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return err
	}
	defer s.rc.Return(conn)
	// delete configuration
	key := fmt.Sprintf(ConfigFmt, group.OrgID, group.GroupID)

	r, err := redis.Strings(conn.Do("KEYS", key+"*"))
	if err != nil {
		return err
	}
	if err := conn.Send("MULTI"); err != nil {
		return err
	}

	for _, key := range r {
		if err := conn.Send("DEL", key); err != nil {
			return err
		}
	}

	_, err = conn.Do("EXEC")
	return err
}

func (s *State) PushAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses []*am.ScanGroupAddress) error {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return err
	}
	defer s.rc.Return(conn)
	/*
		key := fmt.Sprintf(AddrFmt, userContext.GetOrgID(), scanGroupID)

		if err := conn.Send("MULTI"); err != nil {
			return err
		}

		for _, addr := range addresses {
			conn.Send("HMSET")
		}
	*/
	return nil
}
