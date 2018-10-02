package generators_test

import (
	"testing"

	"github.com/linkai-io/am/pkg/generators"
)

func TestInsecureAlphabetString(t *testing.T) {
	length := 8
	result := generators.InsecureAlphabetString(length)
	if len(result) != length {
		t.Fatalf("expected length of %d\n", length)
	}
}
