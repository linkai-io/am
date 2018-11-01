package module

import "github.com/linkai-io/am/am"

func DefaultNSConfig() *am.NSModuleConfig {
	return &am.NSModuleConfig{
		RequestsPerSecond: 10,
	}
}

func DefaultWebConfig() *am.WebModuleConfig {
	return &am.WebModuleConfig{
		TakeScreenShots:       true,
		RequestsPerSecond:     50,
		MaxLinks:              1,
		ExtractJS:             true,
		FingerprintFrameworks: true,
	}
}

func DefaultPortConfig() *am.PortModuleConfig {
	return &am.PortModuleConfig{
		RequestsPerSecond: 50,
		CustomPorts:       []int32{80, 443},
	}
}

func DefaultBruteConfig() *am.BruteModuleConfig {
	return &am.BruteModuleConfig{
		MaxDepth:          2,
		RequestsPerSecond: 50,
		CustomSubNames:    make([]string, 0),
	}
}
