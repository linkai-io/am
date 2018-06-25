package am

// Shade represents the testability of a host
type Shade int

const (
	// RED not to be tested/brute forced
	RED Shade = iota
	// YELLOW to be tested, not brute forced
	YELLOW
	// GREEN fully test
	GREEN
)
