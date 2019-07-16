package module

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/retrier"
	service "github.com/linkai-io/am/protocservices/module/portscan"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type tokenAuth struct {
	token string
}

func (t tokenAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (tokenAuth) RequireTransportSecurity() bool {
	return true
}

type PortScanConfig struct {
	Token         string
	ModuleType    am.ModuleType
	ServerAddress string
	Timeout       int
}

type PortScanClient struct {
	client         service.PortScanModuleClient
	conn           *grpc.ClientConn
	defaultTimeout time.Duration
	config         *PortScanConfig
	key            string
}

func NewPortScanClient() *PortScanClient {
	return &PortScanClient{defaultTimeout: (time.Minute * 30)}
}

func (c *PortScanClient) Init(data []byte) error {
	var err error
	c.config, err = c.parseConfig(data)
	if err != nil {
		return err
	}

	if c.config.Timeout != 0 {
		c.defaultTimeout = (time.Second * time.Duration(c.config.Timeout))
	}

	if c.config.ServerAddress == "" {
		return errors.New("server address required in configuration")
	}

	conn, err := grpc.DialContext(context.Background(),
		fmt.Sprintf("dns:///%s", c.config.ServerAddress),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		grpc.WithPerRPCCredentials(tokenAuth{token: c.config.Token}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time.Second * 20,
			PermitWithoutStream: true,
		}))

	if err != nil {
		return err
	}

	c.conn = conn
	c.client = service.NewPortScanModuleClient(conn)
	return nil
}

func (c *PortScanClient) SetTimeout(timeout time.Duration) {
	c.defaultTimeout = timeout
}

func (c *PortScanClient) parseConfig(data []byte) (*PortScanConfig, error) {
	config := &PortScanConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *PortScanClient) AddGroup(ctx context.Context, userContext am.UserContext, group *am.ScanGroup) error {
	var err error
	var resp *service.PortScanGroupAddedResponse
	in := &service.PortScanAddGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Group:       convert.DomainToScanGroup(group),
	}

	ctxDeadline, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.AddGroup(ctxDeadline, in)
		if retryErr != nil {
			log.Warn().Str("client", c.key).Err(retryErr).Msg("portscan module analyze returned error")
		}
		return retryErr
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return err
	}
	resp.Size() // to avoid compiler error but not assign to empty in case fields are added
	return nil
}

func (c *PortScanClient) RemoveGroup(ctx context.Context, userContext am.UserContext, orgID, groupID int) error {
	var err error
	var resp *service.PortScanGroupRemovedResponse
	in := &service.PortScanRemoveGroupRequest{
		UserContext: convert.DomainToUserContext(userContext),
		OrgID:       int32(orgID),
		GroupID:     int32(groupID),
	}

	ctxDeadline, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.RemoveGroup(ctxDeadline, in)
		if retryErr != nil {
			log.Warn().Str("client", c.key).Err(retryErr).Msg("portscan module analyze returned error")
		}
		return retryErr
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return err
	}
	resp.Size() // to avoid compiler error but not assign to empty in case fields are added
	return nil
}

func (c *PortScanClient) Analyze(ctx context.Context, userContext am.UserContext, address *am.ScanGroupAddress) (*am.ScanGroupAddress, *am.PortResults, error) {
	var err error
	var resp *service.PortScanAnalyzedResponse
	in := &service.PortScanAnalyzeRequest{
		UserContext: convert.DomainToUserContext(userContext),
		Address:     convert.DomainToAddress(address),
	}

	ctxDeadline, cancel := context.WithTimeout(context.Background(), c.defaultTimeout)
	defer cancel()

	err = retrier.RetryIfNot(func() error {
		var retryErr error

		resp, retryErr = c.client.Analyze(ctxDeadline, in)
		if retryErr != nil {
			log.Warn().Str("client", c.key).Err(retryErr).Msg("portscan module analyze returned error")
		}
		return retryErr
	}, "rpc error: code = Unavailable desc")

	if err != nil {
		return nil, nil, err
	}

	returnedAddr := convert.AddressToDomain(resp.Address)
	portResults := convert.PortResultsToDomain(resp.PortResult)

	return returnedAddr, portResults, nil
}
