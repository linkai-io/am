package dispatcher

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"

	service "github.com/linkai-io/am/protocservices/module"
	"google.golang.org/grpc"
)

type Client struct {
	client service.ModuleClient
}

func New() *Client {
	return &Client{}
}

func (c *Client) Init(config []byte) error {
	conn, err := grpc.Dial(string(config), grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.client = service.NewModuleClient(conn)
	return nil
}

func (c *Client) Analyze(ctx context.Context, address *am.ScanGroupAddress) error {
	in := &service.AnalyzeRequest{
		Address: convert.DomainToAddress(address),
	}

	_, err := c.client.Analyze(ctx, in)
	return err
}
