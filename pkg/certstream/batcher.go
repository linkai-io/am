package certstream

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/rs/zerolog/log"
)

type Batcher struct {
	bigDataClient am.BigDataService
	userContext   am.UserContext

	// for pooling
	batchCount int
	count      int32
	results    chan *am.CTSubdomain

	//
	doneCh chan struct{}
}

func NewBatcher(userContext am.UserContext, bigDataClient am.BigDataService, batchCount int) *Batcher {
	return &Batcher{
		bigDataClient: bigDataClient,
		userContext:   userContext,
		batchCount:    batchCount,
	}
}

func (b *Batcher) Init() error {
	b.doneCh = make(chan struct{})
	b.results = make(chan *am.CTSubdomain, b.batchCount)
	go b.InsertBatch()
	return nil
}

func (b *Batcher) Add(result *am.CTSubdomain) {
	select {
	case b.results <- result:
		atomic.AddInt32(&b.count, 1)
	}
}

func (b *Batcher) Drain() map[string]*am.CTSubdomain {
	results := make(map[string]*am.CTSubdomain, 0)
	for {
		select {
		case result := <-b.results:
			results[result.Subdomain] = result
			atomic.AddInt32(&b.count, -1)
			if len(results) >= b.batchCount {
				log.Info().Int("count", len(results)).Msg("Uploader Drained")
				return results
			}
		default:
			return results
		}
	}
}

func (b *Batcher) Count() int32 {
	return atomic.LoadInt32(&b.count)
}

func (b *Batcher) InsertBatch() {
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	for {
		select {
		case <-b.doneCh:
			addrs := b.Drain()
			b.update(addrs)
			return
		case <-t.C:
			addrs := b.Drain()
			b.update(addrs)
		}
	}
}

func (b *Batcher) Done() {
	addrs := b.Drain()
	b.update(addrs)
	close(b.doneCh)
}

func (b *Batcher) update(allSubdomains map[string]*am.CTSubdomain) {
	var err error

	if allSubdomains == nil || len(allSubdomains) == 0 {
		return
	}

	ctx := context.Background()
	// ugh, need to make them into seperate maps (etld -> map[subdomain]*CTSubdomain)
	etlds := make(map[string]map[string]*am.CTSubdomain)
	for _, subdomain := range allSubdomains {
		if _, exist := etlds[subdomain.ETLD]; !exist {
			etlds[subdomain.ETLD] = make(map[string]*am.CTSubdomain)
		}
		etlds[subdomain.ETLD][subdomain.Subdomain] = subdomain
	}

	// TODO: this doesn't make sense to set it to any date since we aren't querying bigquery *shrug*....
	devModeQueryTime := time.Date(2019, time.February, 13, 0, 0, 0, 0, time.Local)

	for etld, subdomains := range etlds {
		err = b.bigDataClient.AddCTSubdomains(ctx, b.userContext, etld, devModeQueryTime, subdomains)
		if err != nil {
			log.Error().Err(err).Msg("Unable to insert batch of certstream subdomains")
			continue
		}
		log.Info().Int("count", len(subdomains)).Msg("inserted subdomains")
	}

}
