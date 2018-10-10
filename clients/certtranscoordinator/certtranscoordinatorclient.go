package coordinator

import (
	"io"
	"time"

	"github.com/bsm/grpclb"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/rs/zerolog/log"

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

func (c *Client) GetServer(stream service.CertTransCoordinator_GetServerClient) error {
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := stream.Send(&service.GetServerRequest{})
		if err == io.EOF {
			log.Warn().Msg("ct coordinator client got eof from server")
			return nil
		}

		server, err := c.cs.GetServer(ctx)
		if err != nil {
			log.Error().Err(err).Msg("got error calling ct coordinator service")
			continue
		}

		if err := stream.Send(&certtranscoordinator.GetServerResponse{Server: convert.DomainToCTServer(server)}); err != nil {
			log.Error().Err(err).Str("certificate_server", server.URL).Msg("error sending server to client")
		} else {
			server.Index += int64(server.Step)
			server.IndexUpdated = time.Now().UnixNano()
		}

		c.cs.ReturnServer(ctx, server)
	}
}
