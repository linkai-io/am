package convert

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/linkai-io/am/am"
)

func APIToUserContext(in *events.APIGatewayProxyRequest) *am.UserContextData {

	return &am.UserContextData{
		TraceID: in.RequestContext.RequestID,
	}
}
