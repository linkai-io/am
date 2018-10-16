package filestorage_test

import (
	"context"
	"os"
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/filestorage"
)

var testTempPath = "/tmp"

func TestLocalStorage(t *testing.T) {
	s := filestorage.NewLocalStorage(testTempPath)
	if err := s.Init(nil); err != nil {
		t.Fatalf("error initializing local storage: %v\n", err)
	}
	address := &am.ScanGroupAddress{
		OrgID:   1,
		GroupID: 1,
	}
	data := []byte("some data")
	if err := s.Write(context.Background(), address, data); err != nil {
		t.Fatalf("error writing file: %v\n", err)
	}

	defer os.RemoveAll(testTempPath + "1")

	if _, err := os.Stat(testTempPath + "/1/1/b/a/f/3/4/baf34551fecb48acc3da868eb85e1b6dac9de356"); err != nil {
		t.Fatalf("expected file to exist, was not there")
	}
}
