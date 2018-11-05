package bq

import "github.com/linkai-io/am/am"

func CTBigQueryResultToDomain(certHash string, result *Result) *am.CTRecord {
	return &am.CTRecord{
		CertificateID:      0,
		InsertedTime:       result.Time.UnixNano(),
		CertHash:           certHash,
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
