package am

import (
	"context"
	"time"
)

const (
	BigDataServiceKey = "bigdataservice"
	RNBigData         = "lrn:service:bigdata:feature:bigdata"
)

type CTRecord struct {
	CertificateID      int64  `json:"certificate_id"`
	InsertedTime       int64  `json:"time"`
	ServerName         string `json:"server_name"`
	ServerIndex        int64  `json:"server_index"`
	CertHash           string `json:"cert_hash"`
	SerialNumber       string `json:"serial_number"`
	NotBefore          int64  `json:"not_before"`
	NotAfter           int64  `json:"not_after"`
	Country            string `json:"country"`
	Organization       string `json:"organization"`
	OrganizationalUnit string `json:"organizational_unit"`
	CommonName         string `json:"common_name"`
	VerifiedDNSNames   string `json:"verified_dns_names"`
	UnverifiedDNSNames string `json:"unverified_dns_names"`
	IPAddresses        string `json:"ip_addresses"`
	EmailAddresses     string `json:"email_addresses"`
	ETLD               string `json:"etld"`
}

type SonarData struct {
}

type CommonCrawlData struct {
}

type BigDataService interface {
	DeleteCT(ctx context.Context, userContext UserContext, etld string) error
	GetCT(ctx context.Context, userContext UserContext, etld string) (time.Time, map[string]*CTRecord, error)
	AddCT(ctx context.Context, userContext UserContext, etld string, queryTime time.Time, ctRecords map[string]*CTRecord) error
	//GetSonar(ctx context.Context, userContext UserContext, zone string) ([]*SonarData, error)
	//AddSonar(ctx context.Context, userContext UserContext, sonarData []*SonarData) error
	//GetCommonCrawl(ctx context.Context, userContext UserContext, zone string) (*CommonCrawlData, error)
}
