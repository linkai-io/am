package webtech_test

import (
	"testing"

	"github.com/linkai-io/am/pkg/webtech"
)

func TestWappalyzer(t *testing.T) {
	w := webtech.NewWappalyzer("https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/apps.json")
	if err := w.Init(); err != nil {
		t.Fatalf("error loading wappalyzer data: %v\n", err)
	}
}
