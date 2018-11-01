package protoc

import (
	"context"

	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/webdata"

	"github.com/linkai-io/am/am"
)

type WebDataProtocService struct {
	wds am.WebDataService
}

func New(implementation am.WebDataService) *WebDataProtocService {
	return &WebDataProtocService{wds: implementation}
}

func (s *WebDataProtocService) Add(ctx context.Context, in *webdata.AddRequest) (*webdata.AddedResponse, error) {
	oid, err := s.wds.Add(ctx, convert.UserContextToDomain(in.UserContext), convert.WebDataToDomain(in.Data))
	if err != nil {
		return nil, err
	}

	return &webdata.AddedResponse{OrgID: int32(oid)}, nil
}

func (s *WebDataProtocService) GetResponses(ctx context.Context, in *webdata.GetResponsesRequest) (*webdata.GetResponsesResponse, error) {
	oid, responses, err := s.wds.GetResponses(ctx, convert.UserContextToDomain(in.UserContext), convert.WebResponseFilterToDomain(in.Filter))
	if err != nil {
		return nil, err
	}

	return &webdata.GetResponsesResponse{OrgID: int32(oid), Responses: convert.DomainToHTTPResponses(responses)}, nil
}

func (s *WebDataProtocService) GetCertificates(ctx context.Context, in *webdata.GetCertificatesRequest) (*webdata.GetCertificatesResponse, error) {
	oid, certs, err := s.wds.GetCertificates(ctx, convert.UserContextToDomain(in.UserContext), convert.WebCertificateFilterToDomain(in.Filter))
	if err != nil {
		return nil, err
	}

	return &webdata.GetCertificatesResponse{OrgID: int32(oid), Certificates: convert.DomainToWebCertificates(certs)}, nil
}

func (s *WebDataProtocService) GetSnapshots(ctx context.Context, in *webdata.GetSnapshotsRequest) (*webdata.GetSnapshotsResponse, error) {
	oid, snapshots, err := s.wds.GetSnapshots(ctx, convert.UserContextToDomain(in.UserContext), convert.WebSnapshotFilterToDomain(in.Filter))
	if err != nil {
		return nil, err
	}

	return &webdata.GetSnapshotsResponse{OrgID: int32(oid), Snapshots: convert.DomainToWebSnapshots(snapshots)}, nil

}
