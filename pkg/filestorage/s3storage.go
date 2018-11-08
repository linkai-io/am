package filestorage

import (
	"bytes"
	"context"
	"log"

	"github.com/linkai-io/am/pkg/secrets"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/pkg/retrier"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/linkai-io/am/am"
)

type S3Storage struct {
	region     string
	env        string
	session    *session.Session
	bucketPath string
}

func NewS3Storage(env, region string) *S3Storage {
	return &S3Storage{region: region, env: env}
}

func (s *S3Storage) Init(cache *secrets.SecretsCache) error {
	var err error
	s.session, err = session.NewSession(&aws.Config{Region: aws.String(s.region)})
	if err != nil {
		return err
	}
	s.bucketPath, err = cache.WebFilePath()
	return err
}

func (s *S3Storage) Write(ctx context.Context, address *am.ScanGroupAddress, data []byte) (string, string, error) {
	hashName := convert.HashData(data)
	fileName := PathFromData(address, hashName)
	if fileName == "null" {
		return "", "", nil
	}

	headObjectInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketPath),
		Key:    aws.String(fileName),
	}
	s3session := s3.New(s.session)

	out, err := s3session.HeadObjectWithContext(ctx, headObjectInput)
	if err != nil {
		return hashName, s.bucketPath + fileName, s.uploadWithRetry(ctx, s3session, fileName, data)
	}

	// already exists don't bother uploading again
	if out != nil {
		return hashName, s.bucketPath + fileName, nil
	}
	return "", "", nil
}

func (s *S3Storage) uploadWithRetry(ctx context.Context, s3session *s3.S3, fileName string, data []byte) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucketPath),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(data),
	}

	retryErr := retrier.RetryAttempts(func() error {
		_, err := s3session.PutObjectWithContext(ctx, input)

		if err == nil {
			return nil
		}
		if awsErr, ok := err.(awserr.Error); ok {
			log.Printf("%#v , %s\n", awsErr, awsErr.Message())
		}
		return err
	}, 5)
	return retryErr
}
