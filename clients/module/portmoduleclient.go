package module

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	service "github.com/linkai-io/am/protocservices/module"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

type PortConfig struct {
	ModuleType am.ModuleType
	Timeout    int
}

type PortClient struct {
	client         service.PortModuleClient
	conn           *grpc.ClientConn
	defaultTimeout time.Duration
	config         *Config
	key            string
}

func NewPortClient() *PortClient {
	return &PortClient{defaultTimeout: (time.Second * 60)}
}

func (c *PortClient) Init(data []byte) error {
	var err error
	c.config, err = c.parseConfig(data)
	if err != nil {
		return err
	}

	if c.config.Timeout != 0 {
		c.defaultTimeout = (time.Second * time.Duration(c.config.Timeout))
	}

	c.key = am.KeyFromModuleType(c.config.ModuleType)
	if c.key == "" {
		return errors.New("unknown module type passed to init")
	}

	conn, err := grpc.DialContext(context.Background(), "srv://consul/"+c.key, grpc.WithInsecure(), grpc.WithBalancerName(roundrobin.Name))
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = service.NewPortModuleClient(conn)
	return nil
}

func (c *PortClient) SetTimeout(timeout time.Duration) {
	c.defaultTimeout = timeout
}

func (c *PortClient) parseConfig(data []byte) (*Config, error) {
	config := &Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *PortClient) AnalyzeWithPorts(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress, portResults *am.PortResults) (*am.ScanGroupAddress, map[string]*am.ScanGroupAddress, *am.Bag, error) {
	var err error
	var resp *service.AnalyzedWithPortsResponse
	in := &service.AnalyzeWithPortsRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Address:     convert.DomainToAddress(address),
		Ports:       convert.DomainToPortResults(portResults),
	}

	ctxDeadline, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.AnalyzeWithPorts(ctxDeadline, in)
		if retryErr != nil {
			log.Warn().Str("client", c.key).Err(retryErr).Msg("module analyze returned error")
		}
		return retryErr
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return nil, nil, nil, err
	}

	addrs := make(map[string]*am.ScanGroupAddress, len(resp.Addresses))
	for key, val := range resp.Addresses {
		addrs[key] = convert.AddressToDomain(val)
	}

	return convert.AddressToDomain(resp.Original), addrs, convert.BagsToDomain(resp.Results), nil
}
