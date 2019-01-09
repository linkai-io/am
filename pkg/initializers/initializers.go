package initializers

import (
	"context"
	"encoding/json"
	"net"
	"strconv"
	"time"

	"github.com/linkai-io/am/pkg/discovery"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/address"
	bdc "github.com/linkai-io/am/clients/bigdata"
	"github.com/linkai-io/am/clients/coordinator"
	"github.com/linkai-io/am/clients/dispatcher"
	"github.com/linkai-io/am/clients/module"
	"github.com/linkai-io/am/clients/scangroup"
	"github.com/linkai-io/am/clients/webdata"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/pkg/state/redis"
	"github.com/rs/zerolog/log"
)

// AppConfig represents values taken from environment variables
type AppConfig struct {
	Env          string
	Region       string
	SelfRegister string
	ServiceKey   string
	Addr         string
}

// Self registers if SelfRegister is set to anything. Assumes valid host:port pair in appConfig.Addr is set
func Self(ctx context.Context, appConfig *AppConfig) {
	if appConfig.SelfRegister == "" {
		return
	}

	discoveryAddr := "127.0.0.1:8500"

	host, portStr, err := net.SplitHostPort(appConfig.Addr)
	if err != nil {
		log.Fatal().Err(err).Str("serviceKey", appConfig.ServiceKey).Msg("unable to get hostport")
	}

	port, _ := strconv.Atoi(portStr)

	disco := discovery.New(discoveryAddr, appConfig.ServiceKey, host, port, time.Second*60)
	if err := disco.SelfRegister(ctx); err != nil {
		log.Fatal().Err(err).Str("serviceKey", appConfig.ServiceKey).Msg("unable to get self register")
	}
}

// DB for environment, in region, for serviceKey service.
func DB(appConfig *AppConfig) (string, *pgx.ConnPool) {
	sec := secrets.NewSecretsCache(appConfig.Env, appConfig.Region)
	dbstring, err := sec.DBString(appConfig.ServiceKey)
	if err != nil {
		log.Fatal().Err(err).Str("serviceKey", appConfig.ServiceKey).Msg("unable to get dbstring")
	}

	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		log.Fatal().Err(err).Str("serviceKey", appConfig.ServiceKey).Msg("error parsing connection string")
	}

	var p *pgx.ConnPool

	err = retrier.RetryUntil(func() error {
		p, err = pgx.NewConnPool(pgx.ConnPoolConfig{
			ConnConfig:     conf,
			MaxConnections: 5,
		})
		return err
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Str("serviceKey", appConfig.ServiceKey).Msg("failed to connect to postgresql")
	}
	return dbstring, p
}

// State connects to the state system (redis)
func State(appConfig *AppConfig) *redis.State {
	redisState := redis.New()
	sec := secrets.NewSecretsCache(appConfig.Env, appConfig.Region)
	addr, err := sec.StateAddr()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get state address")
	}
	pass, err := sec.StatePassword()
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get state password")
	}

	err = retrier.RetryUntil(func() error {
		log.Info().Msg("attempting to connect to redis")
		return redisState.Init(addr, pass)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to redis")
	}
	return redisState
}

func DispatcherClient(loadBalancerAddr string) am.DispatcherService {
	dispatcherClient := dispatcher.New()

	err := retrier.RetryUntil(func() error {
		return dispatcherClient.Init([]byte(loadBalancerAddr))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to dispatcher server")
	}
	return dispatcherClient
}

// SGClient connects to the scangroup service via load balancer
func SGClient(loadBalancerAddr string) am.ScanGroupService {
	scanGroupClient := scangroup.New()

	err := retrier.RetryUntil(func() error {
		return scanGroupClient.Init([]byte(loadBalancerAddr))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to scangroup server")
	}
	return scanGroupClient
}

// AddrClient connects to the address service via load balancer
func AddrClient(loadBalancerAddr string) am.AddressService {
	addrClient := address.New()

	err := retrier.RetryUntil(func() error {
		return addrClient.Init([]byte(loadBalancerAddr))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to address server")
	}
	return addrClient
}

// CoordClient connects to the coordinator service via the load balancer
func CoordClient(loadBalancerAddr string) am.CoordinatorService {
	coordClient := coordinator.New()

	err := retrier.RetryUntil(func() error {
		return coordClient.Init([]byte(loadBalancerAddr))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("unable to connect to coordinator client")
	}
	return coordClient
}

// WebDataClient connects to the webdata service via load balancer
func WebDataClient(loadBalancerAddr string) am.WebDataService {
	webDataClient := webdata.New()

	err := retrier.RetryUntil(func() error {
		return webDataClient.Init([]byte(loadBalancerAddr))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to webdata server")
	}
	return webDataClient
}

// BigDataClient connects to the bigdata service via load balancer
func BigDataClient(loadBalancerAddr string) am.BigDataService {
	bigDataClient := bdc.New()

	err := retrier.RetryUntil(func() error {
		return bigDataClient.Init([]byte(loadBalancerAddr))
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to bigdata server")
	}
	return bigDataClient
}

// Module returns the connected module depending on moduleType
func Module(state *redis.State, loadBalancerAddr string, moduleType am.ModuleType) am.ModuleService {
	switch moduleType {
	case am.NSModule:
		nsClient := module.New()
		cfg := &module.Config{Addr: loadBalancerAddr, ModuleType: am.NSModule}
		data, _ := json.Marshal(cfg)

		err := retrier.RetryUntil(func() error {
			return nsClient.Init(data)
		}, time.Minute*1, time.Second*3)

		if err != nil {
			log.Fatal().Err(err).Msg("unable to connect to ns module client")
		}
		return nsClient
	case am.BruteModule:
		bruteClient := module.New()
		cfg := &module.Config{Addr: loadBalancerAddr, ModuleType: am.BruteModule, Timeout: 600}
		data, _ := json.Marshal(cfg)

		err := retrier.RetryUntil(func() error {
			return bruteClient.Init(data)
		}, time.Minute*1, time.Second*3)

		if err != nil {
			log.Fatal().Err(err).Msg("unable to connect to brute module client")
		}
		return bruteClient
	case am.WebModule:
		webClient := module.New()
		cfg := &module.Config{Addr: loadBalancerAddr, ModuleType: am.WebModule, Timeout: 600}
		data, _ := json.Marshal(cfg)

		err := retrier.RetryUntil(func() error {
			return webClient.Init(data)
		}, time.Minute*1, time.Second*3)

		if err != nil {
			log.Fatal().Err(err).Msg("unable to connect to web module client")
		}
		return webClient
	}
	return nil
}

// Modules initializes all moduels and connects to them via load balancer address.
func Modules(state *redis.State, loadBalancerAddr string) map[am.ModuleType]am.ModuleService {
	modules := make(map[am.ModuleType]am.ModuleService)
	modules[am.NSModule] = Module(state, loadBalancerAddr, am.NSModule)
	modules[am.BruteModule] = Module(state, loadBalancerAddr, am.BruteModule)
	modules[am.WebModule] = Module(state, loadBalancerAddr, am.WebModule)
	return modules
}
