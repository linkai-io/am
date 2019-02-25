package convert

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/protocservices/prototypes"
)

func DomainToWebCertificate(in *am.WebCertificate) *prototypes.WebCertificate {
	if in == nil {
		return nil
	}

	return &prototypes.WebCertificate{
		OrgID:                             int32(in.OrgID),
		GroupID:                           int32(in.GroupID),
		CertificateID:                     in.CertificateID,
		ResponseTimestamp:                 in.ResponseTimestamp,
		HostAddress:                       in.HostAddress,
		Port:                              in.Port,
		Protocol:                          in.Protocol,
		KeyExchange:                       in.KeyExchange,
		KeyExchangeGroup:                  in.KeyExchangeGroup,
		Cipher:                            in.Cipher,
		Mac:                               in.Mac,
		CertificateValue:                  int32(in.CertificateValue),
		SubjectName:                       in.SubjectName,
		SanList:                           in.SanList,
		Issuer:                            in.Issuer,
		ValidFrom:                         in.ValidFrom,
		ValidTo:                           in.ValidTo,
		CertificateTransparencyCompliance: in.CertificateTransparencyCompliance,
		IsDeleted:                         in.IsDeleted,
		AddressHash:                       in.AddressHash,
		IPAddress:                         in.IPAddress,
	}
}

func WebCertificateToDomain(in *prototypes.WebCertificate) *am.WebCertificate {
	if in == nil {
		return nil
	}

	return &am.WebCertificate{
		OrgID:                             int(in.OrgID),
		GroupID:                           int(in.GroupID),
		CertificateID:                     in.CertificateID,
		ResponseTimestamp:                 in.ResponseTimestamp,
		HostAddress:                       in.HostAddress,
		IPAddress:                         in.IPAddress,
		AddressHash:                       in.AddressHash,
		Port:                              in.Port,
		Protocol:                          in.Protocol,
		KeyExchange:                       in.KeyExchange,
		KeyExchangeGroup:                  in.KeyExchangeGroup,
		Cipher:                            in.Cipher,
		Mac:                               in.Mac,
		CertificateValue:                  int(in.CertificateValue),
		SubjectName:                       in.SubjectName,
		SanList:                           in.SanList,
		Issuer:                            in.Issuer,
		ValidFrom:                         in.ValidFrom,
		ValidTo:                           in.ValidTo,
		CertificateTransparencyCompliance: in.CertificateTransparencyCompliance,
		IsDeleted:                         in.IsDeleted,
	}
}

func DomainToWebCertificates(in []*am.WebCertificate) []*prototypes.WebCertificate {
	if in == nil {
		return nil
	}

	certificateLen := len(in)

	certificates := make([]*prototypes.WebCertificate, certificateLen)
	for i := 0; i < certificateLen; i++ {
		certificates[i] = DomainToWebCertificate(in[i])
	}
	return certificates
}

func WebCertificatesToDomain(in []*prototypes.WebCertificate) []*am.WebCertificate {
	if in == nil {
		return nil
	}

	certificateLen := len(in)

	certificates := make([]*am.WebCertificate, certificateLen)
	for i := 0; i < certificateLen; i++ {
		certificates[i] = WebCertificateToDomain(in[i])
	}
	return certificates
}

func DomainToHTTPResponse(in *am.HTTPResponse) *prototypes.HTTPResponse {
	if in == nil {
		return nil
	}

	return &prototypes.HTTPResponse{
		ResponseID:          in.ResponseID,
		OrgID:               int32(in.OrgID),
		GroupID:             int32(in.GroupID),
		Scheme:              in.Scheme,
		HostAddress:         in.HostAddress,
		IPAddress:           in.IPAddress,
		AddressHash:         in.AddressHash,
		ResponsePort:        in.ResponsePort,
		RequestedPort:       in.RequestedPort,
		Status:              int32(in.Status),
		StatusText:          in.StatusText,
		URL:                 in.URL,
		Headers:             in.Headers,
		MimeType:            in.MimeType,
		RawBodyLink:         in.RawBodyLink,
		RawBodyHash:         in.RawBodyHash,
		ResponseTimestamp:   in.ResponseTimestamp,
		IsDocument:          in.IsDocument,
		WebCertificate:      DomainToWebCertificate(in.WebCertificate),
		IsDeleted:           in.IsDeleted,
		URLRequestTimestamp: in.URLRequestTimestamp,
	}
}

func HTTPResponseToDomain(in *prototypes.HTTPResponse) *am.HTTPResponse {
	if in == nil {
		return nil
	}

	return &am.HTTPResponse{
		ResponseID:          in.ResponseID,
		OrgID:               int(in.OrgID),
		GroupID:             int(in.GroupID),
		Scheme:              in.Scheme,
		AddressHash:         in.AddressHash,
		HostAddress:         in.HostAddress,
		IPAddress:           in.IPAddress,
		ResponsePort:        in.ResponsePort,
		RequestedPort:       in.RequestedPort,
		Status:              int(in.Status),
		StatusText:          in.StatusText,
		URL:                 in.URL,
		Headers:             in.Headers,
		MimeType:            in.MimeType,
		RawBodyLink:         in.RawBodyLink,
		RawBodyHash:         in.RawBodyHash,
		ResponseTimestamp:   in.ResponseTimestamp,
		URLRequestTimestamp: in.URLRequestTimestamp,
		IsDocument:          in.IsDocument,
		WebCertificate:      WebCertificateToDomain(in.WebCertificate),
		IsDeleted:           in.IsDeleted,
	}
}

func DomainToHTTPResponses(in []*am.HTTPResponse) []*prototypes.HTTPResponse {
	if in == nil {
		return nil
	}

	responseLen := len(in)

	responses := make([]*prototypes.HTTPResponse, responseLen)
	for i := 0; i < responseLen; i++ {
		responses[i] = DomainToHTTPResponse(in[i])
	}
	return responses
}

func HTTPResponsesToDomain(in []*prototypes.HTTPResponse) []*am.HTTPResponse {
	if in == nil {
		return nil
	}

	responseLen := len(in)

	responses := make([]*am.HTTPResponse, responseLen)
	for i := 0; i < responseLen; i++ {
		responses[i] = HTTPResponseToDomain(in[i])
	}
	return responses
}

func DomainToWebData(in *am.WebData) *prototypes.WebData {
	if in == nil {
		return nil
	}

	responses := DomainToHTTPResponses(in.Responses)

	return &prototypes.WebData{
		Address:             DomainToAddress(in.Address),
		Responses:           responses,
		SnapshotLink:        in.SnapshotLink,
		SerializedDOMHash:   in.SerializedDOMHash,
		SerializedDOMLink:   in.SerializedDOMLink,
		ResponseTimestamp:   in.ResponseTimestamp,
		URLRequestTimestamp: in.URLRequestTimestamp,
		URL:                 in.URL,
		AddressHash:         in.AddressHash,
		HostAddress:         in.HostAddress,
		IPAddress:           in.IPAddress,
		Scheme:              in.Scheme,
		ResponsePort:        int32(in.ResponsePort),
	}
}

func WebDataToDomain(in *prototypes.WebData) *am.WebData {
	if in == nil {
		return nil
	}

	responses := HTTPResponsesToDomain(in.Responses)

	return &am.WebData{
		Address:             AddressToDomain(in.Address),
		Responses:           responses,
		SnapshotLink:        in.SnapshotLink,
		URL:                 in.URL,
		Scheme:              in.Scheme,
		AddressHash:         in.AddressHash,
		HostAddress:         in.HostAddress,
		IPAddress:           in.IPAddress,
		ResponsePort:        int(in.ResponsePort),
		SerializedDOMHash:   in.SerializedDOMHash,
		SerializedDOMLink:   in.SerializedDOMLink,
		ResponseTimestamp:   in.ResponseTimestamp,
		URLRequestTimestamp: in.URLRequestTimestamp,
	}
}

func DomainToWebSnapshot(in *am.WebSnapshot) *prototypes.WebSnapshot {
	return &prototypes.WebSnapshot{
		OrgID:             int32(in.OrgID),
		GroupID:           int32(in.GroupID),
		SnapshotID:        in.SnapshotID,
		SnapshotLink:      in.SnapshotLink,
		SerializedDOMLink: in.SerializedDOMLink,
		ResponseTimestamp: in.ResponseTimestamp,
		IsDeleted:         in.IsDeleted,
		SerializedDOMHash: in.SerializedDOMHash,
		URL:               in.URL,
		AddressHash:       in.AddressHash,
		HostAddress:       in.HostAddress,
		IPAddress:         in.IPAddress,
		ResponsePort:      int32(in.ResponsePort),
		Scheme:            in.Scheme,
	}
}

func WebSnapshotToDomain(in *prototypes.WebSnapshot) *am.WebSnapshot {
	return &am.WebSnapshot{
		SnapshotID:        in.SnapshotID,
		OrgID:             int(in.OrgID),
		GroupID:           int(in.GroupID),
		SnapshotLink:      in.SnapshotLink,
		SerializedDOMHash: in.SerializedDOMHash,
		SerializedDOMLink: in.SerializedDOMLink,
		ResponseTimestamp: in.ResponseTimestamp,
		IsDeleted:         in.IsDeleted,
		URL:               in.URL,
		AddressHash:       in.AddressHash,
		HostAddress:       in.HostAddress,
		IPAddress:         in.IPAddress,
		ResponsePort:      int(in.ResponsePort),
		Scheme:            in.Scheme,
	}
}

func DomainToWebSnapshots(in []*am.WebSnapshot) []*prototypes.WebSnapshot {
	if in == nil {
		return nil
	}

	snapshotLen := len(in)

	snapshots := make([]*prototypes.WebSnapshot, snapshotLen)
	for i := 0; i < snapshotLen; i++ {
		snapshots[i] = DomainToWebSnapshot(in[i])
	}
	return snapshots
}

func WebSnapshotsToDomain(in []*prototypes.WebSnapshot) []*am.WebSnapshot {
	if in == nil {
		return nil
	}

	snapshotLen := len(in)

	snapshots := make([]*am.WebSnapshot, snapshotLen)
	for i := 0; i < snapshotLen; i++ {
		snapshots[i] = WebSnapshotToDomain(in[i])
	}
	return snapshots
}

func DomainToWebSnapshotFilter(in *am.WebSnapshotFilter) *prototypes.WebSnapshotFilter {
	return &prototypes.WebSnapshotFilter{
		OrgID:   int32(in.OrgID),
		GroupID: int32(in.GroupID),
		Start:   in.Start,
		Limit:   int32(in.Limit),
		Filters: DomainToFilterTypes(in.Filters),
	}
}

func WebSnapshotFilterToDomain(in *prototypes.WebSnapshotFilter) *am.WebSnapshotFilter {
	return &am.WebSnapshotFilter{
		OrgID:   int(in.OrgID),
		GroupID: int(in.GroupID),
		Start:   in.Start,
		Limit:   int(in.Limit),
		Filters: FilterTypesToDomain(in.Filters),
	}
}

func DomainToWebResponseFilter(in *am.WebResponseFilter) *prototypes.WebResponseFilter {
	return &prototypes.WebResponseFilter{
		OrgID:   int32(in.OrgID),
		GroupID: int32(in.GroupID),
		Start:   in.Start,
		Limit:   int32(in.Limit),
		Filters: DomainToFilterTypes(in.Filters),
	}
}

func WebResponseFilterToDomain(in *prototypes.WebResponseFilter) *am.WebResponseFilter {
	return &am.WebResponseFilter{
		OrgID:   int(in.OrgID),
		GroupID: int(in.GroupID),
		Start:   in.Start,
		Limit:   int(in.Limit),
		Filters: FilterTypesToDomain(in.Filters),
	}
}

func DomainToWebCertificateFilter(in *am.WebCertificateFilter) *prototypes.WebCertificateFilter {
	return &prototypes.WebCertificateFilter{
		OrgID:   int32(in.OrgID),
		GroupID: int32(in.GroupID),
		Filters: DomainToFilterTypes(in.Filters),
		Start:   in.Start,
		Limit:   int32(in.Limit),
	}
}

func WebCertificateFilterToDomain(in *prototypes.WebCertificateFilter) *am.WebCertificateFilter {
	return &am.WebCertificateFilter{
		OrgID:   int(in.OrgID),
		GroupID: int(in.GroupID),
		Filters: FilterTypesToDomain(in.Filters),
		Start:   in.Start,
		Limit:   int(in.Limit),
	}
}

func DomainToURLData(in []*am.URLData) []*prototypes.URLData {
	if in == nil {
		return nil
	}
	urls := make([]*prototypes.URLData, len(in))
	for i, inURL := range in {
		urls[i] = &prototypes.URLData{
			ResponseID:  inURL.ResponseID,
			URL:         inURL.URL,
			RawBodyLink: inURL.RawBodyLink,
			MimeType:    inURL.MimeType,
		}
	}
	return urls
}

func DomainToURLListResponse(in *am.URLListResponse) *prototypes.URLListResponse {
	return &prototypes.URLListResponse{
		OrgID:               int32(in.OrgID),
		GroupID:             int32(in.GroupID),
		HostAddress:         in.HostAddress,
		IPAddress:           in.IPAddress,
		URLRequestTimestamp: in.URLRequestTimestamp,
		URLs:                DomainToURLData(in.URLs),
	}
}

func URLDataToDomain(in []*prototypes.URLData) []*am.URLData {
	if in == nil {
		return nil
	}
	urls := make([]*am.URLData, len(in))
	for i, inURL := range in {
		urls[i] = &am.URLData{
			ResponseID:  inURL.ResponseID,
			URL:         inURL.URL,
			RawBodyLink: inURL.RawBodyLink,
			MimeType:    inURL.MimeType,
		}
	}
	return urls
}

func URLListResponseToDomain(in *prototypes.URLListResponse) *am.URLListResponse {
	return &am.URLListResponse{
		OrgID:               int(in.OrgID),
		GroupID:             int(in.GroupID),
		HostAddress:         in.HostAddress,
		IPAddress:           in.IPAddress,
		URLRequestTimestamp: in.URLRequestTimestamp,
		URLs:                URLDataToDomain(in.URLs),
	}
}

func DomainToURLLists(in []*am.URLListResponse) []*prototypes.URLListResponse {
	if in == nil {
		return nil
	}

	urlListLen := len(in)

	urlLists := make([]*prototypes.URLListResponse, urlListLen)
	for i := 0; i < urlListLen; i++ {
		urlLists[i] = DomainToURLListResponse(in[i])
	}
	return urlLists
}

func URLListsToDomain(in []*prototypes.URLListResponse) []*am.URLListResponse {
	if in == nil {
		return nil
	}

	urlListLen := len(in)

	urlLists := make([]*am.URLListResponse, urlListLen)
	for i := 0; i < urlListLen; i++ {
		urlLists[i] = URLListResponseToDomain(in[i])
	}
	return urlLists
}
