package brute_test

import (
	"testing"

	"github.com/linkai-io/am/services/module/brute"
)

func TestNumberMutations(t *testing.T) {
	/*sub := "prod5srv8"
	results := brute.NumberMutation(sub)
	for _, result := range results {
		t.Logf("%s\n", result)
	}*/

	sub := "prod5srv1"
	results := brute.NumberMutation(sub)
	for _, result := range results {
		t.Logf("%s\n", result)
	}
}
