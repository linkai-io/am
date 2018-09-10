package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/linkai-io/am/am"
	"github.com/spaolacci/murmur3"

	"github.com/linkai-io/am/pkg/redisclient"
)

const (
	// OrgID:GroupID
	ConfigFmt        = "%d:%d:configuration"
	AddrFmt          = "%d:%d:address:" // orgid:groupid:address:md5(ip,host)
	QueuesFmt        = ":queues:"
	NSConfigFmt      = ":module:nsconfig"
	BruteConfigFmt   = ":module:dnsbruteconfig"
	PortConfigFmt    = ":module:portconfig"
	WebConfigFmt     = ":module:webconfig"
	KeywordConfigFmt = ":module:keyword"
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

	keys := redisclient.NewRedisKeys(userContext.GetOrgID(), scanGroupID)
	_, err = conn.Do("HSET", keys.Status(), "status", am.GroupStarted)
	return err
}

func (s *State) Stop(ctx context.Context, userContext am.UserContext, scanGroupID int) error {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return err
	}
	defer s.rc.Return(conn)

	keys := redisclient.NewRedisKeys(userContext.GetOrgID(), scanGroupID)
	_, err = conn.Do("HSET", keys.Status(), "status", am.GroupStopped)
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
	keys := redisclient.NewRedisKeys(group.OrgID, group.GroupID)

	// create primary configuration
	if err := conn.Send("HMSET", redis.Args{keys.Config()}.AddFlat(group)...); err != nil {
		return err
	}

	// set scan group status to stopped (until addresses are added)
	if err := conn.Send("SET", keys.Status(), am.GroupStopped); err != nil {
		return err
	}

	// put ns config
	ns := group.ModuleConfigurations.NSModule
	if err := conn.Send("HMSET", redis.Args{keys.NSConfig()}.AddFlat(ns)...); err != nil {
		return err
	}

	// put dns brute config
	brute := group.ModuleConfigurations.BruteModule
	if err := conn.Send("HMSET", redis.Args{keys.BruteConfig()}.AddFlat(brute)...); err != nil {
		return err
	}

	// put dns custom subdomain names
	args := make([]interface{}, len(brute.CustomSubNames)+1)
	args[0] = keys.BruteConfigHosts()
	for i := 1; i < len(args); i++ {
		args[i] = brute.CustomSubNames[i-1]
	}

	if err := conn.Send("LPUSH", args...); err != nil {
		return err
	}

	// put port config
	port := group.ModuleConfigurations.PortModule
	if err := conn.Send("HMSET", redis.Args{keys.PortConfig()}.AddFlat(port)...); err != nil {
		return err
	}

	// put port custom ports
	portArgs := make([]interface{}, len(port.CustomPorts)+1)
	portArgs[0] = keys.PortConfigPorts()
	for i := 1; i < len(portArgs); i++ {
		portArgs[i] = port.CustomPorts[i-1]
	}

	if err := conn.Send("LPUSH", portArgs...); err != nil {
		return err
	}

	// put web config
	web := group.ModuleConfigurations.WebModule
	if err := conn.Send("HMSET", redis.Args{keys.WebConfig()}.AddFlat(web)...); err != nil {
		return err
	}

	// NOTE: we don't store the keyword module because it is empty, just the keywords (as of 2018/9/6)
	keyword := group.ModuleConfigurations.KeywordModule
	keywordArgs := make([]interface{}, len(keyword.Keywords)+1)
	keywordArgs[0] = keys.KeywordConfig()
	for i := 1; i < len(keywordArgs); i++ {
		keywordArgs[i] = keyword.Keywords[i-1]
	}

	if err := conn.Send("LPUSH", keywordArgs...); err != nil {
		return err
	}

	_, err = conn.Do("EXEC")
	return err
}

// GetGroup returns the entire scan group details.
func (s *State) GetGroup(ctx context.Context, userContext am.UserContext, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer s.rc.Return(conn)
	keys := redisclient.NewRedisKeys(userContext.GetOrgID(), scanGroupID)
	group := &am.ScanGroup{}

	value, err := redis.Values(conn.Do("HGETALL", keys.Config()))
	if err != nil {
		return nil, err
	}

	if err := redis.ScanStruct(value, group); err != nil {
		return nil, err
	}

	if wantModules {
		modules, err := s.getModules(keys, conn)
		if err != nil {
			return nil, err
		}
		group.ModuleConfigurations = modules
	}

	return group, nil
}

func (s *State) getModules(keys *redisclient.RedisKeys, conn redis.Conn) (*am.ModuleConfiguration, error) {
	ns := &am.NSModuleConfig{}
	brute := &am.BruteModuleConfig{}
	port := &am.PortModuleConfig{}
	web := &am.WebModuleConfig{}
	keyword := &am.KeywordModuleConfig{}

	// NS Module
	value, err := redis.Values(conn.Do("HGETALL", keys.NSConfig()))
	if err != nil {
		return nil, err
	}

	if err := redis.ScanStruct(value, ns); err != nil {
		return nil, err
	}

	// Brute Module
	value, err = redis.Values(conn.Do("HGETALL", keys.BruteConfig()))
	if err != nil {
		return nil, err
	}

	if err := redis.ScanStruct(value, brute); err != nil {
		return nil, err
	}

	hosts, err := redis.Strings(conn.Do("LRANGE", keys.BruteConfigHosts(), 0, -1))
	if err != nil {
		return nil, err
	}
	brute.CustomSubNames = hosts

	// Port Module
	value, err = redis.Values(conn.Do("HGETALL", keys.PortConfig()))
	if err != nil {
		return nil, err
	}

	if err := redis.ScanStruct(value, port); err != nil {
		return nil, err
	}

	ports, err := redis.Ints(conn.Do("LRANGE", keys.PortConfigPorts(), 0, -1))
	if err != nil {
		return nil, err
	}

	port.CustomPorts = make([]int32, len(ports))
	for i := 0; i < len(ports); i++ {
		port.CustomPorts[i] = int32(ports[i])
	}

	// Web Module
	value, err = redis.Values(conn.Do("HGETALL", keys.WebConfig()))
	if err != nil {
		return nil, err
	}

	if err := redis.ScanStruct(value, web); err != nil {
		return nil, err
	}

	// Keyword Module (just keywords)
	keywords, err := redis.Strings(conn.Do("LRANGE", keys.KeywordConfig(), 0, -1))
	if err != nil {
		return nil, err
	}
	keyword.Keywords = keywords

	return &am.ModuleConfiguration{
		NSModule:      ns,
		BruteModule:   brute,
		PortModule:    port,
		WebModule:     web,
		KeywordModule: keyword,
	}, nil
}

// GroupStatus returns the status of this group in redis (exists, status)
func (s *State) GroupStatus(ctx context.Context, userContext am.UserContext, scanGroupID int) (bool, am.GroupStatus, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return false, am.GroupStopped, err
	}
	defer s.rc.Return(conn)
	keys := redisclient.NewRedisKeys(userContext.GetOrgID(), scanGroupID)

	value, err := redis.Int(conn.Do("GET", keys.Status()))
	if err != nil {
		if err == redis.ErrNil {
			return false, am.GroupStopped, nil
		}
		return false, am.GroupStopped, err
	}
	return true, am.GroupStatus(value), nil
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
	key := fmt.Sprintf("%d:%d", group.OrgID, group.GroupID)

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

// PutAddresses iterates over the addresses and puts each to it's own key oid:gid:address:<addrid> <hash>
// also stores a set as an 'index' for retrieving them.
func (s *State) PutAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses []*am.ScanGroupAddress) error {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return err
	}
	defer s.rc.Return(conn)

	keys := redisclient.NewRedisKeys(userContext.GetOrgID(), scanGroupID)

	if err := conn.Send("MULTI"); err != nil {
		return err
	}

	for _, addr := range addresses {

		if err := conn.Send("SADD", keys.AddrList(), addr.AddressID); err != nil {
			return err
		}

		if err := conn.Send("SADD", keys.AddrHash(), HashAddress(addr.IPAddress, addr.HostAddress)); err != nil {
			return err
		}

		if err := conn.Send("HMSET", redis.Args{keys.Addr(addr.AddressID)}.AddFlat(addr)...); err != nil {
			return err
		}
	}
	_, err = conn.Do("EXEC")
	return err
}

// GetAddresses is a rather convoluted process of extracting addresses from redis. We use a cursor to iterate over
// a predefined set which contains all of the address id's. These address ids are then appended to the address key
// to be able to extract the address data using HGETALL. Since this is done in a MULTI pipeline, we are returned
// a slice of addresses and we need to iterate through each and do some interface type checking and finally call
// ScanStruct to map it back to an am.ScanGroupAddress. We use a map in the event that multiple addresses come back
// during the same iteration (which apparently can happen)
func (s *State) GetAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[int64]*am.ScanGroupAddress, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer s.rc.Return(conn)

	cachedAddrs := make(map[int64]*am.ScanGroupAddress, 0)
	keys := redisclient.NewRedisKeys(userContext.GetOrgID(), scanGroupID)
	iter := 0

	for {

		resp, err := redis.Values(conn.Do("SSCAN", keys.AddrList(), iter, "COUNT", limit))
		if err != nil {
			return nil, err
		}
		addressKeys, err := redis.Int64s(resp[1], err)
		if err != nil {
			return nil, err
		}

		iter, err = redis.Int(resp[0], err)
		if err != nil {
			return nil, err
		}

		if err := conn.Send("MULTI"); err != nil {
			return nil, err
		}

		for _, addressID := range addressKeys {
			if err := conn.Send("HGETALL", keys.Addr(addressID)); err != nil {
				return nil, err
			}
		}
		addrs, err := redis.Values(conn.Do("EXEC"))
		if err != nil {
			return nil, err
		}

		for _, addrData := range addrs {
			if a, ok := addrData.([]interface{}); ok {
				addr := &am.ScanGroupAddress{}
				if err := redis.ScanStruct(a, addr); err != nil {
					return nil, err
				}
				cachedAddrs[addr.AddressID] = addr
			}
		}
		if iter == 0 {
			break
		}
	}

	return cachedAddrs, nil
}

func (s *State) Exists(ctx context.Context, orgID, scanGroupID int, host, ipAddress string) (bool, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return false, err
	}
	defer s.rc.Return(conn)

	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
	return redis.Bool(conn.Do("SISMEMBER", keys.AddrHash(), HashAddress(host, ipAddress)))
}

// HashAddress for ip and host returning a hash key to allow modules to check if hosts exist
func HashAddress(ipAddress, host string) uint64 {
	hash := murmur3.New64()
	hash.Write([]byte(ipAddress))
	hash.Write([]byte(host))
	return hash.Sum64()
}
