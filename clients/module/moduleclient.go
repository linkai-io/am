package dispatcher

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"

	service "github.com/linkai-io/am/protocservices/module"
	"google.golang.org/grpc"
)

type Config struct {
	Addr       string
	ModuleType am.ModuleType
}

type Client struct {
	client service.ModuleClient
}

func New() *Client {
	return &Client{}
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

func (c *Client) Analyze(ctx context.Context, address *am.ScanGroupAddress) error {
	in := &service.AnalyzeRequest{
		Address: convert.DomainToAddress(address),
	}

	return retrier.Retry(func() error {
		_, err := c.client.Analyze(ctx, in)
		return err
	})
}
