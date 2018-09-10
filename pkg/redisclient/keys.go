package redisclient

import "fmt"

type RedisKeys struct {
	orgID     int
	groupID   int
	configFmt string
	statusFmt string
	addrFmt   string
	queueFmt  string
}

func NewRedisKeys(orgID, groupID int) *RedisKeys {
	r := &RedisKeys{orgID: orgID, groupID: groupID}
	r.configFmt = fmt.Sprintf("%d:%d:configuration", orgID, groupID)
	r.addrFmt = fmt.Sprintf("%d:%d:address", orgID, groupID)
	r.statusFmt = fmt.Sprintf("%d:%d:status", orgID, groupID)
	r.queueFmt = fmt.Sprintf("%d:%d:queues", orgID, groupID)
	return r
}

func (r *RedisKeys) Config() string {
	return r.configFmt
}

func (r *RedisKeys) Status() string {
	return r.statusFmt
}

func (r *RedisKeys) Queues() string {
	return r.queueFmt
}

func (r *RedisKeys) NSConfig() string {
	return r.configFmt + ":module:ns:config"
}

func (r *RedisKeys) NSZones() string {
	return r.configFmt + ":module:ns:zones"
}

func (r *RedisKeys) NSZone(zone string) string {
	return r.configFmt + ":module:ns:zones:" + zone
}

func (r *RedisKeys) NSServers() string {
	return r.configFmt + ":module:ns:servers"
}

func (r *RedisKeys) BruteConfig() string {
	return r.configFmt + ":module:dnsbrute:config"
}

func (r *RedisKeys) BruteConfigHosts() string {
	return r.BruteConfig() + ":custom_hosts"
}

func (r *RedisKeys) PortConfig() string {
	return r.configFmt + ":module:port:config"
}

func (r *RedisKeys) PortConfigPorts() string {
	return r.PortConfig() + ":custom_ports"
}

func (r *RedisKeys) WebConfig() string {
	return r.configFmt + ":module:web:config"
}

func (r *RedisKeys) KeywordConfig() string {
	return r.configFmt + ":module:keyword"
}

func (r *RedisKeys) AddrList() string {
	return r.addrFmt + "_list"
}

func (r *RedisKeys) AddrHash() string {
	return r.addrFmt + "_hash"
}

// Addr returns the address key based on supplied addr id
// TODO: look at better more performant options
func (r *RedisKeys) Addr(addrID int64) string {
	return fmt.Sprintf("%d:%d:address:%d", r.orgID, r.groupID, addrID)
}

func (r *RedisKeys) AddrMatch() string {
	return r.addrFmt
}
