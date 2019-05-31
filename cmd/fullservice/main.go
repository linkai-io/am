package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/linkai-io/am/amtest"
	bdservice "github.com/linkai-io/am/services/bigdata"
	"github.com/linkai-io/am/services/event"
	"github.com/linkai-io/am/services/module/bigdata"
	"github.com/linkai-io/am/services/module/brute"
	"github.com/linkai-io/am/services/module/ns"
	"github.com/linkai-io/am/services/module/web"
	"github.com/linkai-io/am/services/organization"
	"github.com/linkai-io/am/services/scangroup"
	"github.com/linkai-io/am/services/webdata"

	"github.com/rs/zerolog/log"

	"github.com/jackc/pgx"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/auth/ladonauth"
	"github.com/linkai-io/am/pkg/bq"
	"github.com/linkai-io/am/pkg/browser"
	"github.com/linkai-io/am/pkg/certstream"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/pkg/filestorage"
	"github.com/linkai-io/am/pkg/initializers"
	"github.com/linkai-io/am/pkg/webtech"
	"github.com/linkai-io/am/services/address"

	"github.com/linkai-io/am/services/dispatcher"
)

var dnsServers = []string{"1.1.1.1:853", "1.0.0.1:853", "64.6.64.6:53", "77.88.8.8:53", "74.82.42.42:53", "8.8.4.4:53", "8.8.8.8:53"}

var inputFile string
var orgName string
var deleteOnExit bool
var createOrg bool

func init() {
	flag.StringVar(&inputFile, "input", "testdata/input.txt", "input file to use")
	flag.StringVar(&orgName, "org", "fullserviceorg", "name of org to use for testing")
	flag.BoolVar(&deleteOnExit, "delete", false, "delete org/data on exit")
	flag.BoolVar(&createOrg, "create", false, "to create the org or not")
}

func createAppConfig(serviceKey string) *initializers.AppConfig {
	appConfig := &initializers.AppConfig{}
	appConfig.Env = os.Getenv("APP_ENV")
	appConfig.Region = os.Getenv("APP_REGION")
	appConfig.SelfRegister = os.Getenv("APP_SELF_REGISTER")
	appConfig.Addr = os.Getenv("APP_ADDR")
	appConfig.ServiceKey = serviceKey
	return appConfig
}

func authorizer(db *pgx.ConnPool) (*ladonauth.LadonAuthorizer, *ladonauth.LadonRoleManager) {
	policyManager := ladonauth.NewPolicyManager(db, "pgx")
	if err := policyManager.Init(); err != nil {
		log.Fatal().Err(err).Msg("initializing policyManager failed")
	}

	roleManager := ladonauth.NewRoleManager(db, "pgx")
	if err := roleManager.Init(); err != nil {
		log.Fatal().Err(err).Msg("initializing roleManager failed")
	}
	return ladonauth.NewLadonAuthorizer(policyManager, roleManager), roleManager
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbstring, db := initializers.DB(createAppConfig(am.AddressServiceKey))
	auth, _ := authorizer(db)
	addressService := address.New(auth)
	if err := addressService.Init([]byte(dbstring)); err != nil {
		log.Fatal().Err(err).Msg("failed to init address")
	}
	log.Info().Msg("address init'd")

	dbstring, db = initializers.DB(createAppConfig(am.ScanGroupServiceKey))
	auth, _ = authorizer(db)
	scangroupService := scangroup.New(auth)
	if err := scangroupService.Init([]byte(dbstring)); err != nil {
		log.Fatal().Err(err).Msg("failed to init scangroup")
	}
	log.Info().Msg("scangroupService init'd")

	dbstring, db = initializers.DB(createAppConfig(am.OrganizationServiceKey))
	auth, roleManager := authorizer(db)
	orgService := organization.New(roleManager, auth)
	if err := orgService.Init([]byte(dbstring)); err != nil {
		log.Fatal().Err(err).Msg("failed to init address")
	}

	log.Info().Msg("address init'd")
	dbstring, db = initializers.DB(createAppConfig(am.WebDataServiceKey))
	auth, _ = authorizer(db)
	webdataService := webdata.New(auth)
	if err := webdataService.Init([]byte(dbstring)); err != nil {
		log.Fatal().Err(err).Msg("failed to init webdata")
	}
	log.Info().Msg("webdataService service init'd")

	dbstring, db = initializers.DB(createAppConfig(am.EventServiceKey))
	auth, _ = authorizer(db)
	eventService := event.New(auth)
	if err := eventService.Init([]byte(dbstring)); err != nil {
		log.Fatal().Err(err).Msg("failed to init event service")
	}
	log.Info().Msg("eventService init'd")

	state := initializers.State(createAppConfig(am.DispatcherServiceKey))
	dispatcher := dispatcher.New(scangroupService, eventService, addressService, createModules(webdataService, eventService), state)
	if err := dispatcher.Init(nil); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize dispatcher")
	}
	log.Info().Msg("dispatcher service init'd")

	t := &testing.T{}
	db = amtest.InitDB("local", t)
	defer db.Close()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		if deleteOnExit {
			log.Info().Msg("deleting group")
			amtest.DeleteOrg(db, orgName, t)
		} else {
			log.Info().Msg("not deleting group")
		}
		os.Exit(1)
	}()

	userContext, groupID := prepGroup(db, orgService, addressService, scangroupService)
	// Run pipeline
	dispatcher.PushAddresses(ctx, userContext, groupID)
	<-c
}

func prepGroup(db *pgx.ConnPool, orgService am.OrganizationService, addressService am.AddressService, scangroupService am.ScanGroupService) (am.UserContext, int) {
	ctx := context.Background()
	t := &testing.T{}
	var groupID int

	if !createOrg {
		log.Info().Msg("not creating group, returning data")
		orgID := amtest.GetOrgID(db, orgName, t)
		userID := amtest.GetUserId(db, orgID, orgName+"email@email.com", t)
		userContext := amtest.CreateUserContext(orgID, userID)
		_, group, err := scangroupService.GetByName(ctx, userContext, orgName+"group")
		if err != nil {
			log.Fatal().Err(err).Msg("failed to get group by name")
		}
		return userContext, group.GroupID
	}
	log.Info().Msg("Creating group")
	orgID, userID, _, _, err := orgService.Create(ctx, &am.UserContextData{OrgID: 1, UserID: 1}, amtest.CreateOrgInstance(orgName), "asdfasdf")
	userContext := &am.UserContextData{
		TraceID:        "test-fullservice",
		OrgID:          orgID,
		OrgCID:         "abdcdef",
		UserID:         userID,
		UserCID:        "asdfasdf",
		Roles:          []string{"owner"},
		IPAddress:      "192.168.1.1",
		SubscriptionID: 1000,
	}

	_, groupID, err = scangroupService.Create(ctx, userContext, &am.ScanGroup{
		OrgID:                orgID,
		GroupID:              groupID,
		GroupName:            orgName + "group",
		CreationTime:         time.Now().UnixNano(),
		CreatedBy:            orgName + "email@email.com",
		CreatedByID:          userID,
		ModifiedBy:           orgName + "email@email.com",
		ModifiedByID:         userID,
		ModifiedTime:         time.Now().UnixNano(),
		OriginalInputS3URL:   "s3://empty",
		ModuleConfigurations: amtest.CreateModuleConfig(),
		Paused:               false,
		Deleted:              false,
		ArchiveAfterDays:     7,
	})

	if err != nil {
		log.Fatal().Err(err).Msg("failed to get group by name")
	}

	log.Info().Msg("open input file")
	addrFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open addr input file")
	}

	log.Info().Msg("AddrsFromInputFile file")
	addresses := amtest.AddrsFromInputFile(orgID, groupID, addrFile, t)
	addrFile.Close()

	addrs := make(map[string]*am.ScanGroupAddress)
	for _, addr := range addresses {
		addrs[addr.AddressHash] = addr
	}

	log.Info().Int("len", len(addrs)).Msg("adding addresses")
	if _, _, err := addressService.Update(ctx, userContext, addrs); err != nil {
		log.Fatal().Err(err).Msg("failed to add addresses")
	}

	_, cnt, err := addressService.Count(ctx, userContext, groupID)
	if err != nil || cnt == 0 {
		log.Fatal().Err(err).Int("cnt", cnt).Msg("failed to get addresses")
	}
	log.Info().Int("cnt", cnt).Msg("added addresses")
	return userContext, groupID
}

func createModules(webDataService am.WebDataService, eventService am.EventService) map[am.ModuleType]am.ModuleService {
	modules := make(map[am.ModuleType]am.ModuleService, 0)

	state := initializers.State(createAppConfig(am.WebModuleServiceKey))
	dc := dnsclient.New(dnsServers, 3)
	nsService := ns.New(eventService, dc, state)
	nsService.Init(nil)

	state = initializers.State(createAppConfig(am.BruteModuleServiceKey))
	dc = dnsclient.New(dnsServers, 3)
	bruteService := brute.New(dc, state)
	bruteFile, err := os.Open("testdata/100.txt")
	if err != nil {
		log.Fatal().Err(err).Msg("error opening brute sub domain")
	}

	bruteService.Init(bruteFile)

	modules[am.BruteModule] = bruteService
	modules[am.NSModule] = nsService
	modules[am.BigDataCTSubdomainModule] = createBigData()
	modules[am.WebModule] = createWeb(webDataService)
	return modules
}

func createBigData() am.ModuleService {
	var bqConfig bq.ClientConfig
	// configure bigquery details, credentials come from secretscache.
	bqConfig.ProjectID = os.Getenv("APP_BQ_PROJECT_ID")
	bqConfig.DatasetName = os.Getenv("APP_BQ_DATASET_NAME")
	bqConfig.TableName = os.Getenv("APP_BQ_TABLENAME")
	if bqConfig.ProjectID == "" || bqConfig.DatasetName == "" || bqConfig.TableName == "" {
		log.Fatal().Msgf("failed to get bigquery details %v", bqConfig)
	}

	appConfig := createAppConfig(am.BigDataModuleServiceKey)

	systemContext := &am.UserContextData{
		TraceID: "bigdata-system",
		OrgID:   1,
		UserID:  1,
	}

	bqCredentials, err := ioutil.ReadFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err != nil {
		log.Fatal().Err(err).Msg("unable to get bigquery credentials")
	}

	state := initializers.State(appConfig)
	dc := dnsclient.New(dnsServers, 3)

	dbstring, db := initializers.DB(createAppConfig(am.AddressServiceKey))
	auth, _ := authorizer(db)
	bdService := bdservice.New(auth)
	if err := bdService.Init([]byte(dbstring)); err != nil {
		log.Fatal().Err(err).Msg("failed to init bigdata service")
	}

	bqClient := initializers.BigQueryClient(&bqConfig, bqCredentials)

	closeCh := make(chan struct{})
	certListener := initializeCertStream(systemContext, bdService, closeCh)

	service := bigdata.New(dc, state, bdService, bqClient, certListener)
	if err := service.Init(nil); err != nil {
		log.Fatal().Err(err).Msg("failed to make big data module service")
	}
	return service
}

func initializeCertStream(systemContext am.UserContext, bdService am.BigDataService, closeCh chan struct{}) certstream.Listener {
	ctx := context.Background()

	batcher := certstream.NewBatcher(systemContext, bdService, 100)
	if err := batcher.Init(); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize cert stream batcher")
	}

	certListener := certstream.New(batcher)
	if err := certListener.Init(closeCh); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize cert stream listener")
	}

	etlds, _ := bdService.GetETLDs(ctx, systemContext)
	if etlds == nil {
		return certListener
	}
	for _, etld := range etlds {
		certListener.AddETLD(etld.ETLD)
	}

	return certListener
}

func createWeb(webDataService am.WebDataService) am.ModuleService {
	appConfig := createAppConfig(am.WebModuleServiceKey)
	state := initializers.State(appConfig)
	dc := dnsclient.New(dnsServers, 3)

	store := filestorage.NewStorage("dev", "us-east-1")
	if err := store.Init(); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize storage")
	}

	appJSON, err := store.GetInfraFile(context.Background(), "linkai-infra", "dev/web/apps.json")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get apps.json file for detectors from storage")
	}

	wapp := webtech.NewWappalyzer()
	if err := wapp.Init(appJSON); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize webtech detector")
	}

	leaser := browser.NewLocalLeaser()
	ctx := context.Background()
	browsers := browser.NewGCDBrowserPool(5, leaser, wapp)
	if err := browsers.Init(); err != nil {
		log.Fatal().Err(err).Msg("failed initializing browsers")
	}
	defer browsers.Close(ctx)
	webModule := web.New(browsers, webDataService, dc, state, store)
	if err := webModule.Init(); err != nil {
		log.Fatal().Err(err).Msg("failed to init web module service")
	}
	return webModule
}
