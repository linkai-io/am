package set_test

import (
	"testing"

	"github.com/linkai-io/am/pkg/set"
)

func TestInt32Intersection(t *testing.T) {
	a := []int32{1, 4, 6}
	b := []int32{4, 8, 9, 10}
	r := set.Int32Intersection(a, b)
	if len(r) != 1 && r[0] != 4 {
		t.Fatalf("expected []int32{4}, got %v\n", r)
	}
}
