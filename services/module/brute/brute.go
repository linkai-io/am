package brute

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"strings"
	"sync/atomic"

	"github.com/linkai-io/am/pkg/dnsclient"
	"golang.org/x/time/rate"
)

type Analyzer struct {
	ns         *dnsclient.Client
	subdomains []string
	domainCh   chan string
	doneCh     chan struct{}
	limiter    *rate.Limiter
	found      int32
}

func New(ns *dnsclient.Client) *Analyzer {
	return &Analyzer{ns: ns}
}

func (a *Analyzer) Init(limit int, bruteFile *os.File) error {
	defer bruteFile.Close()
	fileScanner := bufio.NewScanner(bruteFile)
	a.subdomains = make([]string, 0)
	a.limiter = rate.NewLimiter(rate.Limit(limit), 20)
	for fileScanner.Scan() {
		a.subdomains = append(a.subdomains, strings.TrimSpace(fileScanner.Text()))
	}
	a.domainCh = make(chan string, limit)
	a.doneCh = make(chan struct{})

	for i := 0; i < limit; i++ {
		go a.resolver(a.domainCh, a.doneCh)
	}
	return nil
}

func (a *Analyzer) AnalyzeZone(zone string) {
	var buf bytes.Buffer
	ctx := context.Background()

	for i := 0; i < len(a.subdomains); i++ {
		a.limiter.Wait(ctx)
		buf.WriteString(a.subdomains[i])
		buf.WriteString(".")
		buf.WriteString(zone)
		a.domainCh <- buf.String()
		buf.Reset()
	}
}

func (a *Analyzer) resolver(domainCh chan string, doneCh chan struct{}) {
	for {
		select {
		case domain := <-domainCh:
			r, err := a.ns.ResolveName(domain)
			if err != nil && err != dnsclient.ErrEmptyRecords {
				continue
			}
			if r != nil && len(r) > 0 {
				atomic.AddInt32(&a.found, 1)
				/*
					for _, record := range r {
						log.Info().Printf("%#v\n", record)
					}*/
			}
		case <-doneCh:
			return
		}
	}
}

func (a *Analyzer) Quit() {
	a.doneCh <- struct{}{}
}
