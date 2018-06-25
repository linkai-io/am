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
