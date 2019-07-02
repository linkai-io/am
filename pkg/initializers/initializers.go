package initializers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/linkai-io/am/clients/event"
	"github.com/linkai-io/am/pkg/bq"

	"github.com/linkai-io/am/pkg/discovery"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/clients/address"
	bdc "github.com/linkai-io/am/clients/bigdata"
	"github.com/linkai-io/am/clients/coordinator"
	"github.com/linkai-io/am/clients/dispatcher"
	"github.com/linkai-io/am/clients/module"
	"github.com/linkai-io/am/clients/organization"
	"github.com/linkai-io/am/clients/scangroup"
	"github.com/linkai-io/am/clients/webdata"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/pkg/state/redis"
	"github.com/rs/zerolog/log"
)

const (
	tenMinutes    = 600
	thirtyMinutes = tenMinutes * 3
	sixtyMinutes  = tenMinutes * 6
)

// AppConfig represents values taken from environment variables
type AppConfig struct {
	Env          string
	Region       string
	SelfRegister string
	ServiceKey   string
	Addr         string
}

func ServiceDiscovery(appConfig *AppConfig) string {
	consulAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddr != "" {
		return consulAddr
	}

	resp, err := http.Get("http://169.254.169.254/latest/meta-data/local-ipv4")
	if err != nil {
		log.Fatal().Err(err).Str("serviceKey", appConfig.ServiceKey).Msg("unable to get consul addr")
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err).Str("serviceKey", appConfig.ServiceKey).Msg("unable to get consul addr")
	}

	log.Info().Str("serviceKey", appConfig.ServiceKey).Str("consul_addr", string(data)).Msg("got consul address")
	return fmt.Sprintf("%s:8500", string(data))
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
		log.Info().Str("addr", addr).Str("service", appConfig.ServiceKey).Msg("attempting to connect to redis")
		return redisState.Init(addr, pass)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to redis")
	}
	return redisState
}

// DispatcherClient connects to the dispatcher service
func DispatcherClient() am.DispatcherService {
	dispatcherClient := dispatcher.New()

	err := retrier.RetryUntil(func() error {
		return dispatcherClient.Init(nil)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to dispatcher server")
	}
	return dispatcherClient
}

// SGClient connects to the scangroup service
func SGClient() am.ScanGroupService {
	scanGroupClient := scangroup.New()

	err := retrier.RetryUntil(func() error {
		return scanGroupClient.Init(nil)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to scangroup server")
	}
	return scanGroupClient
}

// OrgClient connects to the organization service
func OrgClient() am.OrganizationService {
	orgClient := organization.New()

	err := retrier.RetryUntil(func() error {
		return orgClient.Init(nil)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to organization server")
	}
	return orgClient
}

// AddrClient connects to the address service
func AddrClient() am.AddressService {
	addrClient := address.New()

	err := retrier.RetryUntil(func() error {
		return addrClient.Init(nil)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to address server")
	}
	return addrClient
}

// AddrClientWithTimeout connects to the address service with specified timeout for all calls
func AddrClientWithTimeout(timeout time.Duration) am.AddressService {
	addrClient := address.New()
	addrClient.SetTimeout(timeout)
	err := retrier.RetryUntil(func() error {
		return addrClient.Init(nil)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to address server")
	}
	return addrClient
}

// EventClient connects to the address service
func EventClient() am.EventService {
	eventClient := event.New()

	err := retrier.RetryUntil(func() error {
		return eventClient.Init(nil)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to event server")
	}
	return eventClient
}

// CoordClient connects to the coordinator service
func CoordClient() am.CoordinatorService {
	coordClient := coordinator.New()

	err := retrier.RetryUntil(func() error {
		return coordClient.Init(nil)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("unable to connect to coordinator client")
	}
	return coordClient
}

// WebDataClient connects to the webdata service
func WebDataClient() am.WebDataService {
	webDataClient := webdata.New()

	err := retrier.RetryUntil(func() error {
		return webDataClient.Init(nil)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to webdata server")
	}
	return webDataClient
}

// WebDataClient connects to the webdata service with specified timeout for all calls
func WebDataClientWithTimeout(timeout time.Duration) am.WebDataService {
	webDataClient := webdata.New()
	webDataClient.SetTimeout(timeout)

	err := retrier.RetryUntil(func() error {
		return webDataClient.Init(nil)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to webdata server")
	}
	return webDataClient
}

// BigDataClient connects to the bigdata service
func BigDataClient() am.BigDataService {
	bigDataClient := bdc.New()

	err := retrier.RetryUntil(func() error {
		return bigDataClient.Init(nil)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error connecting to bigdata server")
	}
	return bigDataClient
}

func BigQueryClient(cfg *bq.ClientConfig, credentials []byte) bq.BigQuerier {
	bqClient := bq.NewClient()
	cfgData, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal bigquery config")
	}

	err = retrier.RetryUntil(func() error {
		return bqClient.Init(cfgData, credentials)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("error initializing bigquery client")
	}
	return bqClient
}

// Module returns the connected module depending on moduleType
func Module(state *redis.State, moduleType am.ModuleType) am.ModuleService {
	switch moduleType {
	case am.NSModule:
		nsClient := module.New()
		cfg := &module.Config{ModuleType: am.NSModule, Timeout: tenMinutes}
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
		cfg := &module.Config{ModuleType: am.BruteModule, Timeout: sixtyMinutes}
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
		cfg := &module.Config{ModuleType: am.WebModule, Timeout: tenMinutes}
		data, _ := json.Marshal(cfg)

		err := retrier.RetryUntil(func() error {
			return webClient.Init(data)
		}, time.Minute*1, time.Second*3)

		if err != nil {
			log.Fatal().Err(err).Msg("unable to connect to web module client")
		}
		return webClient
	case am.BigDataCTSubdomainModule:
		bdClient := module.New()
		cfg := &module.Config{ModuleType: am.BigDataCTSubdomainModule, Timeout: tenMinutes}
		data, _ := json.Marshal(cfg)
		err := retrier.RetryUntil(func() error {
			return bdClient.Init(data)
		}, time.Minute*1, time.Second*3)

		if err != nil {
			log.Fatal().Err(err).Msg("unable to connect to bigdata module client")
		}
		return bdClient
	}
	return nil
}

// PortScanModule connects directly with our port scanner service
func PortScanModule(state *redis.State) am.PortScannerService {
	portScan := module.NewPortScanClient()
	cfg := &module.Config{ModuleType: am.PortScanModule, Timeout: thirtyMinutes}
	data, _ := json.Marshal(cfg)

	err := retrier.RetryUntil(func() error {
		return portScan.Init(data)
	}, time.Minute*1, time.Second*3)

	if err != nil {
		log.Fatal().Err(err).Msg("unable to connect to portScan module client")
	}
	return portScan
}

// Modules initializes all modules and connects to them
func Modules(state *redis.State) map[am.ModuleType]am.ModuleService {
	modules := make(map[am.ModuleType]am.ModuleService)
	modules[am.NSModule] = Module(state, am.NSModule)
	modules[am.BruteModule] = Module(state, am.BruteModule)
	modules[am.WebModule] = Module(state, am.WebModule)
	modules[am.BigDataCTSubdomainModule] = Module(state, am.BigDataCTSubdomainModule)
	return modules
}
