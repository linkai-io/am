package protoc

import (
	"context"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/protocservices/ctworker"
	"github.com/rs/zerolog/log"
)

type CTWorkerProtocService struct {
	cs am.CTWorkerService
}

func New(implementation am.CTWorkerService) *CTWorkerProtocService {
	return &CTWorkerProtocService{cs: implementation}
}

func (c *CTWorkerProtocService) GetCTCertificates(ctx context.Context, in *ctworker.GetCTCertificatesRequest) (*ctworker.GetCTCertificatesResponse, error) {
	server, err := c.cs.GetCTCertificates(ctx, convert.CTServerToDomain(in.Server))
	if err != nil {
		log.Error().Err(err).Msg("got error calling ct worker get ct certificates")
		return nil, err
	}

	return &ctworker.GetCTCertificatesResponse{Server: convert.DomainToCTServer(server)}, nil
}

func (c *CTWorkerProtocService) SetExtractors(ctx context.Context, in *ctworker.SetExtractorsRequest) (*ctworker.ExtractorsSetResponse, error) {
	err := c.cs.SetExtractors(ctx, in.NumExtractors)
	if err != nil {
		log.Error().Err(err).Msg("got error calling ct worker set extractors")
		return nil, err
	}

	return &ctworker.ExtractorsSetResponse{}, nil
}
