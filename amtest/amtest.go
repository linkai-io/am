package amtest

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/linkai-io/am/pkg/inputlist"
	"github.com/linkai-io/am/pkg/secrets"

	"github.com/linkai-io/am/am"

	uuid "github.com/gofrs/uuid"
	"github.com/jackc/pgx"
	"github.com/linkai-io/am/mock"
)

const (
	CreateOrgStmt = `insert into am.organizations (
		organization_name, organization_custom_id, user_pool_id, identity_pool_id, 
		owner_email, first_name, last_name, phone, country, state_prefecture, street, 
		address1, address2, city, postal_code, creation_time, deleted, status_id, subscription_id
	)
	values 
		($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, false, 1000, 1000);`

	CreateUserStmt      = `insert into am.users (organization_id, user_custom_id, email, first_name, last_name, user_status_id, creation_time, deleted) values ($1, $2, $3, $4, $5, $6, $7, false)`
	CreateScanGroupStmt = `insert into am.scan_group (organization_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input_s3_url, configuration, paused, deleted) values 
	($1, $2, $3, $4, $5, $6, $7, $8, false, false) returning scan_group_id`
	DeleteOrgStmt  = "select am.delete_org((select organization_id from am.organizations where organization_name=$1))"
	DeleteUserStmt = "delete from am.users where organization_id=(select organization_id from am.organizations where organization_name=$1)"
	GetOrgIDStmt   = "select organization_id from am.organizations where organization_name=$1"
	GetUserIDStmt  = "select user_id from am.users where organization_id=$1 and email=$2"
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
		addrs[i] = &am.ScanGroupAddress{
			AddressID:           int64(i),
			OrgID:               orgID,
			GroupID:             groupID,
			HostAddress:         "",
			IPAddress:           fmt.Sprintf("192.168.0.%d", i),
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

func AddrsFromInputFile(orgID, groupID int, addrFile *os.File, t *testing.T) []*am.ScanGroupAddress {
	in, _ := inputlist.ParseList(addrFile, 10000)
	addrFile.Close()

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
		i++
	}
	return addrs
}

func CreateUserContext(orgID, userID int) *mock.UserContext {
	userContext := &mock.UserContext{}
	userContext.GetOrgIDFn = func() int {
		return orgID
	}

	userContext.GetUserIDFn = func() int {
		return userID
	}

	return userContext
}

func MockAuthorizer() *mock.Authorizer {
	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}
	return auth
}

func CreateModuleConfig() *am.ModuleConfiguration {
	m := &am.ModuleConfiguration{}
	customSubNames := []string{"sub1", "sub2"}
	m.BruteModule = &am.BruteModuleConfig{CustomSubNames: customSubNames, RequestsPerSecond: 50, MaxDepth: 2}
	customPorts := []int32{1, 2}
	m.NSModule = &am.NSModuleConfig{RequestsPerSecond: 50}
	m.PortModule = &am.PortModuleConfig{RequestsPerSecond: 50, CustomPorts: customPorts}
	m.WebModule = &am.WebModuleConfig{MaxLinks: 10, TakeScreenShots: true, ExtractJS: true, FingerprintFrameworks: true}
	m.KeywordModule = &am.KeywordModuleConfig{Keywords: []string{"company"}}
	return m
}

func MockRoleManager() *mock.RoleManager {
	roleManager := &mock.RoleManager{}
	roleManager.CreateRoleFn = func(role *am.Role) (string, error) {
		return "id", nil
	}
	roleManager.AddMembersFn = func(orgID int, roleID string, members []int) error {
		return nil
	}
	return roleManager
}

func MockEmptyAuthorizer() *mock.Authorizer {
	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	auth.IsUserAllowedFn = func(orgID, userID int, resource, action string) error {
		return nil
	}
	return auth
}

func InitDB(env string, t *testing.T) *pgx.ConnPool {
	sec := secrets.NewDBSecrets(env, "")
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
		OrgName:         orgName,
		OwnerEmail:      orgName + "email@email.com",
		UserPoolID:      "userpool.blah",
		IdentityPoolID:  "identitypool.blah",
		FirstName:       "first",
		LastName:        "last",
		Phone:           "1-111-111-1111",
		Country:         "USA",
		City:            "Beverly Hills",
		StatePrefecture: "CA",
		PostalCode:      "90210",
		Street:          "1 fake lane",
		SubscriptionID:  1000,
		StatusID:        1000,
	}
}

func CreateOrg(p *pgx.ConnPool, name string, t *testing.T) {
	_, err := p.Exec(CreateOrgStmt, name, GenerateID(t), "user_pool_id.blah", "identity_pool_id.blah",
		name+"email@email.com", "first", "last", "1-111-111-1111", "usa", "ca", "1 fake lane", "", "",
		"sf", "90210", time.Now().UnixNano())

	if err != nil {
		t.Fatalf("error creating organization %s: %s\n", name, err)
	}

	orgID := GetOrgID(p, name, t)

	_, err = p.Exec(CreateUserStmt, orgID, GenerateID(t), name+"email@email.com", "first", "last", am.UserStatusActive, time.Now().UnixNano())
	if err != nil {
		t.Fatalf("error creating user for %s, %s\n", name, err)
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

func CreateScanGroup(p *pgx.ConnPool, orgName, groupName string, t *testing.T) int {
	var groupID int
	orgID := GetOrgID(p, orgName, t)
	userID := GetUserId(p, orgID, orgName, t)
	//organization_id, scan_group_name, creation_time, created_by, modified_time, modified_by, original_input, configuration
	err := p.QueryRow(CreateScanGroupStmt, orgID, groupName, 0, userID, 0, userID, "s3://bucket/blah", nil).Scan(&groupID)
	if err != nil {
		t.Fatalf("error creating scan group: %s\n", err)
	}
	return groupID
}

// TestCompareOrganizations does not compare fields that are unknown prior to creation
// time (creation time, org id, orgcid)
func TestCompareOrganizations(expected, returned *am.Organization, t *testing.T) {
	e := expected
	r := returned

	if e.OrgName != r.OrgName {
		t.Fatalf("org name did not match expected: %v got %v\n", e.OrgName, r.OrgName)
	}

	if e.OwnerEmail != r.OwnerEmail {
		t.Fatalf("OwnerEmail did not match expected: %v got %v\n", e.OwnerEmail, r.OwnerEmail)
	}

	if e.UserPoolID != r.UserPoolID {
		t.Fatalf("UserPoolID did not match expected: %v got %v\n", e.UserPoolID, r.UserPoolID)
	}

	if e.IdentityPoolID != r.IdentityPoolID {
		t.Fatalf("IdentityPoolID did not match expected: %v got %v\n", e.IdentityPoolID, r.IdentityPoolID)
	}

	if e.FirstName != r.FirstName {
		t.Fatalf("FirstName did not match expected: %v got %v\n", e.FirstName, r.FirstName)
	}

	if e.LastName != r.LastName {
		t.Fatalf("LastName did not match expected: %v got %v\n", e.LastName, r.LastName)
	}

	if e.Phone != r.Phone {
		t.Fatalf("Phone did not match expected: %v got %v\n", e.Phone, r.Phone)
	}

	if e.Country != r.Country {
		t.Fatalf("Country did not match expected: %v got %v\n", e.Country, r.Country)
	}

	if e.StatePrefecture != r.StatePrefecture {
		t.Fatalf("StatePrefecture did not match expected: %v got %v\n", e.StatePrefecture, r.StatePrefecture)
	}

	if e.Street != r.Street {
		t.Fatalf("Street did not match expected: %v got %v\n", e.Street, r.Street)
	}

	if e.Address1 != r.Address1 {
		t.Fatalf("Address1 did not match expected: %v got %v\n", e.Address1, r.Address1)
	}

	if e.Address2 != r.Address2 {
		t.Fatalf("Address2 did not match expected: %v got %v\n", e.Address2, r.Address2)
	}

	if e.City != r.City {
		t.Fatalf("City did not match expected: %v got %v\n", e.City, r.City)
	}

	if e.PostalCode != r.PostalCode {
		t.Fatalf("PostalCode did not match expected: %v got %v\n", e.PostalCode, r.PostalCode)
	}

	if e.StatusID != r.StatusID {
		t.Fatalf("StatusID did not match expected: %v got %v\n", e.StatusID, r.StatusID)
	}

	if e.Deleted != r.Deleted {
		t.Fatalf("Deleted did not match expected: %v got %v\n", e.StatePrefecture, r.StatePrefecture)
	}

	if e.SubscriptionID != r.SubscriptionID {
		t.Fatalf("SubscriptionID did not match expected: %v got %v\n", e.StatePrefecture, r.StatePrefecture)
	}

	if r.CreationTime <= 0 {
		t.Fatalf("creation time of returned was not set\n")
	}
}

func TestCompareAddresses(expected, returned map[int64]*am.ScanGroupAddress, t *testing.T) {

	expectedKeys := make([]int64, len(expected))
	i := 0
	for k := range expected {
		expectedKeys[i] = k
		i++
	}
	returnedKeys := make([]int64, len(returned))
	i = 0
	for k := range returned {
		returnedKeys[i] = k
		i++
	}

	SortEqualInt64(expectedKeys, returnedKeys, t)

	for addrID := range returned {
		e := expected[addrID]
		r := returned[addrID]

		if !Float32Equals(e.ConfidenceScore, r.ConfidenceScore) {
			t.Fatalf("ConfidenceScore by was different, %v and %v\n", e.ConfidenceScore, r.ConfidenceScore)
		}

		if !Float32Equals(e.UserConfidenceScore, r.UserConfidenceScore) {
			t.Fatalf("UserConfidenceScore by was different, %v and %v\n", e.UserConfidenceScore, r.UserConfidenceScore)
		}

		if e.DiscoveredBy != r.DiscoveredBy {
			t.Fatalf("DiscoveredBy by was different, %v and %v\n", e.DiscoveredBy, r.DiscoveredBy)
		}

		if e.DiscoveryTime != r.DiscoveryTime {
			t.Fatalf("DiscoveryTime by was different, %v and %v\n", e.DiscoveryTime, r.DiscoveryTime)
		}

		if e.GroupID != r.GroupID {
			t.Fatalf("GroupID by was different, %v and %v\n", e.GroupID, r.GroupID)
		}

		if e.HostAddress != r.HostAddress {
			t.Fatalf("HostAddress by was different, %v and %v\n", e.HostAddress, r.HostAddress)
		}

		if e.IPAddress != r.IPAddress {
			t.Fatalf("IPAddress by was different, %v and %v\n", e.IPAddress, r.IPAddress)
		}

		if e.Ignored != r.Ignored {
			t.Fatalf("Ignored by was different, %v and %v\n", e.Ignored, r.Ignored)
		}

		if e.IsHostedService != r.IsHostedService {
			t.Fatalf("IsHostedService by was different, %v and %v\n", e.IsHostedService, r.IsHostedService)
		}

		if e.IsSOA != r.IsSOA {
			t.Fatalf("IsSOA by was different, %v and %v\n", e.IsSOA, r.IsSOA)
		}

		if e.IsWildcardZone != r.IsWildcardZone {
			t.Fatalf("IsWildcardZone by was different, %v and %v\n", e.IsWildcardZone, r.IsWildcardZone)
		}

		if e.LastScannedTime != r.LastScannedTime {
			t.Fatalf("LastScannedTime by was different, %v and %v\n", e.LastScannedTime, r.LastScannedTime)
		}

		if e.LastSeenTime != r.LastSeenTime {
			t.Fatalf("LastSeenTime by was different, %v and %v\n", e.LastSeenTime, r.LastSeenTime)
		}

		if e.NSRecord != r.NSRecord {
			t.Fatalf("NSRecord by was different, %v and %v\n", e.NSRecord, r.NSRecord)
		}

		if e.AddressHash != r.AddressHash {
			t.Fatalf("AddressHash by was different, %v and %v\n", e.AddressHash, r.AddressHash)
		}

		if e.FoundFrom != r.FoundFrom {
			t.Fatalf("FoundFrom by was different, %v and %v\n", e.FoundFrom, r.FoundFrom)
		}
	}
}

func TestCompareScanGroup(group1, group2 *am.ScanGroup, t *testing.T) {
	if group1.CreatedBy != group2.CreatedBy {
		t.Fatalf("created by was different, %d and %d\n", group1.CreatedBy, group2.CreatedBy)
	}

	if group1.ModifiedBy != group2.ModifiedBy {
		t.Fatalf("modified by was different, %d and %d\n", group1.ModifiedBy, group2.ModifiedBy)
	}

	if group1.CreationTime != group2.CreationTime {
		t.Fatalf("creation time by was different, %d and %d\n", group1.CreationTime, group2.CreationTime)
	}

	if group1.ModifiedTime != group2.ModifiedTime {
		t.Fatalf("ModifiedTime by was different, %d and %d\n", group1.CreationTime, group2.CreationTime)
	}

	if group1.GroupID != group2.GroupID {
		t.Fatalf("GroupID by was different, %d and %d\n", group1.GroupID, group2.GroupID)
	}

	if group1.OrgID != group2.OrgID {
		t.Fatalf("OrgID by was different, %d and %d\n", group1.OrgID, group2.OrgID)
	}

	if group1.GroupName != group2.GroupName {
		t.Fatalf("GroupName by was different, %s and %s\n", group1.GroupName, group2.GroupName)
	}

	if string(group1.OriginalInputS3URL) != string(group2.OriginalInputS3URL) {
		t.Fatalf("OriginalInput by was different, %s and %s\n", string(group1.OriginalInputS3URL), string(group2.OriginalInputS3URL))
	}
}

func TestCompareGroupModules(e, r *am.ModuleConfiguration, t *testing.T) {
	if e.BruteModule.RequestsPerSecond != r.BruteModule.RequestsPerSecond {
		t.Fatalf("BruteModule.RequestsPerSecond expected %v got %v\n", e.BruteModule.RequestsPerSecond, r.BruteModule.RequestsPerSecond)
	}

	if e.NSModule.RequestsPerSecond != r.NSModule.RequestsPerSecond {
		t.Fatalf("NSModule.RequestsPerSecond expected %v got %v\n", e.NSModule.RequestsPerSecond, r.NSModule.RequestsPerSecond)
	}

	if e.PortModule.RequestsPerSecond != r.PortModule.RequestsPerSecond {
		t.Fatalf("PortModule.RequestsPerSecond expected %v got %v\n", e.PortModule.RequestsPerSecond, r.PortModule.RequestsPerSecond)
	}

	if e.WebModule.RequestsPerSecond != r.WebModule.RequestsPerSecond {
		t.Fatalf("WebModule.RequestsPerSecond expected %v got %v\n", e.WebModule.RequestsPerSecond, r.WebModule.RequestsPerSecond)
	}

	if !SortEqualString(e.BruteModule.CustomSubNames, r.BruteModule.CustomSubNames, t) {
		t.Fatalf("BruteModule expected %v got %v\n", e.BruteModule.CustomSubNames, r.BruteModule.CustomSubNames)
	}

	if e.BruteModule.MaxDepth != r.BruteModule.MaxDepth {
		t.Fatalf("BruteModule.MaxDepth expected %v got %v\n", e.BruteModule.MaxDepth, r.BruteModule.MaxDepth)
	}

	if !SortEqualString(e.KeywordModule.Keywords, r.KeywordModule.Keywords, t) {
		t.Fatalf("KeywordModule expected %v got %v\n", e.KeywordModule.Keywords, r.KeywordModule.Keywords)
	}

	if !SortEqualInt32(e.PortModule.CustomPorts, r.PortModule.CustomPorts, t) {
		t.Fatalf("PortModule.CustomPorts expected %v got %v\n", e.PortModule.CustomPorts, r.PortModule.CustomPorts)
	}

	if e.WebModule.ExtractJS != r.WebModule.ExtractJS {
		t.Fatalf("WebModule.ExtractJS expected %v got %v\n", e.WebModule.ExtractJS, r.WebModule.ExtractJS)
	}

	if e.WebModule.FingerprintFrameworks != r.WebModule.FingerprintFrameworks {
		t.Fatalf("WebModule.FingerprintFrameworks expected %v got %v\n", e.WebModule.FingerprintFrameworks, r.WebModule.FingerprintFrameworks)
	}

	if e.WebModule.MaxLinks != r.WebModule.MaxLinks {
		t.Fatalf("WebModule.MaxLinks expected %v got %v\n", e.WebModule.MaxLinks, r.WebModule.MaxLinks)
	}

	if e.WebModule.TakeScreenShots != r.WebModule.TakeScreenShots {
		t.Fatalf("WebModule.TakeScreenShots expected %v got %v\n", e.WebModule.TakeScreenShots, r.WebModule.TakeScreenShots)
	}
}

func SortEqualString(expected, returned []string, t *testing.T) bool {
	if len(expected) != len(returned) {
		t.Fatalf("slice did not match size: %#v, %#v\n", expected, returned)
	}
	expectedCopy := make([]string, len(expected))
	returnedCopy := make([]string, len(returned))

	copy(expectedCopy, expected)
	copy(returnedCopy, returned)

	sort.Strings(expectedCopy)
	sort.Strings(returnedCopy)
	return reflect.DeepEqual(expectedCopy, returnedCopy)
}

func SortEqualInt(expected, returned []int, t *testing.T) bool {
	if len(expected) != len(returned) {
		t.Fatalf("slice did not match size: %#v, %#v\n", expected, returned)
	}
	expectedCopy := make([]int, len(expected))
	returnedCopy := make([]int, len(returned))

	copy(expectedCopy, expected)
	copy(returnedCopy, returned)

	sort.Ints(expectedCopy)
	sort.Ints(returnedCopy)

	return reflect.DeepEqual(expectedCopy, returnedCopy)
}

func SortEqualInt32(expected, returned []int32, t *testing.T) bool {
	if len(expected) != len(returned) {
		t.Fatalf("slice did not match size: %#v, %#v\n", expected, returned)
	}
	expectedCopy := make([]int32, len(expected))
	returnedCopy := make([]int32, len(returned))

	copy(expectedCopy, expected)
	copy(returnedCopy, returned)
	sort.Slice(expectedCopy, func(i, j int) bool { return expectedCopy[i] < expectedCopy[j] })
	sort.Slice(returnedCopy, func(i, j int) bool { return returnedCopy[i] < returnedCopy[j] })

	return reflect.DeepEqual(expectedCopy, returnedCopy)
}

func SortEqualInt64(expected, returned []int64, t *testing.T) bool {
	if len(expected) != len(returned) {
		t.Fatalf("slice did not match size: %#v, %#v\n", expected, returned)
	}
	expectedCopy := make([]int64, len(expected))
	returnedCopy := make([]int64, len(returned))

	copy(expectedCopy, expected)
	copy(returnedCopy, returned)
	sort.Slice(expectedCopy, func(i, j int) bool { return expectedCopy[i] < expectedCopy[j] })
	sort.Slice(returnedCopy, func(i, j int) bool { return returnedCopy[i] < returnedCopy[j] })

	return reflect.DeepEqual(expectedCopy, returnedCopy)
}

const epsilon = 1e-4

func Float32Equals(a, b float32) bool {
	return Float32EqualEPS(a, b, epsilon)
}

func Float32EqualEPS(a, b float32, eps float32) bool {
	return (a-b) < eps && (b-a) < eps
}

func Float64Equals(a, b float64) bool {
	return Float64EqualEPS(a, b, epsilon)
}

func Float64EqualEPS(a, b float64, eps float64) bool {
	return (a-b) < eps && (b-a) < eps
}
