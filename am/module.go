package am

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
	Name string `json:"name"`
}

// BruteModuleConfig DNS subdomain brute forcer
type BruteModuleConfig struct {
	Name           string   `json:"name"`
	CustomSubNames []string `json:"custom_subnames"`
	MaxDepth       int32    `json:"max_depth"`
}

// PortModuleConfig for simple port scanning module
type PortModuleConfig struct {
	Name  string  `json:"name"`
	Ports []int32 `json:"ports"`
}

// WebModuleConfig for web related analysis module
type WebModuleConfig struct {
	Name                  string `json:"name"`
	TakeScreenShots       bool   `json:"take_screenshots"`
	MaxLinks              int32  `json:"max_links"`
	ExtractJS             bool   `json:"extract_js"`
	FingerprintFrameworks bool   `json:"fingerprint_frameworks"`
}
