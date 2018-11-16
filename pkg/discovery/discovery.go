package discovery

import (
	"context"
	"fmt"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
)

// HealthFN is a health check function that runs every ttl / 2 to let consul
// know the service is up/healthy
type HealthFN func() error

// Discovery implementation to allow a service to self register with consul
type Discovery struct {
	serviceKey  string
	address     string
	port        int
	ttl         time.Duration
	consulAddr  string
	consulAgent *consul.Agent
	healthFn    HealthFN
	id          string
}

// New Discovery register
func New(consulAddr, serviceKey, address string, port int, ttl time.Duration) *Discovery {
	return &Discovery{
		serviceKey: serviceKey,
		address:    address,
		port:       port,
		ttl:        ttl,
		consulAddr: consulAddr,
		healthFn:   func() error { return nil },
		id:         fmt.Sprintf("%s.%s.%d", serviceKey, address, port),
	}
}

// SetHealthFN for allowing custom health checks to be called every ttl / 2.
func (d *Discovery) SetHealthFN(fn HealthFN) {
	d.healthFn = fn
}

// SelfRegister with consul
func (d *Discovery) SelfRegister(ctx context.Context) error {
	cfg := consul.DefaultConfig()
	cfg.Address = d.consulAddr

	c, err := consul.NewClient(cfg)
	if err != nil {
		return err
	}
	d.consulAgent = c.Agent()

	serviceDef := &consul.AgentServiceRegistration{
		ID:      d.id,
		Name:    d.serviceKey,
		Address: d.address,
		Port:    d.port,
		Check: &consul.AgentServiceCheck{
			TTL: d.ttl.String(),
		},
	}

	if err := d.consulAgent.ServiceRegister(serviceDef); err != nil {
		return err
	}
	go d.updateTTL(ctx, d.healthFn)
	return nil
}

func (d *Discovery) updateTTL(ctx context.Context, fn HealthFN) {
	ticker := time.NewTicker(d.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			d.update(d.healthFn)
		case <-ctx.Done():
			return
		}
	}
}

func (d *Discovery) update(fn HealthFN) {
	note := "OK"
	status := "pass"

	if err := fn(); err != nil {
		note = "NG"
		status = "fail"
	}

	if err := d.consulAgent.UpdateTTL("service:"+d.id, note, status); err != nil {
		log.Warn().Err(err).Msg("failed to update TTL in consul")
	}
}
