package brute

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/brute/state"
	"golang.org/x/time/rate"
)

type Bruter struct {
	st         state.Stater
	dc         *dnsclient.Client
	subdomains []string
	domainCh   chan string
	doneCh     chan struct{}
	limiter    *rate.Limiter
	found      int32

	// for closing subscriptions to listen for group updates
	exitContext context.Context
	cancel      context.CancelFunc
	// concurrent safe cache of scan groups updated via Subscribe callbacks
	groupCache *cache.ScanGroupSubscriber
}

func New(dc *dnsclient.Client, st state.Stater) *Bruter {
	ctx, cancel := context.WithCancel(context.Background())
	b := &Bruter{st: st, exitContext: ctx, cancel: cancel}
	b.dc = dc
	// start cache subscriber and listen for updates
	b.groupCache = cache.NewScanGroupSubscriber(ctx, st)
	return b
}

func (b *Bruter) Init(limit int, bruteFile *os.File) error {
	defer bruteFile.Close()
	fileScanner := bufio.NewScanner(bruteFile)
	b.subdomains = make([]string, 0)
	b.limiter = rate.NewLimiter(rate.Limit(limit), 20)
	for fileScanner.Scan() {
		b.subdomains = append(b.subdomains, strings.TrimSpace(fileScanner.Text()))
	}
	b.domainCh = make(chan string, limit)
	b.doneCh = make(chan struct{})

	for i := 0; i < limit; i++ {
		go b.resolver(b.domainCh, b.doneCh)
	}
	return nil
}

func (b *Bruter) AnalyzeZone(zone string) {
	var buf bytes.Buffer
	ctx := context.Background()

	for i := 0; i < len(b.subdomains); i++ {
		b.limiter.Wait(ctx)
		buf.WriteString(b.subdomains[i])
		buf.WriteString(".")
		buf.WriteString(zone)
		b.domainCh <- buf.String()
		buf.Reset()
	}
}

func (b *Bruter) resolver(domainCh chan string, doneCh chan struct{}) {
	/*for {
		select {
		case domain := <-domainCh:
			r, err := b.ns.ResolveName(domain)
			if err != nil && err != dnsclient.ErrEmptyRecords {
				continue
			}
			if r != nil && len(r) > 0 {
				atomic.AddInt32(&b.found, 1)
				/*
					for _, record := range r {
						log.Info().Printf("%#v\n", record)
					}
			}
		case <-doneCh:
			return
		}
	}
	*/
}

func (b *Bruter) Quit() {
	b.doneCh <- struct{}{}
}
