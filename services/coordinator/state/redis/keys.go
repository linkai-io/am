package redis

import "fmt"

type RedisKeys struct {
	OrgID     int
	GroupID   int
	ConfigFmt string
	AddrFmt   string
}

func NewRedisKeys(orgID, groupID int) *RedisKeys {
	r := &RedisKeys{OrgID: orgID, GroupID: groupID}
	r.ConfigFmt = fmt.Sprintf("%d:%d:configuration", orgID, groupID)
	r.AddrFmt = fmt.Sprintf("%d:%d:address:", orgID, groupID)
	return r
}

func (r *RedisKeys) NSConfig() string {
	return r.ConfigFmt + ":module:nsconfig"
}

func (r *RedisKeys) BruteConfig() string {
	return r.ConfigFmt + ":module:dnsbruteconfig"
}

func (r *RedisKeys) BruteConfigHosts() string {
	return r.BruteConfig() + ":custom_hosts"
}

func (r *RedisKeys) PortConfig() string {
	return r.ConfigFmt + ":module:portconfig"
}

func (r *RedisKeys) PortConfigPorts() string {
	return r.PortConfig() + ":custom_ports"
}

func (r *RedisKeys) WebConfig() string {
	return r.ConfigFmt + ":module:webconfig"
}

func (r *RedisKeys) KeywordConfig() string {
	return r.ConfigFmt + ":module:keyword"
}
