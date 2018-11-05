package brute_test

import (
	"testing"

	"github.com/linkai-io/am/services/module/brute"
)

func TestNumberMutations(t *testing.T) {
	sub := "prod5srv1"
	results := brute.NumberMutation(sub)
	if len(results) != 16 {
		t.Fatalf("expected 16 mutations got: %d\n", len(results))
	}

	for _, result := range results {
		t.Logf("%s\n", result)
	}
}
