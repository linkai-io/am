package dnsclient

import (
	"github.com/miekg/dns"
)

// resultError encapsulates Results and an error.
type resultError struct {
	Result *Results
	Error  error
}

// Results holds results of a lookup
type Results struct {
	IPs         []string // a slice of ip addresses returned from a request
	Hosts       []string // a slice of hosts returned from a request
	RecordType  uint16   // records returned
	RequestType uint16   // where these results came from (TypeA, TypeAAAA etc)
}

// Type returns a string representation of the record type
func (r *Results) Type() string {
	return dns.TypeToString[r.RecordType]
}

type axfrResultError struct {
	Result    []*Results
	NSAddress string
	Error     error
}
