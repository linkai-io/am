package sesmailer_test

import (
	"testing"

	"github.com/linkai-io/am/pkg/mail/sesmailer"
)

func TestSendMail(t *testing.T) {
	t.Skip("skipping mail test")
	m := sesmailer.New("dev", "us-east-1")
	if err := m.Init(nil); err != nil {
		t.Fatalf("failed to initalize mailer: %v\n", err)
	}

	if err := m.SendMail("hello", "isaac.dawson@linkai.io", "<h1>hello from hakken</h1>", "hello from hakken"); err != nil {
		t.Fatalf("error: %#v\n", err)
	}
}
