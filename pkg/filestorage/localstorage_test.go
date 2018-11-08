package filestorage_test

import (
	"context"
	"os"
	"testing"

	"github.com/linkai-io/am/pkg/secrets"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/filestorage"
)

var testTempPath = "/tmp"

func TestLocalStorage(t *testing.T) {
	s := filestorage.NewLocalStorage()
	if err := s.Init(nil); err != nil {
		t.Fatalf("error initializing local storage: %v\n", err)
	}
	address := &am.ScanGroupAddress{
		OrgID:   1,
		GroupID: 1,
	}
	cache := secrets.NewSecretsCache("local", "")

	if err := s.Init(cache); err != nil {
		t.Fatalf("failed to init file storage: %#v\n", err)
	}

	data := []byte("some data")
	if _, _, err := s.Write(context.Background(), address, data); err != nil {
		t.Fatalf("error writing file: %v\n", err)
	}

	defer os.RemoveAll(testTempPath + "1")

	if _, err := os.Stat(testTempPath + "/1/1/b/a/f/3/4/baf34551fecb48acc3da868eb85e1b6dac9de356"); err != nil {
		t.Fatalf("expected file to exist, was not there")
	}
}
