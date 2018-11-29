package convert

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/bq"
)

func CTBigQueryResultToDomain(certHash string, result *bq.Result) *am.CTRecord {
	return &am.CTRecord{
		CertificateID:      0,
		InsertedTime:       result.Time.UnixNano(),
		CertHash:           certHash,
		ServerName:         result.Server,
		ServerIndex:        result.Index,
		SerialNumber:       result.SerialNumber,
		NotBefore:          result.NotBefore.UnixNano(),
		NotAfter:           result.NotAfter.UnixNano(),
		Country:            result.Country,
		Organization:       result.Organization,
		OrganizationalUnit: result.OrganizationalUnit,
		CommonName:         result.CommonName,
		VerifiedDNSNames:   result.VerifiedDNSNames,
		UnverifiedDNSNames: result.UnverifiedDNSNames,
		IPAddresses:        result.IPAddresses,
		EmailAddresses:     result.EmailAddresses,
		ETLD:               result.IPAddresses,
	}
}
