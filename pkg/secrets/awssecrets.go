package secrets

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// AWSSecrets retrieves secrets from AWS Parameter Store
type AWSSecrets struct {
	Region  string
	sess    *session.Session
	manager *ssm.SSM
}

// NewAWSSecrets returns an instance with optional region specified, otherwise uses us-east-1
func NewAWSSecrets(region string) *AWSSecrets {
	if region == "" {
		region = "us-east-1"
	}
	s := &AWSSecrets{Region: region}
	s.sess = session.Must(session.NewSession(&aws.Config{Region: aws.String(s.Region)}))
	s.manager = ssm.New(s.sess)
	return s
}

// WithCredentials creates a session with credentials
func (s *AWSSecrets) WithCredentials(id, key string) {
	creds := credentials.NewStaticCredentials(id, key, "")
	s.sess = session.Must(session.NewSession(&aws.Config{Credentials: creds, Region: aws.String(s.Region)}))
	s.manager = ssm.New(s.sess)
	return
}

// GetSecureParameter retrieves the parameter specified by key, or error otherwise.
func (s *AWSSecrets) GetSecureParameter(key string) ([]byte, error) {
	decrypt := true
	parameter := &ssm.GetParameterInput{
		Name:           &key,
		WithDecryption: &decrypt,
	}
	out, err := s.manager.GetParameter(parameter)
	if err != nil {
		return nil, err
	}

	return []byte(*out.Parameter.Value), nil
}

func (s *AWSSecrets) SetSecureParameter(key, value string) error {
	parameter := &ssm.PutParameterInput{
		KeyId:     aws.String("alias/aws/ssm"),
		Name:      &key,
		Overwrite: aws.Bool(true),
		Type:      aws.String("SecureString"),
		Value:     aws.String(value),
	}
	_, err := s.manager.PutParameter(parameter)
	return err
}
