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
func (q *SQSQueue) Create(name string) error {
	_, err := q.service.CreateQueue(&awssqs.CreateQueueInput{
		QueueName: aws.String(q.Environment + "_" + name),
		Attributes: aws.StringMap(map[string]string{
			"ReceiveMessageWaitTimeSeconds": strconv.Itoa(q.QueueTimeout),
		}),
	})
	return err
}

func (q *SQSQueue) Pause(name string) error {
	return nil
}

func (q *SQSQueue) Delete(name string) error {
	return nil
}

func (q *SQSQueue) Stats(name string) error {
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
