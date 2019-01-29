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
	s := filestorage.NewLocalStorage()

	address := &am.ScanGroupAddress{
		OrgID:   1,
		GroupID: 1,
	}

	userContext := &am.UserContextData{
		OrgID:  1,
		UserID: 1,
		OrgCID: "/tmp",
	}
	data := []byte("some data")
	if _, _, err := s.Write(context.Background(), userContext, address, data); err != nil {
		t.Fatalf("error writing file: %v\n", err)
	}

	defer os.RemoveAll(testTempPath + "/b")

	if _, err := os.Stat(testTempPath + "/b/a/f/3/4/baf34551fecb48acc3da868eb85e1b6dac9de356"); err != nil {
		t.Fatalf("expected file to exist, was not there")
	}
}
