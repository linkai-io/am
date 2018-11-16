package am

import (
	"context"
)

type ModuleType int

const (
	NSModule      ModuleType = 1
	BruteModule   ModuleType = 2
	PortModule    ModuleType = 3
	WebModule     ModuleType = 4
	KeywordModule ModuleType = 5

	NSModuleServiceKey      = "nsmoduleservice"
	BruteModuleServiceKey   = "brutemoduleservice"
	PortModuleServiceKey    = "portmoduleservice"
	WebModuleServiceKey     = "webmoduleservice"
	KeywordModuleServiceKey = "keywordmoduleservice"
	BigDataModuleServiceKey = "bigdatamoduleservice"
)

func KeyFromModuleType(moduleType ModuleType) string {
	switch moduleType {
	case NSModule:
		return NSModuleServiceKey
	case BruteModule:
		return BruteModuleServiceKey
	case PortModule:
		return PortModuleServiceKey
	case WebModule:
		return WebModuleServiceKey
	case KeywordModule:
		return KeywordModuleServiceKey
	}
	return ""
}

// ModuleConfiguration contains all the module configurations
type ModuleConfiguration struct {
	NSModule      *NSModuleConfig      `json:"ns_module"`
	BruteModule   *BruteModuleConfig   `json:"dnsbrute_module"`
	PortModule    *PortModuleConfig    `json:"port_module"`
	WebModule     *WebModuleConfig     `json:"web_module"`
	KeywordModule *KeywordModuleConfig `json:"keyword_module"`
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
type PortModuleConfig struct {
	RequestsPerSecond int32   `json:"requests_per_second"`
	CustomPorts       []int32 `json:"custom_ports" redis:"-"`
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
