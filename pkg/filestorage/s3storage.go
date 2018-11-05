package filestorage

import (
	"bytes"
	"context"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/pkg/retrier"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/linkai-io/am/am"
)

type S3Storage struct {
	region  string
	env     string
	session *session.Session
}

func NewS3Storage(env, region string) *S3Storage {
	return &S3Storage{region: region, env: env}
}

func (s *S3Storage) Init(config []byte) error {
	var err error
	s.session, err = session.NewSession(&aws.Config{Region: aws.String(s.region)})
	return err
}

func (s *S3Storage) Write(ctx context.Context, address *am.ScanGroupAddress, data []byte) (string, string, error) {
	hashName := convert.HashData(data)
	fileName := PathFromData(address, hashName)
	if fileName == "null" {
		return "", "", nil
	}
	headObjectInput := &s3.HeadObjectInput{}
	s3session := s3.New(s.session)

	link := ""
	out, err := s3session.HeadObject(headObjectInput)
	if err != nil {
		return hashName, link, s.uploadWithRetry(ctx, s3session, fileName, data)
	}
	// already exists don't bother uploading again
	if out != nil {
		return "", "", nil
	}
	return "", "", nil
}

func (s *S3Storage) uploadWithRetry(ctx context.Context, s3session *s3.S3, fileName string, data []byte) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(fileName),
		Body:   bytes.NewReader(data),
	}

	retryErr := retrier.Retry(func() error {
		_, err := s3session.PutObject(input)
		if err == nil {
			return nil
		}
		return err
	})
	return retryErr
}
