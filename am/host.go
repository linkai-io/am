package am

// Host represents host names and ip addresses and associated data
type Host struct {
	Names        []string
	IPs          []string
	Testability  Shade
	RecordType   int16
	FoundBy      []int
	FoundTime    int64
	LastSeenTime int64
}
