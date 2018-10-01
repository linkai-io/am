package module

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	service "github.com/linkai-io/am/protocservices/module"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Config struct {
	Addr       string
	ModuleType am.ModuleType
}

type Client struct {
	client         service.ModuleClient
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 60)}
}

func (c *Client) Init(data []byte) error {
	config, err := c.parseConfig(data)
	if err != nil {
		return err
	}

	balancer := grpc.RoundRobin(grpclb.NewResolver(&grpclb.Options{
		Address: config.Addr,
	}))

	key := am.KeyFromModuleType(config.ModuleType)
	if key == "" {
		return errors.New("unknown module type passed to init")
	}

	conn, err := grpc.Dial(key, grpc.WithInsecure(), grpc.WithBalancer(balancer))
	if err != nil {
		return err
	}
	c.client = service.NewModuleClient(conn)
	return nil
}

func (c *Client) parseConfig(data []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	if config.Addr == "" {
		return nil, errors.New("module did not have Addr set")
	}

	return config, nil
}

func (c *Client) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {
	var err error
	var resp *service.AnalyzedResponse
	in := &service.AnalyzeRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Address:     convert.DomainToAddress(address),
	}

	ctxDeadline, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	err = retrier.Retry(func() error {
		var retryErr error

		resp, retryErr = c.client.Analyze(ctxDeadline, in)
		if retryErr != nil {
			log.Info().Msgf("module analyze returned, cancel? %v", ctxDeadline.Err())
		}
		return retryErr
	})

	if err != nil {
		return nil, nil, err
	}

	addrs := make(map[string]*am.ScanGroupAddress, len(resp.Addresses))
	for key, val := range resp.Addresses {
		addrs[key] = convert.AddressToDomain(val)
	}

	return convert.AddressToDomain(resp.Original), addrs, nil
}
