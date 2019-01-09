package consul

import (
	"context"
	"strconv"
	"sync"
	"time"

	api "github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/resolver"
)

// RegisterDefault resolver how often to watch for updates
// Address comes from CONSUL_HTTP_ADDR env var
func RegisterDefault(watchInterval time.Duration, addr string) {
	cfg := api.DefaultConfig()
	cfg.Address = addr
	resolver.Register(&ResolverBuilder{
		WatchInterval:      watchInterval,
		ConsulClientConfig: cfg,
	})
}

type Resolver struct {
	lock          sync.RWMutex
	target        resolver.Target
	cc            resolver.ClientConn
	consul        *api.Client
	addr          chan []resolver.Address
	done          chan struct{}
	watchInterval time.Duration
}

func (r *Resolver) ResolveNow(resolver.ResolveNowOption) {
	r.resolve()
}

func (r *Resolver) Close() {
	close(r.done)
}

func (r *Resolver) updater() {
	for {
		select {
		case addrs := <-r.addr:
			r.cc.NewAddress(addrs)
		case <-r.done:
			return
		}
	}
}

func (r *Resolver) watcher() {
	ticker := time.NewTicker(r.watchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.resolve()
		case <-r.done:
			return
		}
	}
}

func (r *Resolver) resolve() {
	r.lock.Lock()
	defer r.lock.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	q := &api.QueryOptions{}
	q.WithContext(ctx)

	services, _, err := r.consul.Catalog().Service(r.target.Endpoint, "", q)
	if err != nil {
		log.Error().Err(err).Msg("failed to update services")
		return
	}

	addresses := make([]resolver.Address, 0, len(services))

	for _, s := range services {
		address := s.ServiceAddress
		port := s.ServicePort
		if address == "" {
			address = s.Address
		}

		addresses = append(addresses, resolver.Address{
			Addr:       address + ":" + strconv.Itoa(port),
			ServerName: r.target.Endpoint,
		})
	}
	r.addr <- addresses
}
