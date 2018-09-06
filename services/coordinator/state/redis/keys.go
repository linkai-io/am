package redis

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
	r.addrFmt = fmt.Sprintf("%d:%d:address:", orgID, groupID)
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
	return r.configFmt + ":module:nsconfig"
}

func (r *RedisKeys) BruteConfig() string {
	return r.configFmt + ":module:dnsbruteconfig"
}

func (r *RedisKeys) BruteConfigHosts() string {
	return r.BruteConfig() + ":custom_hosts"
}

func (r *RedisKeys) PortConfig() string {
	return r.configFmt + ":module:portconfig"
}

func (r *RedisKeys) PortConfigPorts() string {
	return r.PortConfig() + ":custom_ports"
}

func (r *RedisKeys) WebConfig() string {
	return r.configFmt + ":module:webconfig"
}

func (r *RedisKeys) KeywordConfig() string {
	return r.configFmt + ":module:keyword"
}
