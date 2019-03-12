package convert_test

import (
	"testing"

	"github.com/linkai-io/am/pkg/convert"
)

func TestHashAddress(t *testing.T) {
	val := convert.HashAddress("2600:9000:2015:800:8:5c48:ab80:93a1", "test.linkai.io")
	val2 := convert.HashAddress("2600:9000:201b:3a00:8:5c48:ab80:93a1", "test.linkai.io")
	if val == val2 {
		t.Fatalf("values were equal")
	}
	val3 := convert.HashAddress("13.249.44.82", "test.linkai.io")
	t.Logf("%s", val3)
}
