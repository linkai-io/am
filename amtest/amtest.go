package amtest

import (
	"fmt"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/pkg/inputlist"
	"github.com/linkai-io/am/pkg/secrets"

	"github.com/linkai-io/am/am"

	uuid "github.com/gofrs/uuid"
	"github.com/jackc/pgx"
)

const (
	CreateOrgStmt = `insert into am.organizations (
		organization_name, organization_custom_id, user_pool_id, identity_pool_id, user_pool_client_id, user_pool_client_secret, user_pool_jwk,
		owner_email, first_name, last_name, phone, country, state_prefecture, street, 
		address1, address2, city, postal_code, creation_time, deleted, status_id, subscription_id,
		limit_tld, limit_tld_reached, limit_hosts, limit_hosts_reached, limit_custom_web_flows, limit_custom_web_flows_reached
	)
	values ($1, $2, $3, $4, $5, $6, $7, 
			$8, $9, $10, $11, $12, $13, $14, 
			$15, $16, $17, $18, $19, $20, $21, $22, 
			$23, $24, $25, $26, $27, $28);`

	CreateUserStmt      = `insert into am.users (organization_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted) values ($1, $2, $3, $4, $5, $6, $7, false)`
	CreateScanGroupStmt = `insert into am.scan_group (organization_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input_s3_url, configuration, paused, deleted) values 
	($1, $2, $3, $4, $5, $6, $7, $8, false, false) returning scan_group_id`
	CreateAddressStmt = `insert into am.scan_group_addresses as sga (
		organization_id, 
		scan_group_id,
		host_address,
		ip_address,
		discovered_timestamp,
		discovery_id,
		last_scanned_timestamp,
		last_seen_timestamp,
		confidence_score,
		user_confidence_score,
		is_soa,
		is_wildcard_zone,
		is_hosted_service,
		ignored,
		found_from,
		ns_record,
		address_hash
	) values ($1, $2, $3, $4, $5, (select discovery_id from am.scan_address_discovered_by where discovered_by=$6), $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17) returning address_id;`

	DeleteOrgStmt  = "select am.delete_org((select organization_id from am.organizations where organization_name=$1))"
	DeleteUserStmt = "delete from am.users where organization_id=(select organization_id from am.organizations where organization_name=$1)"
	GetOrgIDStmt   = "select organization_id from am.organizations where organization_name=$1"
	GetUserIDStmt  = "select user_id from am.users where organization_id=$1 and email=$2"
	GetUserStmt    = `select organization_id, user_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted, agreement_accepted, agreement_accepted_timestamp
						from am.users where organization_id=$1 and email=$2`
)

func GenerateID(t *testing.T) string {
	id, err := uuid.NewV4()
	if err != nil {
		t.Fatalf("error generating ID: %s\n", err)
	}
	return id.String()
}

func GenerateAddrs(orgID, groupID, count int) []*am.ScanGroupAddress {
	addrs := make([]*am.ScanGroupAddress, count)
	for i := 0; i < count; i++ {
		ip := fmt.Sprintf("192.168.0.%d", i)
		addrs[i] = &am.ScanGroupAddress{
			AddressID:           int64(i),
			OrgID:               orgID,
			GroupID:             groupID,
			HostAddress:         "",
			IPAddress:           ip,
			AddressHash:         convert.HashAddress(ip, ""),
			DiscoveryTime:       time.Now().UnixNano(),
			DiscoveredBy:        "input_list",
			LastScannedTime:     0,
			LastSeenTime:        0,
			ConfidenceScore:     100.0,
			UserConfidenceScore: 0.0,
		}
	}
	return addrs
}

func BuildCTRecords(etld string, insertedTS int64, count int) map[string]*am.CTRecord {
	records := make(map[string]*am.CTRecord, count)
	for i := 0; i < count; i++ {
		numStr := strconv.Itoa(i)

		records[numStr] = &am.CTRecord{
			CertificateID:      0,
			InsertedTime:       insertedTS,
			ServerName:         "someserver",
			ServerIndex:        123,
			CertHash:           numStr,
			SerialNumber:       "1234",
			NotBefore:          time.Now().UnixNano(),
			NotAfter:           time.Now().UnixNano(),
			Country:            "JP",
			Organization:       "ORG",
			OrganizationalUnit: "ORG-U",
			CommonName:         etld,
			VerifiedDNSNames:   numStr + "." + etld + " test." + etld,
			UnverifiedDNSNames: "",
			IPAddresses:        "",
			EmailAddresses:     "",
			ETLD:               etld,
		}

	}
	return records
}

func BuildSubdomainsForCT(etld string, count int) map[string]*am.CTSubdomain {
	records := make(map[string]*am.CTSubdomain, count)
	for i := 0; i < count; i++ {
		subdomain := fmt.Sprintf("%d.%s", i, etld)
		records[subdomain] = &am.CTSubdomain{ETLD: etld, Subdomain: subdomain}
	}
	return records
}

func AddrsFromInputFile(orgID, groupID int, addrFile io.Reader, t *testing.T) []*am.ScanGroupAddress {
	in, err := inputlist.ParseList(addrFile, 10000)
	if err != nil {
		for _, r := range err {
			t.Fatalf("parser errors: %v\n", r)
		}
	}
	addrs := make([]*am.ScanGroupAddress, len(in))
	i := 0
	for addr := range in {

		addrs[i] = &am.ScanGroupAddress{
			AddressID:           int64(i),
			OrgID:               orgID,
			GroupID:             groupID,
			DiscoveredBy:        "input_list",
			DiscoveryTime:       time.Now().UnixNano(),
			ConfidenceScore:     100.0,
			UserConfidenceScore: 0.0,
		}

		if inputlist.IsIP(addr) {
			addrs[i].IPAddress = addr
		} else {
			addrs[i].HostAddress = addr
		}
		addrs[i].AddressHash = convert.HashAddress(addrs[i].IPAddress, addrs[i].HostAddress)
		i++
	}
	return addrs
}

func RunAggregates(db *pgx.ConnPool, t *testing.T) {
	aggregates := []string{"am.do_daily_discovered_aggregation", "am.do_daily_seen_aggregation", "am.do_daily_scanned_aggregation",
		"am.do_trihourly_discovered_aggregation", "am.do_trihourly_seen_aggregation", "am.do_trihourly_scanned_aggregation"}

	for _, agg := range aggregates {
		var start int
		var end int
		if err := db.QueryRow(fmt.Sprintf("select * from %s()", agg)).Scan(&start, &end); err != nil {
			t.Fatalf("failed to run aggregation functions: %v\n", err)
		}
		t.Logf("%s - %d %d\n", agg, start, end)
	}

	if _, err := db.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY am.webdata_server_counts_mv"); err != nil {
		t.Fatalf("failed to run aggregation functions")
	}
}

func CreateModuleConfig() *am.ModuleConfiguration {
	m := &am.ModuleConfiguration{}
	customSubNames := []string{"sub1", "sub2"}
	m.BruteModule = &am.BruteModuleConfig{CustomSubNames: customSubNames, RequestsPerSecond: 10, MaxDepth: 2}
	customPorts := []int32{1, 2}
	m.NSModule = &am.NSModuleConfig{RequestsPerSecond: 10}
	m.PortModule = &am.PortModuleConfig{RequestsPerSecond: 10, CustomPorts: customPorts}
	m.WebModule = &am.WebModuleConfig{MaxLinks: 10, TakeScreenShots: true, ExtractJS: true, FingerprintFrameworks: true}
	m.KeywordModule = &am.KeywordModuleConfig{Keywords: []string{"company"}}
	return m
}

func BuildScanGroup(orgID, groupID int) *am.ScanGroup {
	return &am.ScanGroup{
		OrgID:                orgID,
		GroupID:              groupID,
		GroupName:            fmt.Sprintf("testgroup%d", groupID),
		CreationTime:         time.Now().UnixNano(),
		CreatedBy:            "test",
		CreatedByID:          1,
		ModifiedBy:           "test",
		ModifiedByID:         1,
		ModifiedTime:         time.Now().UnixNano(),
		OriginalInputS3URL:   "test",
		ModuleConfigurations: CreateModuleConfig(),
	}
}

func InitDB(env string, t *testing.T) *pgx.ConnPool {
	sec := secrets.NewSecretsCache(env, "")
	dbstring, err := sec.DBString("linkai_admin")
	if err != nil {
		t.Fatalf("unable to get dbstring: %s\n", err)
	}

	conf, err := pgx.ParseConnectionString(dbstring)
	if err != nil {
		t.Fatalf("error parsing connection string")
	}
	p, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: conf})
	if err != nil {
		t.Fatalf("error connecting to db: %s\n", err)
	}

	return p
}

func CreateOrgInstance(orgName string) *am.Organization {
	return &am.Organization{
		OrgCID:                     GenerateID(nil),
		OrgName:                    orgName,
		OwnerEmail:                 orgName + "email@email.com",
		UserPoolID:                 "userpool.blah",
		UserPoolAppClientID:        "userpoolclient.id",
		UserPoolAppClientSecret:    "userpoolclient.secret",
		IdentityPoolID:             "identitypool.blah",
		UserPoolJWK:                "userpool.jwk",
		FirstName:                  "first",
		LastName:                   "last",
		Phone:                      "1-111-111-1111",
		Country:                    "USA",
		StatePrefecture:            "CA",
		Street:                     "1 fake lane",
		Address1:                   "",
		Address2:                   "",
		City:                       "Beverly Hills",
		PostalCode:                 "90210",
		CreationTime:               time.Now().UnixNano(),
		StatusID:                   1000,
		Deleted:                    false,
		SubscriptionID:             1000,
		LimitTLD:                   9999,
		LimitTLDReached:            false,
		LimitHosts:                 9999,
		LimitHostsReached:          false,
		LimitCustomWebFlows:        9999,
		LimitCustomWebFlowsReached: false,
	}
}

func CreateOrg(p *pgx.ConnPool, name string, t *testing.T) {
	_, err := p.Exec(CreateOrgStmt, name, GenerateID(t), "user_pool_id.blah", "userpoolclient.id", "userpoolclient.secret", "identity_pool_id.blah", "user_pool_jwk.blah",
		name+"email@email.com", "first", "last", "1-111-111-1111", "usa", "ca", "1 fake lane",
		"", "", "sf", "90210", time.Now(), false, am.OrgStatusActive, am.SubscriptionEnterprise, 9999, false, 9999, false, 9999, false)

	if err != nil {
		t.Fatalf("error creating organization %s: %s\n", name, err)
	}

	orgID := GetOrgID(p, name, t)

	_, err = p.Exec(CreateUserStmt, orgID, GenerateID(t), name+"email@email.com", "first", "last", am.UserStatusActive, time.Now())
	if err != nil {
		t.Fatalf("error creating user for %s, %s\n", name, err)
	}
}

func CreateSmallOrg(p *pgx.ConnPool, name string, t *testing.T) {
	org := CreateOrgInstance(name)
	org.SubscriptionID = am.SubscriptionMonthlySmall
	org.LimitCustomWebFlows = 1
	org.LimitHosts = 25
	org.LimitTLD = 1
	CreateOrgFromOrg(p, org, t)
}

func CreateMediumOrg(p *pgx.ConnPool, name string, t *testing.T) {
	org := CreateOrgInstance(name)
	org.SubscriptionID = am.SubscriptionMonthlyMedium
	org.LimitCustomWebFlows = 10
	org.LimitHosts = 250
	org.LimitTLD = 3
	CreateOrgFromOrg(p, org, t)
}

func CreateEnterpriseOrg(p *pgx.ConnPool, name string, t *testing.T) {
	org := CreateOrgInstance(name)
	org.SubscriptionID = am.SubscriptionMonthlySmall
	org.LimitCustomWebFlows = 100
	org.LimitHosts = 10000
	org.LimitTLD = 200
	CreateOrgFromOrg(p, org, t)
}

func CreateOrgFromOrg(p *pgx.ConnPool, org *am.Organization, t *testing.T) {
	_, err := p.Exec(CreateOrgStmt, org.OrgName, org.OrgCID, org.UserPoolID, org.IdentityPoolID, org.UserPoolAppClientID, org.UserPoolAppClientSecret, org.UserPoolJWK,
		org.OwnerEmail, org.FirstName, org.LastName, org.Phone, org.Country, org.StatePrefecture, org.Street,
		org.Address1, org.Address2, org.City, org.PostalCode, time.Now(), false, org.StatusID, org.SubscriptionID,
		org.LimitTLD, org.LimitTLDReached, org.LimitHosts, org.LimitHostsReached, org.LimitCustomWebFlows, org.LimitCustomWebFlowsReached)

	if err != nil {
		t.Fatalf("error creating organization %s: %s\n", org.OrgName, err)
	}

	orgID := GetOrgID(p, org.OrgName, t)

	_, err = p.Exec(CreateUserStmt, orgID, GenerateID(t), org.OwnerEmail, org.FirstName, org.LastName, am.UserStatusActive, time.Now())
	if err != nil {
		t.Fatalf("error creating user for %s, %s\n", org.OrgName, err)
	}
}

func DeleteOrg(p *pgx.ConnPool, name string, t *testing.T) {
	p.Exec(DeleteOrgStmt, name)
}

func GetOrgID(p *pgx.ConnPool, name string, t *testing.T) int {
	var orgID int
	err := p.QueryRow(GetOrgIDStmt, name).Scan(&orgID)
	if err != nil {
		t.Fatalf("error finding org id for %s: %s\n", name, err)
	}
	return orgID
}

func GetUserId(p *pgx.ConnPool, orgID int, name string, t *testing.T) int {
	var userID int
	err := p.QueryRow(GetUserIDStmt, orgID, name+"email@email.com").Scan(&userID)
	if err != nil {
		t.Fatalf("error finding user id for %s: %s\n", name, err)
	}
	return userID
}

func GetUser(p *pgx.ConnPool, orgID int, name string, t *testing.T) *am.User {
	user := &am.User{}
	var agreeTime time.Time
	var createTime time.Time
	err := p.QueryRow(GetUserStmt, orgID, name+"email@email.com").Scan(&user.OrgID, &user.UserID, &user.UserCID, &user.UserEmail, &user.FirstName, &user.LastName, &user.StatusID,
		&createTime, &user.Deleted, &user.AgreementAccepted, &agreeTime)

	if err != nil {
		t.Fatalf("error finding user for %s: %s\n", name, err)
	}
	user.AgreementAcceptedTimestamp = agreeTime.UnixNano()
	user.CreationTime = createTime.UnixNano()
	return user
}

func CreateScanGroup(p *pgx.ConnPool, orgName, groupName string, t *testing.T) int {
	var groupID int
	orgID := GetOrgID(p, orgName, t)
	userID := GetUserId(p, orgID, orgName, t)
	//organization_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input, configuration
	err := p.QueryRow(CreateScanGroupStmt, orgID, groupName, time.Now(), userID, time.Now(), userID, "s3://bucket/blah", nil).Scan(&groupID)
	if err != nil {
		t.Fatalf("error creating scan group: %s\n", err)
	}
	return groupID
}

func CreateScanGroupAddress(p *pgx.ConnPool, orgID, groupID int, t *testing.T) *am.ScanGroupAddress {
	host := "example.com"
	ip := "93.184.216.34"
	address := &am.ScanGroupAddress{
		OrgID:               orgID,
		GroupID:             groupID,
		HostAddress:         host,
		IPAddress:           ip,
		AddressHash:         convert.HashAddress(ip, host),
		DiscoveryTime:       time.Now().UnixNano(),
		DiscoveredBy:        "input_list",
		LastScannedTime:     0,
		LastSeenTime:        0,
		ConfidenceScore:     100.0,
		UserConfidenceScore: 0.0,
	}
	id := CreateScanGroupAddressWithAddress(p, address, t)
	address.AddressID = id
	return address
}

func CreateScanGroupAddressWithAddress(p *pgx.ConnPool, a *am.ScanGroupAddress, t *testing.T) int64 {
	var id int64
	err := p.QueryRow(CreateAddressStmt, a.OrgID, a.GroupID, a.HostAddress, a.IPAddress,
		time.Unix(0, a.DiscoveryTime), a.DiscoveredBy, time.Unix(0, a.LastScannedTime), time.Unix(0, a.LastSeenTime), a.ConfidenceScore,
		a.UserConfidenceScore, a.IsSOA, a.IsWildcardZone, a.IsHostedService, a.Ignored, a.FoundFrom,
		a.NSRecord, a.AddressHash).Scan(&id)

	if err != nil {
		if pgxErr, ok := err.(*pgx.PgError); ok {
			t.Fatalf("error creating scan group address: %v\n", pgxErr)
		}
		t.Fatalf("error creaing scan group address:%v\n", err)
	}
	return id
}

func CreateMultiWebData(address *am.ScanGroupAddress, host, ip string) []*am.WebData {
	webData := make([]*am.WebData, 0)
	insertHost := host

	responses := make([]*am.HTTPResponse, 0)
	urlIndex := 0
	groupIdx := 0

	for i := 1; i < 101; i++ {
		headers := make(map[string]string, 0)
		headers["host"] = host
		headers["server"] = fmt.Sprintf("Apache 1.0.%d", i)
		headers["content-type"] = "text/html"

		response := &am.HTTPResponse{
			OrgID:               address.OrgID,
			GroupID:             address.GroupID,
			Scheme:              "http",
			AddressHash:         convert.HashAddress(ip, host),
			HostAddress:         host,
			IPAddress:           ip,
			ResponsePort:        "80",
			RequestedPort:       "80",
			Status:              200,
			StatusText:          "HTTP 200 OK",
			URL:                 fmt.Sprintf("http://%s/%d", host, urlIndex),
			Headers:             headers,
			MimeType:            "text/html",
			RawBody:             "",
			RawBodyLink:         "s3://data/1/1/1/1",
			RawBodyHash:         "1111",
			ResponseTimestamp:   time.Now().UnixNano(),
			URLRequestTimestamp: 0,
			IsDocument:          true,
			WebCertificate: &am.WebCertificate{
				ResponseTimestamp: time.Now().UnixNano(),
				HostAddress:       host,
				IPAddress:         ip,
				AddressHash:       convert.HashAddress(ip, host),
				Port:              "443",
				Protocol:          "h2",
				KeyExchange:       "kex",
				KeyExchangeGroup:  "keg",
				Cipher:            "aes",
				Mac:               "1234",
				CertificateValue:  0,
				SubjectName:       host,
				SanList: []string{
					"www." + insertHost,
					insertHost,
				},
				Issuer:                            "",
				ValidFrom:                         time.Now().Unix(),
				ValidTo:                           time.Now().Add(time.Hour * time.Duration(24*i)).Unix(),
				CertificateTransparencyCompliance: "unknown",
				IsDeleted:                         false,
			},
			IsDeleted: false,
		}
		responses = append(responses, response)
		urlIndex++

		if i%10 == 0 {
			groupIdx++
			data := &am.WebData{
				Address:             address,
				Responses:           responses,
				SnapshotLink:        "s3://snapshot/1",
				URL:                 fmt.Sprintf("http://%s/%d", host, urlIndex),
				Scheme:              "http",
				AddressHash:         convert.HashAddress(ip, host),
				HostAddress:         host,
				IPAddress:           ip,
				ResponsePort:        80,
				RequestedPort:       80,
				SerializedDOMHash:   fmt.Sprintf("1234%d", i),
				SerializedDOMLink:   "s3:/1/2/3/4",
				ResponseTimestamp:   time.Now().UnixNano(),
				URLRequestTimestamp: time.Now().Add(time.Hour * -time.Duration(groupIdx*24)).UnixNano(),
				DetectedTech: map[string]*am.WebTech{"3dCart": &am.WebTech{
					Matched:  "1.1.11,1.1.11",
					Version:  "1.1.11",
					Location: "headers",
				},
					"jQuery": &am.WebTech{
						Matched:  "1.1.11,1.1.11",
						Version:  "1.1.11",
						Location: "script",
					},
				},
			}
			urlIndex = 0
			webData = append(webData, data)

			insertHost = fmt.Sprintf("%d.%s", i, host)
			responses = make([]*am.HTTPResponse, 0)
		}
	}

	return webData
}

func CreateMultiWebDataWithSub(address *am.ScanGroupAddress, host, ip string, total int) []*am.WebData {
	webData := make([]*am.WebData, 0)
	insertHost := host

	responses := make([]*am.HTTPResponse, 0)
	urlIndex := 0
	groupIdx := 0

	for i := 1; i < total+1; i++ {
		headers := make(map[string]string, 0)
		headers["host"] = host
		headers["server"] = fmt.Sprintf("Apache 1.0.%d", i)
		headers["content-type"] = "text/html"

		response := &am.HTTPResponse{
			OrgID:               address.OrgID,
			GroupID:             address.GroupID,
			Scheme:              "http",
			AddressHash:         convert.HashAddress(ip, host),
			HostAddress:         host,
			IPAddress:           ip,
			ResponsePort:        "80",
			RequestedPort:       "80",
			Status:              200,
			StatusText:          "HTTP 200 OK",
			URL:                 fmt.Sprintf("http://%s/%d", host, urlIndex),
			Headers:             headers,
			MimeType:            "text/html",
			RawBody:             "",
			RawBodyLink:         "s3://data/1/1/1/1",
			RawBodyHash:         "1111",
			ResponseTimestamp:   time.Now().UnixNano(),
			URLRequestTimestamp: 0,
			IsDocument:          true,
			WebCertificate: &am.WebCertificate{
				ResponseTimestamp: time.Now().UnixNano(),
				HostAddress:       host,
				IPAddress:         ip,
				AddressHash:       convert.HashAddress(ip, host),
				Port:              "443",
				Protocol:          "h2",
				KeyExchange:       "kex",
				KeyExchangeGroup:  "keg",
				Cipher:            "aes",
				Mac:               "1234",
				CertificateValue:  0,
				SubjectName:       host,
				SanList: []string{
					"www." + insertHost,
					insertHost,
				},
				Issuer:                            "",
				ValidFrom:                         time.Now().Unix(),
				ValidTo:                           time.Now().Add(time.Hour * time.Duration(24*i)).Unix(),
				CertificateTransparencyCompliance: "unknown",
				IsDeleted:                         false,
			},
			IsDeleted: false,
		}
		if i != 0 {
			response.HostAddress = fmt.Sprintf("%d.%s", i, host)
		}
		responses = append(responses, response)
		urlIndex++

		if i%10 == 0 {
			groupIdx++
			data := &am.WebData{
				Address:             address,
				Responses:           responses,
				SnapshotLink:        "s3://snapshot/1",
				URL:                 fmt.Sprintf("http://%s/%d", host, urlIndex),
				Scheme:              "http",
				AddressHash:         convert.HashAddress(ip, host),
				HostAddress:         host,
				IPAddress:           ip,
				ResponsePort:        80,
				RequestedPort:       80,
				SerializedDOMHash:   fmt.Sprintf("1234%d", i),
				SerializedDOMLink:   "s3:/1/2/3/4",
				ResponseTimestamp:   time.Now().UnixNano(),
				URLRequestTimestamp: time.Now().Add(time.Hour * -time.Duration(groupIdx*24)).UnixNano(),
				DetectedTech: map[string]*am.WebTech{"3dCart": &am.WebTech{
					Matched:  "1.1.11,1.1.11",
					Version:  "1.1.11",
					Location: "headers",
				},
					"jQuery": &am.WebTech{
						Matched:  "1.1.11,1.1.11",
						Version:  "1.1.11",
						Location: "script",
					},
				},
			}
			urlIndex = 0
			webData = append(webData, data)

			insertHost = fmt.Sprintf("%d.%s", i, host)
			responses = make([]*am.HTTPResponse, 0)
		}
	}

	return webData
}

func CreateWebData(address *am.ScanGroupAddress, host, ip string) *am.WebData {
	headers := make(map[string]string, 0)
	headers["host"] = host
	headers["content-type"] = "text/html"
	now := time.Now().UnixNano()
	response := &am.HTTPResponse{
		Scheme:              "http",
		AddressHash:         convert.HashAddress(ip, host),
		HostAddress:         host,
		IPAddress:           ip,
		ResponsePort:        "80",
		RequestedPort:       "80",
		Status:              200,
		StatusText:          "HTTP 200 OK",
		URL:                 fmt.Sprintf("http://%s/", host),
		Headers:             headers,
		MimeType:            "text/html",
		RawBody:             "",
		RawBodyLink:         "s3://data/1/1/1/1",
		RawBodyHash:         "1111",
		ResponseTimestamp:   now + 100000,
		URLRequestTimestamp: now,
		IsDocument:          true,
		WebCertificate: &am.WebCertificate{
			ResponseTimestamp: now,
			HostAddress:       host,
			IPAddress:         ip,
			AddressHash:       convert.HashAddress(ip, host),
			Port:              "443",
			Protocol:          "h2",
			KeyExchange:       "kex",
			KeyExchangeGroup:  "keg",
			Cipher:            "aes",
			Mac:               "1234",
			CertificateValue:  0,
			SubjectName:       host,
			SanList: []string{
				"www." + host,
				host,
			},
			Issuer:                            "",
			ValidFrom:                         time.Now().Unix(),
			ValidTo:                           time.Now().Add(time.Hour * time.Duration(24)).Unix(),
			CertificateTransparencyCompliance: "unknown",
			IsDeleted:                         false,
		},
		IsDeleted: false,
	}
	responses := make([]*am.HTTPResponse, 1)
	responses[0] = response

	webData := &am.WebData{
		Address:             address,
		Responses:           responses,
		Snapshot:            "",
		SnapshotLink:        "s3://snapshot/1",
		URL:                 fmt.Sprintf("http://%s/", host),
		Scheme:              "http",
		AddressHash:         convert.HashAddress(ip, host),
		HostAddress:         host,
		IPAddress:           ip,
		ResponsePort:        80,
		RequestedPort:       80,
		SerializedDOMHash:   "1234",
		SerializedDOMLink:   "s3:/1/2/3/4",
		ResponseTimestamp:   time.Now().UnixNano() + (100000),
		URLRequestTimestamp: now,
		LoadURL:             fmt.Sprintf("http://%s/", host),
		DetectedTech: map[string]*am.WebTech{"3dCart": &am.WebTech{
			Matched:  "1.1.11,1.1.11",
			Version:  "1.1.11",
			Location: "headers",
		},
			"jQuery": &am.WebTech{
				Matched:  "1.1.11,1.1.11",
				Version:  "1.1.11",
				Location: "script",
			},
		},
	}

	return webData
}
