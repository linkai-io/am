package sesmailer

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/pkg/errors"
)

const (
	charSet = "UTF-8"
)

type Mail struct {
	env    string
	region string
	sender string
	svc    *ses.SES
}

func New(env, region string) *Mail {
	sender := fmt.Sprintf("noreply@%s.noreply.linkai.io", env)
	if env == "prod" {
		sender = fmt.Sprintf("noreply@noreply.linkai.io")
	}

	return &Mail{env: env, region: region, sender: sender}
}

func (m *Mail) Init(config []byte) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(m.region)},
	)
	if err != nil {
		return err
	}
	// Create an SES session.
	m.svc = ses.New(sess)
	return nil
}

func (m *Mail) SendMail(subject, to, html, raw string) error {
	// Assemble the email.
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(to),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(html),
				},
				Text: &ses.Content{
					Charset: aws.String(charSet),
					Data:    aws.String(raw),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String(charSet),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(m.sender),
		// Uncomment to use a configuration set
		//ConfigurationSetName: aws.String(ConfigurationSet),
	}

	// Attempt to send the email.
	_, err := m.svc.SendEmail(input)

	// Display error messages if they occur.
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return errors.Wrap(aerr, "failed to send mail")
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			errors.Wrap(err, "failed to send mail")
		}
	}
	return nil
}
