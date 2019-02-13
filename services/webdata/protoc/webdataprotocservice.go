package protoc

import (
	"context"

	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/metrics/load"
	"github.com/linkai-io/am/protocservices/webdata"

	"github.com/linkai-io/am/am"
)

type WebDataProtocService struct {
	wds      am.WebDataService
	reporter *load.RateReporter
}

func New(implementation am.WebDataService, reporter *load.RateReporter) *WebDataProtocService {
	return &WebDataProtocService{wds: implementation, reporter: reporter}
}

func (s *WebDataProtocService) Add(ctx context.Context, in *webdata.AddRequest) (*webdata.AddedResponse, error) {
	s.reporter.Increment(1)
	oid, err := s.wds.Add(ctx, convert.UserContextToDomain(in.UserContext), convert.WebDataToDomain(in.Data))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}

	return &webdata.AddedResponse{OrgID: int32(oid)}, nil
}

func (s *WebDataProtocService) GetResponses(ctx context.Context, in *webdata.GetResponsesRequest) (*webdata.GetResponsesResponse, error) {
	s.reporter.Increment(1)
	oid, responses, err := s.wds.GetResponses(ctx, convert.UserContextToDomain(in.UserContext), convert.WebResponseFilterToDomain(in.Filter))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}

	return &webdata.GetResponsesResponse{OrgID: int32(oid), Responses: convert.DomainToHTTPResponses(responses)}, nil
}

func (s *WebDataProtocService) GetCertificates(ctx context.Context, in *webdata.GetCertificatesRequest) (*webdata.GetCertificatesResponse, error) {
	s.reporter.Increment(1)
	oid, certs, err := s.wds.GetCertificates(ctx, convert.UserContextToDomain(in.UserContext), convert.WebCertificateFilterToDomain(in.Filter))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}

	return &webdata.GetCertificatesResponse{OrgID: int32(oid), Certificates: convert.DomainToWebCertificates(certs)}, nil
}

func (s *WebDataProtocService) GetSnapshots(ctx context.Context, in *webdata.GetSnapshotsRequest) (*webdata.GetSnapshotsResponse, error) {
	s.reporter.Increment(1)
	oid, snapshots, err := s.wds.GetSnapshots(ctx, convert.UserContextToDomain(in.UserContext), convert.WebSnapshotFilterToDomain(in.Filter))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}

	return &webdata.GetSnapshotsResponse{OrgID: int32(oid), Snapshots: convert.DomainToWebSnapshots(snapshots)}, nil
}

func (s *WebDataProtocService) GetURLList(ctx context.Context, in *webdata.GetURLListRequest) (*webdata.GetURLListResponse, error) {
	s.reporter.Increment(1)
	oid, urls, err := s.wds.GetURLList(ctx, convert.UserContextToDomain(in.UserContext), convert.WebResponseFilterToDomain(in.Filter))
	s.reporter.Increment(-1)
	if err != nil {
		return nil, err
	}

	return &webdata.GetURLListResponse{OrgID: int32(oid), URLList: convert.DomainToURLLists(urls)}, nil
}
