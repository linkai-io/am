package am

import (
	"context"
	"strings"
)

type ModuleType int

const (
	NSModule                 ModuleType = 1
	BruteModule              ModuleType = 2
	PortScanModule           ModuleType = 3
	WebModule                ModuleType = 4
	KeywordModule            ModuleType = 5
	BigDataCTSubdomainModule ModuleType = 6

	NSModuleServiceKey       = "nsmoduleservice"
	BruteModuleServiceKey    = "brutemoduleservice"
	PortScanModuleServiceKey = "portscanmoduleservice"
	WebModuleServiceKey      = "webmoduleservice"
	KeywordModuleServiceKey  = "keywordmoduleservice"
	BigDataModuleServiceKey  = "bigdatamoduleservice"
)

func KeyFromModuleType(moduleType ModuleType) string {
	switch moduleType {
	case NSModule:
		return NSModuleServiceKey
	case BruteModule:
		return BruteModuleServiceKey
	case PortScanModule:
		return PortScanModuleServiceKey
	case WebModule:
		return WebModuleServiceKey
	case KeywordModule:
		return KeywordModuleServiceKey
	case BigDataCTSubdomainModule:
		return BigDataModuleServiceKey
	}
	return ""
}

// ModuleConfiguration contains all the module configurations
type ModuleConfiguration struct {
	NSModule      *NSModuleConfig       `json:"ns_module"`
	BruteModule   *BruteModuleConfig    `json:"dnsbrute_module"`
	PortModule    *PortScanModuleConfig `json:"port_module"`
	WebModule     *WebModuleConfig      `json:"web_module"`
	KeywordModule *KeywordModuleConfig  `json:"keyword_module"`
}

// Module represents a module of work such as brute force, web scrape etc.
type Module interface {
	Name() string
	Config() map[string]interface{}
}

// ModuleStats contains a
type ModuleStats struct {
	Running   int64
	WorkCount int64
	Remaining int64
}

// NSModuleConfig for NS module
type NSModuleConfig struct {
	RequestsPerSecond int32 `json:"requests_per_second"`
}

// BruteModuleConfig DNS subdomain brute forcer
type BruteModuleConfig struct {
	CustomSubNames    []string `json:"custom_subnames" redis:"-"`
	RequestsPerSecond int32    `json:"requests_per_second"`
	MaxDepth          int32    `json:"max_depth"`
}

// PortModuleConfig for simple port scanning module
type PortScanModuleConfig struct {
	RequestsPerSecond int32    `json:"requests_per_second"`
	PortScanEnabled   bool     `json:"port_scan_enabled"`
	CustomWebPorts    []int32  `json:"custom_ports" redis:"-"`
	TCPPorts          []int32  `json:"tcp_ports" redis:"-"`
	UDPPorts          []int32  `json:"udp_ports" redis:"-"`
	AllowedTLDs       []string `json:"allowed_tlds" redis:"-"`
	AllowedHosts      []string `json:"allowed_hosts" redis:"-"`
	DisallowedTLDs    []string `json:"disallowed_tlds" redis:"-"`
	DisallowedHosts   []string `json:"disallowed_hosts" redis:"-"`
}

// CanPortScan takes the etld and host and determines if this host is allowed to be port scanned
// first check that it's enabled
// then check that the host is not in the disallowed list (return false if it is)
// then check that the host is in our allowed hosts (overrides TLD check) return true if it is
// then check taht the host is in our disallowed TLDs (return false if it is)
// finally check that the host is in our allowed TLDs (return true if it is)
// other wise return false
func (c *PortScanModuleConfig) CanPortScan(etld, host string) bool {
	if c.PortScanEnabled == false {
		return false
	}

	if host == "" {
		return false
	}

	if c.inSlice(host, c.DisallowedHosts) {
		return false
	}

	if c.inSlice(host, c.AllowedHosts) {
		return true
	}

	if etld == "" {
		return false
	}

	if c.inSlice(etld, c.DisallowedTLDs) {
		return false
	}

	if c.inSlice(etld, c.AllowedTLDs) {
		return true
	}
	return false
}

// CanPortScanIP is similar to above, but for IP addresses (no ETLD checks)
// also we fail 'open' assuming if it's not in disallowed *or* allowed, then we are allowed to scan it.
func (c *PortScanModuleConfig) CanPortScanIP(ip string) bool {
	if c.PortScanEnabled == false {
		return false
	}

	if ip == "" {
		return false
	}

	if c.inSlice(ip, c.DisallowedHosts) {
		return false
	}

	return true
}

func (c *PortScanModuleConfig) inSlice(needle string, haystack []string) bool {
	needle = strings.ToLower(needle)
	for _, element := range haystack {
		if strings.ToLower(element) == needle {
			return true
		}
	}
	return false
}

// WebModuleConfig for web related analysis module
type WebModuleConfig struct {
	TakeScreenShots       bool  `json:"take_screenshots"`
	RequestsPerSecond     int32 `json:"requests_per_second"`
	MaxLinks              int32 `json:"max_links"`
	ExtractJS             bool  `json:"extract_js"`
	FingerprintFrameworks bool  `json:"fingerprint_frameworks"`
}

type KeywordModuleConfig struct {
	Keywords []string `json:"keywords" redis:"-"`
}

// ModuleService is the default interface for analyzing an address and spitting out potentially
// more addresses
type ModuleService interface {
	Analyze(ctx context.Context, userContext UserContext, address *ScanGroupAddress) (*ScanGroupAddress, map[string]*ScanGroupAddress, error)
}

// PortModuleService is for modules which react/analyze open ports
type PortModuleService interface {
	AnalyzeWithPorts(ctx context.Context, userContext UserContext, address *ScanGroupAddress, ports *PortResults) (*ScanGroupAddress, map[string]*ScanGroupAddress, *Bag, error)
}
