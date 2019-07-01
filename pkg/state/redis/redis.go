package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	"github.com/linkai-io/am/am"
	"github.com/pkg/errors"

	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/redisclient"
	"github.com/linkai-io/am/pkg/state"
)

var (
	//ErrEmptyRCAddress missing redis address
	ErrEmptyRCAddress = errors.New("rc_addr was empty or invalid")
	// ErrEmptyRCPassword missing redis password
	ErrEmptyRCPassword = errors.New("rc_pass was empty or invalid")
)

// State manager
type State struct {
	rc *redisclient.Client
}

// New redis backed state
func New() *State {
	return &State{}
}

// Init by passing address and password
func (s *State) Init(addr, pass string) error {

	s.rc = redisclient.New(addr, pass)

	return s.rc.Init()
}

// Start set scan group status to started
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

// Stop set scan group status to stopped
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

// Put the scan group configuration and publish to the scan group RN that it has been put
// or updated
// TODO: PUT SCANGROUP IN SET
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
	if err := conn.Send("HSET", keys.Status(), "status", am.GroupStopped); err != nil {
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
	if len(brute.CustomSubNames) != 0 {
		args := make([]interface{}, len(brute.CustomSubNames)+1)
		args[0] = keys.BruteConfigHosts()
		for i := 1; i < len(args); i++ {
			args[i] = brute.CustomSubNames[i-1]
		}

		if err := conn.Send("LPUSH", args...); err != nil {
			return err
		}
	}

	// put port config
	port := group.ModuleConfigurations.PortModule
	if err := conn.Send("HMSET", redis.Args{keys.PortConfig()}.AddFlat(port)...); err != nil {
		return err
	}

	// put port custom ports
	if len(port.CustomPorts) != 0 {
		portArgs := make([]interface{}, len(port.CustomPorts)+1)
		portArgs[0] = keys.PortConfigPorts()
		for i := 1; i < len(portArgs); i++ {
			portArgs[i] = port.CustomPorts[i-1]
		}

		if err := conn.Send("LPUSH", portArgs...); err != nil {
			return err
		}
	}

	// put tcp ports
	if len(port.TCPPorts) != 0 {
		portArgs := make([]interface{}, len(port.TCPPorts)+1)
		portArgs[0] = keys.PortConfigTCPPorts()
		for i := 1; i < len(portArgs); i++ {
			portArgs[i] = port.TCPPorts[i-1]
		}

		if err := conn.Send("LPUSH", portArgs...); err != nil {
			return err
		}
	}

	// put udp ports
	if len(port.UDPPorts) != 0 {
		portArgs := make([]interface{}, len(port.UDPPorts)+1)
		portArgs[0] = keys.PortConfigUDPPorts()
		for i := 1; i < len(portArgs); i++ {
			portArgs[i] = port.UDPPorts[i-1]
		}

		if err := conn.Send("LPUSH", portArgs...); err != nil {
			return err
		}
	}

	// put allowed TLDs
	if len(port.AllowedTLDs) != 0 {
		args := make([]interface{}, len(port.AllowedTLDs)+1)
		args[0] = keys.PortConfigAllowedTLDs()
		for i := 1; i < len(args); i++ {
			args[i] = port.AllowedTLDs[i-1]
		}

		if err := conn.Send("LPUSH", args...); err != nil {
			return err
		}
	}

	// put allowed hosts
	if len(port.AllowedHosts) != 0 {
		args := make([]interface{}, len(port.AllowedHosts)+1)
		args[0] = keys.PortConfigAllowedHosts()
		for i := 1; i < len(args); i++ {
			args[i] = port.AllowedHosts[i-1]
		}

		if err := conn.Send("LPUSH", args...); err != nil {
			return err
		}
	}

	// put disallowed TLDs
	if len(port.DisallowedTLDs) != 0 {
		args := make([]interface{}, len(port.DisallowedTLDs)+1)
		args[0] = keys.PortConfigDisallowedTLDs()
		for i := 1; i < len(args); i++ {
			args[i] = port.DisallowedTLDs[i-1]
		}

		if err := conn.Send("LPUSH", args...); err != nil {
			return err
		}
	}

	// put disallowed hosts
	if len(port.DisallowedHosts) != 0 {
		args := make([]interface{}, len(port.DisallowedHosts)+1)
		args[0] = keys.PortConfigDisallowedHosts()
		for i := 1; i < len(args); i++ {
			args[i] = port.DisallowedHosts[i-1]
		}

		if err := conn.Send("LPUSH", args...); err != nil {
			return err
		}
	}

	// put web config
	web := group.ModuleConfigurations.WebModule
	if err := conn.Send("HMSET", redis.Args{keys.WebConfig()}.AddFlat(web)...); err != nil {
		return err
	}

	// NOTE: we don't store the keyword module because it is empty, just the keywords (as of 2018/9/6)
	keyword := group.ModuleConfigurations.KeywordModule
	if len(keyword.Keywords) != 0 {
		keywordArgs := make([]interface{}, len(keyword.Keywords)+1)
		keywordArgs[0] = keys.KeywordConfig()
		for i := 1; i < len(keywordArgs); i++ {
			keywordArgs[i] = keyword.Keywords[i-1]
		}

		if err := conn.Send("LPUSH", keywordArgs...); err != nil {
			return err
		}
	}

	if err := conn.Send("PUBLISH", am.RNScanGroupGroups, keys.Config()); err != nil {
		return err
	}

	_, err = conn.Do("EXEC")
	return err
}

// GetGroup returns the entire scan group details.
func (s *State) GetGroup(ctx context.Context, orgID, scanGroupID int, wantModules bool) (*am.ScanGroup, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer s.rc.Return(conn)
	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
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
	port := &am.PortScanModuleConfig{}
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

	tcpPorts, err := redis.Ints(conn.Do("LRANGE", keys.PortConfigTCPPorts(), 0, -1))
	if err != nil {
		return nil, err
	}

	port.TCPPorts = make([]int32, len(tcpPorts))
	for i := 0; i < len(tcpPorts); i++ {
		port.TCPPorts[i] = int32(tcpPorts[i])
	}

	udpPorts, err := redis.Ints(conn.Do("LRANGE", keys.PortConfigUDPPorts(), 0, -1))
	if err != nil {
		return nil, err
	}

	port.UDPPorts = make([]int32, len(udpPorts))
	for i := 0; i < len(udpPorts); i++ {
		port.UDPPorts[i] = int32(udpPorts[i])
	}

	allowedTLDs, err := redis.Strings(conn.Do("LRANGE", keys.PortConfigAllowedTLDs(), 0, -1))
	if err != nil {
		return nil, err
	}
	port.AllowedTLDs = allowedTLDs

	allowedHosts, err := redis.Strings(conn.Do("LRANGE", keys.PortConfigAllowedHosts(), 0, -1))
	if err != nil {
		return nil, err
	}
	port.AllowedHosts = allowedHosts

	disallowedTLDs, err := redis.Strings(conn.Do("LRANGE", keys.PortConfigDisallowedTLDs(), 0, -1))
	if err != nil {
		return nil, err
	}
	port.DisallowedTLDs = disallowedTLDs

	disallowedHosts, err := redis.Strings(conn.Do("LRANGE", keys.PortConfigDisallowedHosts(), 0, -1))
	if err != nil {
		return nil, err
	}
	port.DisallowedHosts = disallowedHosts

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

	value, err := redis.Int(conn.Do("HGET", keys.Status(), "status"))
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

// PutAddresses puts addresses that are in slice form into the work queue, exists set, and the address data
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
		if err := s.putAddress(conn, keys, addr); err != nil {
			return err
		}
	}
	_, err = conn.Do("EXEC")
	return err
}

// PutAddressMap puts addresses that are in map form into the work queue, exists set, and the address data
func (s *State) PutAddressMap(ctx context.Context, userContext am.UserContext, scanGroupID int, addresses map[string]*am.ScanGroupAddress) error {
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
		if err := s.putAddress(conn, keys, addr); err != nil {
			return err
		}
	}
	_, err = conn.Do("EXEC")
	return err
}

// putAddress adds a hash if it does not have one, adds it to work queue, the exist set and the address data.
func (s *State) putAddress(conn redis.Conn, keys *redisclient.RedisKeys, address *am.ScanGroupAddress) error {
	if address.AddressHash == "" {
		address.AddressHash = convert.HashAddress(address.IPAddress, address.HostAddress)
	}

	// addrworkqueue is used to pop addresses out of the set
	if err := conn.Send("SADD", keys.AddrWorkQueue(), address.AddressHash); err != nil {
		return err
	}

	// addrexistshash is used for testing if an address is stored in redis
	if err := conn.Send("SADD", keys.AddrExistsHash(), address.AddressHash); err != nil {
		return err
	}

	// the actual address data stored in a hashset
	if err := conn.Send("HMSET", redis.Args{keys.Addr(address.AddressHash)}.AddFlat(address)...); err != nil {
		return err
	}
	return nil
}

// PopAddresses pops the addresses hashes from the work queue key, uses that to call HGETALL to return the address data
func (s *State) PopAddresses(ctx context.Context, userContext am.UserContext, scanGroupID int, limit int) (map[string]*am.ScanGroupAddress, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer s.rc.Return(conn)

	cachedAddrs := make(map[string]*am.ScanGroupAddress, 0)
	keys := redisclient.NewRedisKeys(userContext.GetOrgID(), scanGroupID)

	resp, err := redis.Values(conn.Do("SPOP", keys.AddrWorkQueue(), limit))
	if err != nil {
		return nil, errors.Wrap(err, "failed to pop from work queue")
	}

	addressHashKeys, err := redis.Strings(resp, err)
	if err != nil {
		return nil, err
	}

	if err := conn.Send("MULTI"); err != nil {
		return nil, err
	}

	for _, addressHash := range addressHashKeys {
		if err := conn.Send("HGETALL", keys.Addr(addressHash)); err != nil {
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
			// check against an empty record, by ensuring the group ids
			// match, if empty, group id will be 0.
			if addr.GroupID == scanGroupID {
				cachedAddrs[addr.AddressHash] = addr
			}

		}
	}

	return cachedAddrs, nil
}

// Exists checks if a host/ipaddress pair is in our list of *known* addreses for this group
func (s *State) Exists(ctx context.Context, orgID, scanGroupID int, host, ipAddress string) (bool, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return false, err
	}
	defer s.rc.Return(conn)

	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
	return redis.Bool(conn.Do("SISMEMBER", keys.AddrExistsHash(), convert.HashAddress(host, ipAddress)))
}

// FilterNew returns only new addresses
func (s *State) FilterNew(ctx context.Context, orgID, scanGroupID int, addresses map[string]*am.ScanGroupAddress) (map[string]*am.ScanGroupAddress, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer s.rc.Return(conn)

	keys := redisclient.NewRedisKeys(orgID, scanGroupID)

	hashes := make([]interface{}, len(addresses)+1)
	tempID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	// key to store hashes at
	hashes[0] = tempID.String()

	// since we already need to loop over address hashes, use it to also
	// create a map we will use to filter out duplicates after we get the
	// response from SINTER
	i := 1
	for _, v := range addresses {
		hashes[i] = v.AddressHash
		i++
	}

	if err := conn.Send("MULTI"); err != nil {
		return nil, err
	}

	if err := conn.Send("SADD", hashes...); err != nil {
		return nil, err
	}

	if err := conn.Send("SINTER", keys.AddrExistsHash(), tempID.String()); err != nil {
		return nil, err
	}

	if err := conn.Send("DEL", tempID.String()); err != nil {
		return nil, err
	}

	val, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return nil, err
	}

	// take the 1th element which contains the results of SINTER
	exists, err := redis.Strings(val[1], nil)
	if err != nil {
		return nil, err
	}

	// remove dupelicates (already exist) from our map
	for _, dupe := range exists {
		if _, exist := addresses[dupe]; exist {
			delete(addresses, dupe)
		}
	}

	return addresses, nil
}

// Subscribe to listen for group state updates
func (s *State) Subscribe(ctx context.Context, onStartFn state.SubOnStart, onMessageFn state.SubOnMessage, channels ...string) error {
	return s.rc.Subscribe(ctx, onStartFn, onMessageFn, channels...)
}

// DoBruteETLD global method to rate limit how many ETLDs we will brute force concurrently.
func (s *State) DoBruteETLD(ctx context.Context, orgID, scanGroupID, expireSeconds int, maxAllowed int, etld string) (int, bool, error) {
	// create redis keys for this org/group
	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return 0, false, err
	}
	defer s.rc.Return(conn)

	// adaptation of https://redis.io/commands/incr (see bottom of the page for example)
	key := keys.BruteETLD(etld)
	count, err := redis.Int(conn.Do("LLEN", key))
	if err != nil {
		return 0, false, err
	}

	if count >= maxAllowed {
		return count, false, nil
	}

	exist, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return 0, false, nil
	}

	if exist {
		count, err = redis.Int(conn.Do("RPUSHX", key, key))
		return count, (err == nil), err
	}

	if err := conn.Send("MULTI"); err != nil {
		return 0, false, err
	}

	if err := conn.Send("RPUSH", key, key); err != nil {
		return 0, false, err
	}

	if err := conn.Send("EXPIRE", key, expireSeconds); err != nil {
		return 0, false, err
	}

	if _, err = redis.Values(conn.Do("EXEC")); err != nil {
		return 0, false, err
	}
	return count + 1, true, nil
}

// DoNSRecords org:group:module:ns:zone:<zonename> sets the zone as already being checked or, if it already exists
// return that we shouldn't do NS records for this zone.
func (s *State) DoNSRecords(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	// create redis keys for this org/group
	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
	return s.do(ctx, orgID, scanGroupID, expireSeconds, keys.NSZone(zone), zone)
}

// DoBruteDomain org:group:module:dnsbrute:zones:brute:<zonename> sets the zone as already being checked or, if it already exists
// return that we shouldn't do analysis.
func (s *State) DoBruteDomain(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	// create redis keys for this org/group
	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
	return s.do(ctx, orgID, scanGroupID, expireSeconds, keys.BruteZone(zone), zone)
}

// DoMutateDomain org:group:module:dnsbrute:zones:mutate:<zonename> sets the zone as already being checked or, if it already exists
// return that we shouldn't do analysis.
func (s *State) DoMutateDomain(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	// create redis keys for this org/group
	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
	return s.do(ctx, orgID, scanGroupID, expireSeconds, keys.MutateZone(zone), zone)
}

// DoWebDomain org:group:module:web:zones:analyze:<zonename> sets the zone as already being checked or, if it already exists
// return that we shouldn't do analysis.
func (s *State) DoWebDomain(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	// create redis keys for this org/group
	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
	return s.do(ctx, orgID, scanGroupID, expireSeconds, keys.WebZone(zone), zone)
}

// DoCTDomain org:group:":module:bigdata:zones:<zonename> sets the zone as already being checked or, if it already exists
// return that we shouldn't look up in bigdata.
func (s *State) DoCTDomain(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	// create redis keys for this org/group
	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
	return s.do(ctx, orgID, scanGroupID, expireSeconds, keys.BigDataZone(zone), zone)
}

// DoPortScan org:group:":module:port:zones:<zonename> sets the zone as already being checked or, if it already exists
// return that we shouldn't port scan this zone.
func (s *State) DoPortScan(ctx context.Context, orgID, scanGroupID int, expireSeconds int, zone string) (bool, error) {
	// create redis keys for this org/group
	keys := redisclient.NewRedisKeys(orgID, scanGroupID)
	return s.do(ctx, orgID, scanGroupID, expireSeconds, keys.PortZone(zone), zone)
}

// Sets and checks if a value exists in a key. If it already exists, we don't need to do whatever 'key's work is, as
// it's already been done.
func (s *State) do(ctx context.Context, orgID, scanGroupID int, expireSeconds int, key, zone string) (bool, error) {
	conn, err := s.rc.GetContext(ctx)
	if err != nil {
		return false, err
	}
	defer s.rc.Return(conn)

	ret, err := redis.String(conn.Do("SET", key, time.Now().UnixNano(), "NX", "EX", expireSeconds))
	if err != nil {
		// redis will return ErrNil if value is already set.
		if err == redis.ErrNil {
			return false, nil
		}
		return false, err
	}
	return ret == "OK", nil
}

// TestGetConn for testing
func (s *State) TestGetConn() redis.Conn {
	return s.rc.Get()
}
