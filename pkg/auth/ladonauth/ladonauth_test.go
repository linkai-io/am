package ladonauth_test

import (
	"flag"
)

var env string

func init() {
	flag.StringVar(&env, "env", "local", "environment we are running tests in")
	flag.Parse()
}
