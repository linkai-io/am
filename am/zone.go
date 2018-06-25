package am

// Zone represents a domain zone, can be a subdomain
// If it has an SOA, it's a zone.
type Zone struct {
	Host       *Host
	NS         []*Host
	MX         []*Host
	A          []*Host
	AAAA       []*Host
	CNAME      []*Host
	IsWildcard bool
}
