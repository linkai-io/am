package protoc

import (
	"context"
	"time"

	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/bigdata"

	"github.com/linkai-io/am/am"
)

type BigDataProtocService struct {
	bds am.BigDataService
}

func New(implementation am.BigDataService) *BigDataProtocService {
	return &BigDataProtocService{bds: implementation}
}

func (s *BigDataProtocService) AddCT(ctx context.Context, in *bigdata.AddCTRequest) (*bigdata.CTAddedResponse, error) {
	err := s.bds.AddCT(ctx, convert.UserContextToDomain(in.UserContext), in.ETLD, time.Unix(0, in.QueryTime), convert.CTRecordsToDomain(in.Records))
	if err != nil {
		return nil, err
	}

	return &bigdata.CTAddedResponse{}, nil
}

func (s *BigDataProtocService) GetCT(ctx context.Context, in *bigdata.GetCTRequest) (*bigdata.GetCTResponse, error) {
	ts, records, err := s.bds.GetCT(ctx, convert.UserContextToDomain(in.UserContext), in.ETLD)
	if err != nil {
		return nil, err
	}

	return &bigdata.GetCTResponse{Time: ts.UnixNano(), Records: convert.DomainToCTRecords(records)}, nil
}

func (s *BigDataProtocService) DeleteCT(ctx context.Context, in *bigdata.DeleteCTRequest) (*bigdata.CTDeletedResponse, error) {
	err := s.bds.DeleteCT(ctx, convert.UserContextToDomain(in.UserContext), in.ETLD)
	if err != nil {
		return nil, err
	}

	return &bigdata.CTDeletedResponse{}, nil
}
