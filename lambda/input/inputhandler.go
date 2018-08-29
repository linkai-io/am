package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/linkai-io/am/clients/address"
	"github.com/linkai-io/am/clients/scangroup"
	"github.com/linkai-io/am/pkg/convert"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	sgAddress       = os.Getenv("SG_SERVER")
	scanGroupClient = scangroup.New()
	addrAddress     = os.Getenv("ADDRESS_SERVER")
	addressClient   = address.New()
)

func init() {
	if err := scanGroupClient.Init([]byte(sgAddress)); err != nil {
		log.Fatalf("error initializing sg client: %s\n", err)
	}

	if err := addressClient.Init([]byte(addrAddress)); err != nil {
		log.Fatalf("error initializing addressclient: %s\n", err)
	}
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received: ", request.Body)
	fmt.Printf("REQUEST: %#v\n", request.PathParameters)
	_ = convert.APIToUserContext(&request)
	return events.APIGatewayProxyResponse{Body: request.Body + "DORK", StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
