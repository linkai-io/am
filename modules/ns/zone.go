package ns

import "gopkg.linkai.io/v1/repos/am/am"

type ZoneRecords struct {
	MX   []*am.Host
	NS   []*am.Host
	A    []*am.Host
	AAAA []*am.Host
}
