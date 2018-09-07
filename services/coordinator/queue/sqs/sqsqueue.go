package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"sync"

	"github.com/linkai-io/am/am"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awssqs "github.com/aws/aws-sdk-go/service/sqs"
)

// BatchOutput struct
type BatchOutput struct {
	Output *awssqs.SendMessageBatchOutput
	Error  error
}

// SQSQueue manages the scangroup state, queues and address state
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
	out, err := q.service.ListQueues(&awssqs.ListQueuesInput{
		QueueNamePrefix: aws.String(""),
	})
	if err != nil {
		return nil, err
	}
	for _, url := range out.QueueUrls {
		log.Printf("%s\n", *url)
	}

	return nil, nil
}

// Create the SQS queue for long polling + create dead letter queue
func (q *SQSQueue) Create(name string) (string, error) {
	deadLetterQueue := name + "_dead"
	deadQueue, err := q.service.CreateQueue(&awssqs.CreateQueueInput{
		QueueName: aws.String(q.Environment + "_" + deadLetterQueue),
	})

	attr, err := q.service.GetQueueAttributes(&awssqs.GetQueueAttributesInput{
		AttributeNames: aws.StringSlice([]string{"QueueArn"}),
		QueueUrl:       deadQueue.QueueUrl,
	})

	policy := map[string]string{
		"maxReceiveCount":     "3",
		"deadLetterTargetArn": *attr.Attributes["QueueArn"],
	}

	b, err := json.Marshal(policy)
	if err != nil {
		fmt.Println("Failed to marshal policy:", err)
		return "", err
	}

	out, err := q.service.CreateQueue(&awssqs.CreateQueueInput{
		QueueName: aws.String(q.Environment + "_" + name),
		Attributes: aws.StringMap(map[string]string{
			"ReceiveMessageWaitTimeSeconds": strconv.Itoa(q.QueueTimeout),
			"RedrivePolicy":                 string(b),
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

func (q *SQSQueue) PushAddresses(ctx context.Context, queue string, addresses []*am.ScanGroupAddress) error {
	var wg sync.WaitGroup
	var batchOutput []*BatchOutput

	addrLen := len(addresses)
	maxlen := 10
	times := addrLen / maxlen
	more := addrLen % maxlen
	result := make(chan *BatchOutput)

	do := func(addrs []*am.ScanGroupAddress) {
		defer wg.Done()
		runtime.Gosched()
		var b = &BatchOutput{}
		b.Output, b.Error = q.sendBatch(queue, addrs)
		result <- b
	}

	wg.Add(times)

	// push remainder
	if more > 0 {
		wg.Add(1)
		go do(addresses[maxlen*times : maxlen*times+more])
	}

	// push batches
	for i := 0; i < times; i++ {
		go do(addresses[maxlen*i : maxlen*(i+1)])
	}

	// collect output
	batchOutput = make([]*BatchOutput, 0)
	go func() {
		for {
			select {
			case v, ok := <-result:
				if ok {
					batchOutput = append(batchOutput, v)
				}
			}
		}
	}()
	// wait for completion
	wg.Wait()

	for _, output := range batchOutput {
		if output.Error != nil {
			log.Printf("error batch processing: %s\n", output.Error.Error())
		}
	}
	return nil
}

// sendBatch to send batch messages.
func (q *SQSQueue) sendBatch(queue string, addresses []*am.ScanGroupAddress) (*awssqs.SendMessageBatchOutput, error) {
	var entries []*awssqs.SendMessageBatchRequestEntry
	entries = make([]*awssqs.SendMessageBatchRequestEntry, len(addresses))

	for i, body := range addresses {
		data, err := q.encodeForQueue(body)
		if err != nil {
			log.Printf("error encoding scan address %d %s\n", body.AddressID, err)
			i-- // decrement so we don't have holes in our entries.
			continue
		}

		entries[i] = &awssqs.SendMessageBatchRequestEntry{
			Id:          aws.String(string(body.AddressID)),
			MessageBody: aws.String(data),
		}
	}

	return q.service.SendMessageBatch(&awssqs.SendMessageBatchInput{
		Entries:  entries,
		QueueUrl: aws.String(queue),
	})
}

// encodedForQueue, for now use json, consider protobuf if faster
func (q *SQSQueue) encodeForQueue(address *am.ScanGroupAddress) (string, error) {
	data, err := json.Marshal(address)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
