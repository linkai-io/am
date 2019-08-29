package webhooks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

// AWSSender uses a pre-defined Lambda to handle sending webhooks
type AWSSender struct {
	Env          string
	Region       string
	functionName string
	sess         *session.Session
	svc          *lambda.Lambda
}

// NewAWSSender creates an env/region specific sender of webhook events
func NewAWSSender(env, region string) *AWSSender {
	s := &AWSSender{Env: env, Region: region}
	s.functionName = fmt.Sprintf("%s-function-event-webhooks", s.Env)
	s.sess = session.Must(session.NewSession(&aws.Config{Region: aws.String(s.Region)}))
	s.svc = lambda.New(s.sess)
	return s
}

// Init not necessary but other impl may need
func (s *AWSSender) Init() error {
	return nil
}

// Send the events using the env/region defined lambda function and get response
func (s *AWSSender) Send(ctx context.Context, evt *Data) (*DataResponse, error) {
	d, err := json.Marshal(evt)
	if err != nil {
		return nil, err
	}
	input := &lambda.InvokeInput{
		ClientContext:  nil,
		FunctionName:   aws.String(s.functionName),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
		LogType:        aws.String(lambda.LogTypeTail),
		Payload:        d,
		Qualifier:      nil,
	}

	resp, err := s.svc.InvokeWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	dataResponse := &DataResponse{}
	if err := json.Unmarshal(resp.Payload, dataResponse); err != nil {
		return nil, err
	}

	return dataResponse, nil
}
