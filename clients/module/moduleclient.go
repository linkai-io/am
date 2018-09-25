package module

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bsm/grpclb"
	balancerpb "github.com/bsm/grpclb/grpclb_balancer_v1"
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
	go debug(key, config.Addr)
	c.client = service.NewModuleClient(conn)
	return nil
}

func debug(key, addr string) {
	for {
		time.Sleep(5 * time.Second)
		cc, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Printf("error dialing address: %v\n", err)
			continue
		}
		defer cc.Close()

		bc := balancerpb.NewLoadBalancerClient(cc)
		resp, err := bc.Servers(context.Background(), &balancerpb.ServersRequest{
			Target: key,
		})
		if err != nil {
			log.Printf("error in resp: %v\n", err)
			continue
		}

		if len(resp.Servers) == 0 {
			fmt.Printf("no %s servers\n", key)
		}
	}
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

func (c *Client) Analyze(ctx context.Context, address *am.ScanGroupAddress) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, error) {
	var err error
	var resp *service.AnalyzedResponse
	in := &service.AnalyzeRequest{
		Address: convert.DomainToAddress(address),
	}

	err = retrier.Retry(func() error {
		resp, err = c.client.Analyze(ctx, in)
		return err
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
