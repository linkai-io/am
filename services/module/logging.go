package module

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/am"
)

func DefaultLogger(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) context.Context {
	l := log.With().
		Int("OrgID", userContext.GetOrgID()).
		Str("OrgCID", userContext.GetOrgCID()).
		Int("UserID", userContext.GetUserID()).
		Str("TraceID", userContext.GetTraceID()).
		Str("IPAddress", address.IPAddress).
		Str("HostAddress", address.HostAddress).
		Int64("AddressID", address.AddressID).
		Str("AddressHash", address.AddressHash).Logger()
	return l.WithContext(ctx)

}
