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

func DomainToHTTPResponse(in *am.HTTPResponse) *prototypes.HTTPResponse {
	return &prototypes.HTTPResponse{
		ResponseID:        in.ResponseID,
		OrgID:             int32(in.OrgID),
		GroupID:           int32(in.GroupID),
		AddressID:         in.AddressID,
		Scheme:            in.Scheme,
		HostAddress:       in.HostAddress,
		IPAddress:         in.IPAddress,
		ResponsePort:      in.ResponsePort,
		RequestedPort:     in.RequestedPort,
		Status:            int32(in.Status),
		StatusText:        in.StatusText,
		URL:               in.URL,
		Headers:           in.Headers,
		MimeType:          in.MimeType,
		RawBodyLink:       in.RawBodyLink,
		RawBodyHash:       in.RawBodyHash,
		ResponseTimestamp: in.ResponseTimestamp,
		IsDocument:        in.IsDocument,
		WebCertificate:    DomainToWebCertificate(in.WebCertificate),
		IsDeleted:         in.IsDeleted,
	}
}

func HTTPResponseToDomain(in *prototypes.HTTPResponse) *am.HTTPResponse {
	if in == nil {
		return nil
	}
	return &am.HTTPResponse{
		ResponseID:        in.ResponseID,
		OrgID:             int(in.OrgID),
		GroupID:           int(in.GroupID),
		AddressID:         in.AddressID,
		Scheme:            in.Scheme,
		HostAddress:       in.HostAddress,
		IPAddress:         in.IPAddress,
		ResponsePort:      in.ResponsePort,
		RequestedPort:     in.RequestedPort,
		Status:            int(in.Status),
		StatusText:        in.StatusText,
		URL:               in.URL,
		Headers:           in.Headers,
		MimeType:          in.MimeType,
		RawBodyLink:       in.RawBodyLink,
		RawBodyHash:       in.RawBodyHash,
		ResponseTimestamp: in.ResponseTimestamp,
		IsDocument:        in.IsDocument,
		WebCertificate:    WebCertificateToDomain(in.WebCertificate),
		IsDeleted:         in.IsDeleted,
	}
}

func DomainToWebData(in *am.WebData) *prototypes.WebData {
	if in == nil {
		return nil
	}

	responseLen := 0
	if in.Responses != nil {
		responseLen = len(in.Responses)
	}

	responses := make([]*prototypes.HTTPResponse, responseLen)
	for i := 0; i < responseLen; i++ {
		responses[i] = DomainToHTTPResponse(in.Responses[i])
	}

	return &prototypes.WebData{
		Address:           DomainToAddress(in.Address),
		Responses:         responses,
		SnapshotLink:      in.SnapshotLink,
		SerializedDOMHash: in.SerializedDOMHash,
		SerializedDOMLink: in.SerializedDOMLink,
		ResponseTimestamp: in.ResponseTimestamp,
	}
}

func WebDataToDomain(in *prototypes.WebData) *am.WebData {
	if in == nil {
		return nil
	}

	responseLen := 0
	if in.Responses != nil {
		responseLen = len(in.Responses)
	}

	responses := make([]*am.HTTPResponse, responseLen)
	for i := 0; i < responseLen; i++ {
		responses[i] = HTTPResponseToDomain(in.Responses[i])
	}

	return &am.WebData{
		Address:           AddressToDomain(in.Address),
		Responses:         responses,
		SnapshotLink:      in.SnapshotLink,
		SerializedDOMHash: in.SerializedDOMHash,
		SerializedDOMLink: in.SerializedDOMLink,
		ResponseTimestamp: in.ResponseTimestamp,
	}
}

func DomainToWebSnapshot(in *am.WebSnapshot) *prototypes.WebSnapshot {
	return &prototypes.WebSnapshot{
		OrgID:             int32(in.OrgID),
		GroupID:           int32(in.GroupID),
		AddressID:         in.AddressID,
		SnapshotID:        in.SnapshotID,
		SnapshotLink:      in.SnapshotLink,
		SerializedDOMLink: in.SerializedDOMLink,
		ResponseTimestamp: in.ResponseTimestamp,
		IsDeleted:         in.IsDeleted,
	}
}

func WebSnapshotToDomain(in *prototypes.WebSnapshot) *am.WebSnapshot {
	return &am.WebSnapshot{
		OrgID:             int(in.OrgID),
		GroupID:           int(in.GroupID),
		AddressID:         in.AddressID,
		SnapshotID:        in.SnapshotID,
		SnapshotLink:      in.SnapshotLink,
		SerializedDOMLink: in.SerializedDOMLink,
		ResponseTimestamp: in.ResponseTimestamp,
		IsDeleted:         in.IsDeleted,
	}
}

func DomainToWebSnapshotFilter(in *am.WebSnapshotFilter) *prototypes.WebSnapshotFilter {
	return &prototypes.WebSnapshotFilter{
		OrgID:             int32(in.OrgID),
		GroupID:           int32(in.GroupID),
		WithResponseTime:  in.WithResponseTime,
		SinceResponseTime: in.SinceResponseTime,
		Start:             in.Start,
		Limit:             int32(in.Limit),
	}
}

func WebSnapshotFilterToDomain(in *prototypes.WebSnapshotFilter) *am.WebSnapshotFilter {
	return &am.WebSnapshotFilter{
		OrgID:             int(in.OrgID),
		GroupID:           int(in.GroupID),
		WithResponseTime:  in.WithResponseTime,
		SinceResponseTime: in.SinceResponseTime,
		Start:             in.Start,
		Limit:             int(in.Limit),
	}
}

func DomainToWebResponseFilter(in *am.WebResponseFilter) *prototypes.WebResponseFilter {
	return &prototypes.WebResponseFilter{
		OrgID:             int32(in.OrgID),
		GroupID:           int32(in.GroupID),
		WithResponseTime:  in.WithResponseTime,
		SinceResponseTime: in.SinceResponseTime,
		Start:             in.Start,
		Limit:             int32(in.Limit),
	}
}

func WebResponseFilterToDomain(in *prototypes.WebResponseFilter) *am.WebResponseFilter {
	return &am.WebResponseFilter{
		OrgID:             int(in.OrgID),
		GroupID:           int(in.GroupID),
		WithResponseTime:  in.WithResponseTime,
		SinceResponseTime: in.SinceResponseTime,
		Start:             in.Start,
		Limit:             int(in.Limit),
	}
}

func DomainToWebCertificateFilter(in *am.WebCertificateFilter) *prototypes.WebCertificateFilter {
	return &prototypes.WebCertificateFilter{
		OrgID:             int32(in.OrgID),
		GroupID:           int32(in.GroupID),
		WithResponseTime:  in.WithResponseTime,
		SinceResponseTime: in.SinceResponseTime,
		WithValidTo:       in.WithValidTo,
		ValidToTime:       in.ValidToTime,
		Start:             in.Start,
		Limit:             int32(in.Limit),
	}
}

func WebCertificateFilterToDomain(in *prototypes.WebCertificateFilter) *am.WebCertificateFilter {
	return &am.WebCertificateFilter{
		OrgID:             int(in.OrgID),
		GroupID:           int(in.GroupID),
		WithResponseTime:  in.WithResponseTime,
		SinceResponseTime: in.SinceResponseTime,
		WithValidTo:       in.WithValidTo,
		ValidToTime:       in.ValidToTime,
		Start:             in.Start,
		Limit:             int(in.Limit),
	}
}
