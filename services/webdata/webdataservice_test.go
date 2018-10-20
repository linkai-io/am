package webdata_test

import (
	"context"
	"flag"
	"testing"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/mock"
	"github.com/linkai-io/am/pkg/secrets"
	"github.com/linkai-io/am/services/webdata"
)

var env string
var dbstring string

func init() {
	var err error
	flag.StringVar(&env, "env", "local", "environment we are running tests in")
	flag.Parse()
	sec := secrets.NewDBSecrets(env, "")
	dbstring, err = sec.DBString(am.WebDataServiceKey)
	if err != nil {
		panic("error getting dbstring secret")
	}
}

func TestNew(t *testing.T) {
	auth := &mock.Authorizer{}
	auth.IsAllowedFn = func(subject, resource, action string) error {
		return nil
	}
	service := webdata.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}

}

func TestAdd(t *testing.T) {
	ctx := context.Background()

	orgName := "addweb"
	groupName := "addweb"

	auth := amtest.MockEmptyAuthorizer()

	service := webdata.New(auth)

	if err := service.Init([]byte(dbstring)); err != nil {
		t.Fatalf("error initalizing address service: %s\n", err)
	}

	db := amtest.InitDB(env, t)
	defer db.Close()

	amtest.CreateOrg(db, orgName, t)
	orgID := amtest.GetOrgID(db, orgName, t)
	//defer amtest.DeleteOrg(db, orgName, t)

	groupID := amtest.CreateScanGroup(db, orgName, groupName, t)
	userContext := amtest.CreateUserContext(orgID, 1)
	address := amtest.CreateScanGroupAddress(db, orgID, groupID, t)

	headers := make(map[string]interface{}, 0)
	headers["host"] = "example.com"

	response := &am.HTTPResponse{
		Scheme:            "http",
		HostAddress:       "example.com",
		IPAddress:         "93.184.216.34",
		ResponsePort:      "80",
		RequestedPort:     "80",
		Status:            200,
		StatusText:        "HTTP 200 OK",
		URL:               "http://example.com/",
		Headers:           headers,
		MimeType:          "text/html",
		RawBody:           "",
		RawBodyLink:       "s3://data/1/1/1/1",
		RawBodyHash:       "1111",
		ResponseTimestamp: time.Now().UnixNano(),
		IsDocument:        true,
		WebCertificate: &am.WebCertificate{
			Protocol:                          "h2",
			KeyExchange:                       "kex",
			KeyExchangeGroup:                  "keg",
			Cipher:                            "aes",
			Mac:                               "1234",
			CertificateId:                     0,
			SubjectName:                       "example.com",
			SanList:                           []string{"www.example.com", "example.com"},
			ValidFrom:                         time.Now().UnixNano(),
			ValidTo:                           time.Now().UnixNano(),
			CertificateTransparencyCompliance: "unknown",
		},
	}
	responses := make([]*am.HTTPResponse, 1)
	responses[0] = response

	webData := &am.WebData{
		Address:           address,
		Responses:         responses,
		SerializedDOM:     "",
		SerializedDOMHash: "1234",
		SerializedDOMLink: "s3:/1/2/3/4",
		Snapshot:          "",
		SnapshotLink:      "s3://snapshot/1",
		ResponseTimestamp: time.Now().UnixNano(),
	}

	_, err := service.Add(ctx, userContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}

	// test adding again
	_, err = service.Add(ctx, userContext, webData)
	if err != nil {
		t.Fatalf("failed: %v\n", err)
	}
}
