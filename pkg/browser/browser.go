package browser

import (
	"context"

	"github.com/linkai-io/am/am"
)

type Browser interface {
	LoadForDiff(ctx context.Context, address *am.ScanGroupAddress, scheme, port string) (url, dom string, err error)
	// Load a web page, return the dom string, responses
	Load(ctx context.Context, address *am.ScanGroupAddress, scheme, port string) (webData *am.WebData, err error)
}
