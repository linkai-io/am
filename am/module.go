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
	Name string
}

// BruteModuleConfig DNS subdomain brute forcer
type BruteModuleConfig struct {
	Name           string
	CustomSubNames []string
	MaxDepth       int32
}

// PortModuleConfig for simple port scanning module
type PortModuleConfig struct {
	Name  string
	Ports []int32
}

// WebModuleConfig for web related analysis module
type WebModuleConfig struct {
	Name                  string
	TakeScreenShots       bool
	MaxLinks              int32
	ExtractJS             bool
	FingerprintFrameworks bool
}
