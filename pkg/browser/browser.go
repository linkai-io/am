package browser

import (
	"context"

	"github.com/linkai-io/am/am"
)

type Browser interface {
	Load(ctx context.Context, address *am.ScanGroupAddress, scheme, port string) (dom string, responses []*am.HTTPResponse, err error)
}
