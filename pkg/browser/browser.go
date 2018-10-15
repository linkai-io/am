package browser

import (
	"context"

	"github.com/linkai-io/am/am"
)

type Browser interface {
	// Load a web page, return the dom string, responses
	Load(ctx context.Context, address *am.ScanGroupAddress, scheme, port string) (dom string, responses []*am.HTTPResponse, err error)
}
