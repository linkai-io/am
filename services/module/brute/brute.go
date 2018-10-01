package brute

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/cache"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/brute/state"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
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

func (b *Bruter) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {
	nsLog := log.With().
		Int("OrgID", userContext.GetOrgID()).
		Int("UserID", userContext.GetUserID()).
		Str("TraceID", userContext.GetTraceID()).
		Str("IPAddress", address.IPAddress).
		Str("HostAddress", address.HostAddress).
		Int64("AddressID", address.AddressID).
		Str("AddressHash", address.AddressHash).Logger()
	nsRecords := make(map[string]*am.ScanGroupAddress, 0)
	if !b.shouldAnalyze(address) {
		nsLog.Info().Msg("not analyzing")
		return address, nsRecords, nil
	}

	return address, nsRecords, nil
}

// shouldAnalyze determines if we should analyze the specific address or not
func (b *Bruter) shouldAnalyze(address *am.ScanGroupAddress) bool {
	if address.IsWildcardZone || address.IsHostedService {
		return false
	}

	switch uint16(address.NSRecord) {
	case dns.TypeMX, dns.TypeNS, dns.TypeSRV:
		return false
	}
	return true
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
