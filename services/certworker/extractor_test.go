package certworker_test

import (
	"context"
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/services/certworker"
)

type mockUploader struct {
}

func (m *mockUploader) Add(result *certworker.Result) {

}

func TestExtractor(t *testing.T) {

	ctServer := &am.CTServer{
		URL:   "ct.googleapis.com/logs/argon2020/",
		Index: 0,
		Step:  64,
	}
	uploader := &mockUploader{}
	e := certworker.NewExtractor(uploader, ctServer, 5)
	serv, err := e.Run(context.Background())
	if err != nil {
		t.Fatalf("error running: %v\n", err)
	}
	if serv.Index == 0 {
		t.Fatalf("expected index to be incremented, got: 0")
	}
}
