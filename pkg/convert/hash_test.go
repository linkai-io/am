package convert_test

import (
	"testing"

	"github.com/linkai-io/am/pkg/convert"
)

func TestHashAddress(t *testing.T) {
	/*
		val := convert.HashAddress("2600:9000:2015:800:8:5c48:ab80:93a1", "test.linkai.io")
		val2 := convert.HashAddress("2600:9000:201b:3a00:8:5c48:ab80:93a1", "test.linkai.io")
		if val == val2 {
			t.Fatalf("values were equal")
		}*/
	val4 := convert.HashAddress("2600:9000:203c:8c00:8:5c48:ab80:93a1", "test.linkai.io") //                         | 878dae8433f62a78a1a4b76ffe562919
	//"13.249.44.76")//                         | 8f1c3d94763f991b11ebee48f1e98361
	//"13.249.44.38")//                         | 39e916789af9c452020cb8e74c595980

	t.Logf("%s", val4)
	/*
		/ test.linkai.io | 06e219d8ccd6bd302b5883b95d08488a |
		test.linkai.io | 9598ac321746a729ed57d4950ff2a7ca | 54.192.81.106
		test.linkai.io | 11aab53066aab7cd8974a006b2183bb0 | 54.192.81.106
		test.linkai.io | d67ee8bbeaaa084cb80c600a095e5725 | 54.192.81.106
		test.linkai.io | eddc25ea655d88aabb346c40df458c6f | 54.192.81.106
	*/
}
