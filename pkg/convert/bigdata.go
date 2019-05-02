package convert

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/protocservices/prototypes"
)

func DomainToCTETLD(in *am.CTETLD) *prototypes.CTETLD {
	return &prototypes.CTETLD{
		EtldId:         in.ETLD_ID,
		Etld:           in.ETLD,
		QueryTimestamp: in.QueryTimestamp,
	}
}

func CTETLDToDomain(in *prototypes.CTETLD) *am.CTETLD {
	return &am.CTETLD{
		ETLD_ID:        in.EtldId,
		ETLD:           in.Etld,
		QueryTimestamp: in.QueryTimestamp,
	}
}

func DomainToCTETLDs(in []*am.CTETLD) []*prototypes.CTETLD {
	if in == nil {
		return make([]*prototypes.CTETLD, 0)
	}

	etlds := make([]*prototypes.CTETLD, len(in))
	for i, v := range in {
		etlds[i] = DomainToCTETLD(v)
	}
	return etlds
}

func CTETLDsToDomain(in []*prototypes.CTETLD) []*am.CTETLD {
	if in == nil {
		return make([]*am.CTETLD, 0)
	}

	etlds := make([]*am.CTETLD, len(in))
	for i, v := range in {
		etlds[i] = CTETLDToDomain(v)
	}
	return etlds
}

func DomainToCTRecord(in *am.CTRecord) *prototypes.CTRecord {
	return &prototypes.CTRecord{
		CertificateID:      in.CertificateID,
		InsertedTime:       in.InsertedTime,
		CertHash:           in.CertHash,
		SerialNumber:       in.SerialNumber,
		NotBefore:          in.NotBefore,
		NotAfter:           in.NotAfter,
		Country:            in.Country,
		Organization:       in.Organization,
		OrganizationalUnit: in.OrganizationalUnit,
		CommonName:         in.CommonName,
		VerifiedDNSNames:   in.VerifiedDNSNames,
		UnverifiedDNSNames: in.UnverifiedDNSNames,
		IPAddresses:        in.IPAddresses,
		EmailAddresses:     in.EmailAddresses,
		ETLD:               in.ETLD,
	}
}

func CTRecordToDomain(in *prototypes.CTRecord) *am.CTRecord {
	return &am.CTRecord{
		CertificateID:      in.CertificateID,
		InsertedTime:       in.InsertedTime,
		CertHash:           in.CertHash,
		SerialNumber:       in.SerialNumber,
		NotBefore:          in.NotBefore,
		NotAfter:           in.NotAfter,
		Country:            in.Country,
		Organization:       in.Organization,
		OrganizationalUnit: in.OrganizationalUnit,
		CommonName:         in.CommonName,
		VerifiedDNSNames:   in.VerifiedDNSNames,
		UnverifiedDNSNames: in.UnverifiedDNSNames,
		IPAddresses:        in.IPAddresses,
		EmailAddresses:     in.EmailAddresses,
		ETLD:               in.ETLD,
	}
}

func DomainToCTRecords(in map[string]*am.CTRecord) map[string]*prototypes.CTRecord {
	ctRecords := make(map[string]*prototypes.CTRecord, len(in))
	for k, v := range in {
		ctRecords[k] = DomainToCTRecord(v)
	}
	return ctRecords
}

func CTRecordsToDomain(in map[string]*prototypes.CTRecord) map[string]*am.CTRecord {
	ctRecords := make(map[string]*am.CTRecord, len(in))
	for k, v := range in {
		ctRecords[k] = CTRecordToDomain(v)
	}
	return ctRecords
}

func DomainToCTSubdomainRecord(in *am.CTSubdomain) *prototypes.CTSubdomain {
	return &prototypes.CTSubdomain{
		SubdomainID:  in.SubdomainID,
		InsertedTime: in.InsertedTime,
		CommonName:   in.Subdomain,
		ETLD:         in.ETLD,
	}
}

func CTSubdomainRecordToDomain(in *prototypes.CTSubdomain) *am.CTSubdomain {
	return &am.CTSubdomain{
		SubdomainID:  in.SubdomainID,
		InsertedTime: in.InsertedTime,
		Subdomain:    in.CommonName,
		ETLD:         in.ETLD,
	}
}

func DomainToCTSubdomainRecords(in map[string]*am.CTSubdomain) map[string]*prototypes.CTSubdomain {
	subRecords := make(map[string]*prototypes.CTSubdomain, len(in))
	for k, v := range in {
		subRecords[k] = DomainToCTSubdomainRecord(v)
	}
	return subRecords
}

func CTSubdomainRecordsToDomain(in map[string]*prototypes.CTSubdomain) map[string]*am.CTSubdomain {
	subRecords := make(map[string]*am.CTSubdomain, len(in))
	for k, v := range in {
		subRecords[k] = CTSubdomainRecordToDomain(v)
	}
	return subRecords
}
