package convert

import (
	"github.com/linkai-io/am/am"
	"github.com/wirepair/gcd/gcdapi"
)

// NetworkCertificateToWebCertificate converts from gcdapi to an am.WebCertificate
func NetworkCertificateToWebCertificate(in *gcdapi.NetworkSecurityDetails) *am.WebCertificate {
	return &am.WebCertificate{
		Protocol:                          in.Protocol,
		KeyExchange:                       in.KeyExchange,
		KeyExchangeGroup:                  in.KeyExchangeGroup,
		Cipher:                            in.Cipher,
		Mac:                               in.Mac,
		CertificateId:                     in.CertificateId,
		SubjectName:                       in.SubjectName,
		SanList:                           in.SanList,
		Issuer:                            in.Issuer,
		ValidFrom:                         in.ValidFrom,
		ValidTo:                           in.ValidTo,
		CertificateTransparencyCompliance: in.CertificateTransparencyCompliance,
	}
}
