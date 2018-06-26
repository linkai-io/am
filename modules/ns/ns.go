package ns

import (
	"encoding/json"
	"errors"

	"gopkg.linkai.io/v1/repos/am/am"
	"gopkg.linkai.io/v1/repos/am/modules/ns/state"
	"gopkg.linkai.io/v1/repos/libraries/dnsclient"
)

var (
	// ErrEmptyDNSServer missing dns server
	ErrEmptyDNSServer = errors.New("dns_server was empty or invalid")
)

// Config represents this modules configuration data to be passed in on
// initialization.
type Config struct {
	OrgID     string `json:"org_id"`
	DNSServer string `json:"dns_server"`
}

// NS module for extracting NS related information for an input list.
type NS struct {
	st     state.Stater
	config *Config
	dc     *dnsclient.Client
}

// New creates a new NS module for identifying zone information via DNS
// and storing the results in Redis.
func New(st state.Stater) *NS {
	return &NS{st: st}
}

// Init the redisclient and dns client.
func (ns *NS) Init(config []byte) error {
	var err error

	if ns.config, err = ns.parseConfig(config); err != nil {
		return err
	}

	ns.dc = dnsclient.New(ns.config.DNSServer, 3)
	return nil
}

// Name returns the module name
func (ns *NS) Name() string {
	return "NS"
}

// Analyze a domain zone, extracts NS, MX, A, AAAA, CNAME records
func (ns *NS) Analyze(zone string) *am.Zone {
	if !ns.st.IsNew(zone) {
		return nil
	}
	return nil
}

// parseConfig parses the configuration options and validates they are sane.
func (ns *NS) parseConfig(config []byte) (*Config, error) {
	var v *Config
	if err := json.Unmarshal(config, v); err != nil {
		return nil, err
	}

	if v.DNSServer == "" {
		return nil, ErrEmptyDNSServer
	}

	return v, nil
}
