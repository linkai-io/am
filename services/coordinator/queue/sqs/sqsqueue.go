package sqs

import (
	"context"
	"strconv"

	"github.com/linkai-io/am/am"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awssqs "github.com/aws/aws-sdk-go/service/sqs"
)

type SQSQueue struct {
	Environment  string
	Region       string
	QueueTimeout int
	service      *awssqs.SQS
	session      *session.Session
}

// New sqsq queue for the specified env in region with a long polling time out
func New(env string, region string, timeout int) *SQSQueue {
	return &SQSQueue{Environment: env, Region: region, QueueTimeout: timeout}
}

// Init the sqs queue service
func (q *SQSQueue) Init() error {
	var err error
	q.session, err = session.NewSession(&aws.Config{
		Region: aws.String(q.Region)},
	)
	if q.Environment == "local" {
		q.session, err = session.NewSession(&aws.Config{
			LogLevel: aws.LogLevel(aws.LogDebugWithHTTPBody),
			Region:   aws.String(q.Region),
			Endpoint: aws.String("http://localhost:4576")},
		)
	}

	if err != nil {
		return err
	}
	// Create a SQS service client.
	q.service = awssqs.New(q.session)
	return nil
}

// List all known sqs queues
func (q *SQSQueue) List() (map[string]string, error) {
	return nil, nil
}

// Create the SQS queue for long polling
func (q *SQSQueue) Create(name string) (string, error) {
	out, err := q.service.CreateQueue(&awssqs.CreateQueueInput{
		QueueName: aws.String(q.Environment + "_" + name),
		Attributes: aws.StringMap(map[string]string{
			"ReceiveMessageWaitTimeSeconds": strconv.Itoa(q.QueueTimeout),
		}),
	})

	if err != nil {
		return "", err
	}

	return *out.QueueUrl, err
}

func (q *SQSQueue) Pause(queue string) error {
	return nil
}

func (q *SQSQueue) Delete(queue string) error {
	_, err := q.service.DeleteQueue(&awssqs.DeleteQueueInput{
		QueueUrl: aws.String(queue),
	})
	return err
}

func (q *SQSQueue) Stats(queue string) error {
	return nil
}

func (q *SQSQueue) PushAddresses(ctx context.Context, addresses []*am.ScanGroupAddress) error {
	/*batch := &awssqs.SendMessageBatchInput{}
	for _, addr := range addresses {
		entry := &awssqs.SendMessageBatchRequestEntry{}
	}
	q.service.SendMessageBatch()
	*/
	return nil
}
