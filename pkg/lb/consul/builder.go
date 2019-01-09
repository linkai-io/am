package consul

import (
	"time"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/resolver"
)

// Scheme to prefix for lookups (srv://consul/<service>)
const (
	Scheme = "srv"
)

// ResolverBuilder builds our name resolver for Consul
type ResolverBuilder struct {
	WatchInterval      time.Duration
	ConsulClientConfig *api.Config
}

// Scheme returns the scheme of this builder
func (b *ResolverBuilder) Scheme() string {
	return Scheme
}

// Build the consul resolver by calling the consul api and starting watchers/updaters for resolving server addresses
func (b *ResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	consul, err := api.NewClient(b.ConsulClientConfig)
	if err != nil {
		return nil, err
	}

	r := Resolver{
		target:        target,
		cc:            cc,
		consul:        consul,
		addr:          make(chan []resolver.Address, 1),
		done:          make(chan struct{}, 1),
		watchInterval: b.WatchInterval,
	}

	go r.updater()
	go r.watcher()
	r.resolve()

	return &r, nil
}
