package amtest

import (
	"reflect"
	"sort"
	"testing"

	"github.com/linkai-io/am/am"
)

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

	if e.UserPoolAppClientID != r.UserPoolAppClientID {
		t.Fatalf("UserPoolAppClientID did not match expected: %v got %v\n", e.UserPoolAppClientID, r.UserPoolAppClientID)
	}

	if e.UserPoolAppClientSecret != r.UserPoolAppClientSecret {
		t.Fatalf("UserPoolAppClientSecret did not match expected: %v got %v\n", e.UserPoolAppClientSecret, r.UserPoolAppClientSecret)
	}

	if e.IdentityPoolID != r.IdentityPoolID {
		t.Fatalf("IdentityPoolID did not match expected: %v got %v\n", e.IdentityPoolID, r.IdentityPoolID)
	}

	if e.UserPoolJWK != r.UserPoolJWK {
		t.Fatalf("UserPoolJWK did not match expected: %v got %v\n", e.UserPoolJWK, r.UserPoolJWK)
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

// TestCompareAddresses tests all addresses in both maps' details
func TestCompareAddresses(expected, returned map[string]*am.ScanGroupAddress, t *testing.T) {

	expectedKeys := make([]string, len(expected))
	i := 0
	for k := range expected {
		expectedKeys[i] = k
		i++
	}
	returnedKeys := make([]string, len(returned))
	i = 0
	for k := range returned {
		returnedKeys[i] = k
		i++
	}

	SortEqualString(expectedKeys, returnedKeys, t)

	for addrID := range returned {
		e := expected[addrID]
		r := returned[addrID]
		TestCompareAddress(e, r, t)
	}
}

// TestCompareAddress details
func TestCompareAddress(e, r *am.ScanGroupAddress, t *testing.T) {
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

func TestCompareScanGroup(group1, group2 *am.ScanGroup, t *testing.T) {
	if group1.CreatedBy != group2.CreatedBy {
		t.Fatalf("created by was different, %v and %v\n", group1.CreatedBy, group2.CreatedBy)
	}

	if group1.ModifiedBy != group2.ModifiedBy {
		t.Fatalf("modified by was different, %v and %v\n", group1.ModifiedBy, group2.ModifiedBy)
	}

	if group1.CreatedByID != group2.CreatedByID {
		t.Fatalf("created byID was different, %v and %v\n", group1.CreatedByID, group2.CreatedByID)
	}

	if group1.ModifiedByID != group2.ModifiedByID {
		t.Fatalf("modified byID was different, %v and %v\n", group1.ModifiedByID, group2.ModifiedByID)
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

func TestCompareCTRecords(expected, returned map[string]*am.CTRecord, t *testing.T) {
	if len(expected) != len(returned) {
		t.Fatalf("lengths did not match expected: %d returned: %d\n", len(expected), len(returned))
	}

	for k, e := range expected {
		r, ok := returned[k]
		if !ok {
			t.Fatalf("expected record %v was not in our returned map\n", e.CertHash)
		}

		if e.ServerName != r.ServerName {
			t.Fatalf("ServerName: %v did not match returned: %v\n", e.ServerName, r.ServerName)
		}
		if e.ServerIndex != r.ServerIndex {
			t.Fatalf("ServerIndex: %v did not match returned: %v\n", e.ServerIndex, r.ServerIndex)
		}

		if e.CertHash != r.CertHash {
			t.Fatalf("CertHash: %v did not match returned: %v\n", e.CertHash, r.CertHash)
		}

		if e.CommonName != r.CommonName {
			t.Fatalf("CommonName: %v did not match returned: %v\n", e.CommonName, r.CommonName)
		}

		if e.Country != r.Country {
			t.Fatalf("Country: %v did not match returned: %v\n", e.Country, r.Country)
		}

		if e.EmailAddresses != r.EmailAddresses {
			t.Fatalf("EmailAddresses: %v did not match returned: %v\n", e.EmailAddresses, r.EmailAddresses)
		}

		if e.ETLD != r.ETLD {
			t.Fatalf("ETLD: %v did not match returned: %v\n", e.ETLD, r.ETLD)
		}

		if e.InsertedTime/1000 != r.InsertedTime/1000 {
			t.Fatalf("InsertedTime: %v did not match returned: %v\n", e.InsertedTime, r.InsertedTime)
		}

		if e.IPAddresses != r.IPAddresses {
			t.Fatalf("IPAddresses: %v did not match returned: %v\n", e.IPAddresses, r.IPAddresses)
		}

		if e.NotAfter != r.NotAfter {
			t.Fatalf("NotAfter: %v did not match returned: %v\n", e.NotAfter, r.NotAfter)
		}

		if e.NotBefore != r.NotBefore {
			t.Fatalf("NotBefore: %v did not match returned: %v\n", e.NotBefore, r.NotBefore)
		}

		if e.Organization != r.Organization {
			t.Fatalf("Organization: %v did not match returned: %v\n", e.Organization, r.Organization)
		}

		if e.OrganizationalUnit != r.OrganizationalUnit {
			t.Fatalf("OrganizationalUnit: %v did not match returned: %v\n", e.OrganizationalUnit, r.OrganizationalUnit)
		}

		if e.SerialNumber != r.SerialNumber {
			t.Fatalf("SerialNumber: %v did not match returned: %v\n", e.SerialNumber, r.SerialNumber)
		}

		if e.UnverifiedDNSNames != r.UnverifiedDNSNames {
			t.Fatalf("UnverifiedDNSNames: %v did not match returned: %v\n", e.UnverifiedDNSNames, r.UnverifiedDNSNames)
		}

		if e.VerifiedDNSNames != r.VerifiedDNSNames {
			t.Fatalf("VerifiedDNSNames: %v did not match returned: %v\n", e.VerifiedDNSNames, r.VerifiedDNSNames)
		}
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
