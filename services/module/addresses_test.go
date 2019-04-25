package module_test

import (
	"testing"

	"github.com/gammazero/workerpool"
)

func TestWorkerPool(t *testing.T) {
	pool := workerpool.New(5)

	out := make(chan string, 5) // how many results we expect
	go func() {
		for result := range out {
			t.Logf("%s\n", result)
		}
	}()
	task := func(h string, out chan<- string) func() {
		return func() {
			t.Logf("saw %s\n", h)
			out <- "hi" + h
		}
	}

	// submit all hosts to our worker pool
	for _, newHost := range []string{"hi", "bhi", "st", "x", "y", "z"} {
		h := newHost
		pool.Submit(task(h, out))
	}

	pool.StopWait()
	close(out)
	t.Logf("DONE\n")

}
