package main

import (
	"fmt"

	"gopkg.linkai.io/v1/repos/am/clients/address"
	"gopkg.linkai.io/v1/repos/am/clients/scangroup"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var addressClient *address.Client
var scanGroupClient *scangroup.Client

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	return events.APIGatewayProxyResponse{Body: request.Body + "DORK", StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
