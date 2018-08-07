package e2e_test

import (
	"flag"
)

var enableTests bool

func init() {
	flag.BoolVar(&enableTests, "enable", false, "pass true to enable e2e tests")
}
