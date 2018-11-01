package module

import (
	"github.com/linkai-io/am/am"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func DefaultLogger(userContext am.UserContext, address *am.ScanGroupAddress) zerolog.Logger {
	return log.With().
		Int("OrgID", userContext.GetOrgID()).
		Int("UserID", userContext.GetUserID()).
		Str("TraceID", userContext.GetTraceID()).
		Str("IPAddress", address.IPAddress).
		Str("HostAddress", address.HostAddress).
		Int64("AddressID", address.AddressID).
		Str("AddressHash", address.AddressHash).Logger()
}
