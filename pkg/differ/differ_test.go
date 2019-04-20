package differ_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/linkai-io/am/pkg/differ"
)

func TestDiff(t *testing.T) {
	f1, err := ioutil.ReadFile("testdata/dom1.html")
	if err != nil {
		t.Fatalf("error reading file 1 %v\n", err)
	}

	f2, err := ioutil.ReadFile("testdata/dom2.html")
	if err != nil {
		t.Fatalf("error reading file 2 %v\n", err)
	}
	fmt.Printf("%s and %s\n", f1[0:10], f2[0:10])
	d := differ.New()
	d.DiffRemove(string(f1), string(f2))
}
