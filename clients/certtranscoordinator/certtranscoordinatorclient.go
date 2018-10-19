package coordinator

import (
	"time"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/am"

	service "github.com/linkai-io/am/protocservices/certtranscoordinator"
	"google.golang.org/grpc"
)

type Client struct {
	client         service.CertTransCoordinatorClient
	defaultTimeout time.Duration
}

func New() *Client {
	return &Client{defaultTimeout: (time.Second * 10)}
}

func (c *Client) Init(config []byte) error {
	balancer := grpc.RoundRobin(grpclb.NewResolver(&grpclb.Options{
		Address: string(config),
	}))

	conn, err := grpc.Dial(am.CoordinatorServiceKey, grpc.WithInsecure(), grpc.WithBalancer(balancer))
	if err != nil {
		return err
	}

	c.client = service.NewCertTransCoordinatorClient(conn)
	return nil
}
