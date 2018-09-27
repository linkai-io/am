package ns_test

import (
	"context"
	"os"
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/ns"
)

const dnsServer = "127.0.0.53:53"
const localServer = "127.0.0.53:53"

func TestNS_Analyze(t *testing.T) {

	tests := []*am.ScanGroupAddress{
		&am.ScanGroupAddress{
			AddressID:    1,
			OrgID:        1,
			GroupID:      1,
			HostAddress:  "linkai.io",
			IPAddress:    "",
			DiscoveredBy: "input_list",
		},
		&am.ScanGroupAddress{
			AddressID:    2,
			OrgID:        1,
			GroupID:      1,
			HostAddress:  "",
			IPAddress:    "13.35.67.123",
			DiscoveredBy: "input_list",
		},
		&am.ScanGroupAddress{
			AddressID:    3,
			OrgID:        1,
			GroupID:      1,
			HostAddress:  "linkai.io",
			IPAddress:    "13.35.67.123",
			DiscoveredBy: "input_list",
		},
		&am.ScanGroupAddress{
			AddressID:    4,
			OrgID:        1,
			GroupID:      1,
			HostAddress:  "zonetransfer.me",
			IPAddress:    "",
			DiscoveredBy: "input_list",
		},
	}
	state := amtest.MockNSState()
	dc := dnsclient.New([]string{localServer}, 3)
	ns := ns.New(dc, state)
	ns.Init(nil)

	ctx := context.Background()

	for _, tt := range tests {
		t.Logf("%d\n", tt.AddressID)
		ns.Analyze(ctx, tt)
	}
}

func TestNetflixInput(t *testing.T) {
	orgID := 1
	groupID := 1
	addrFile, err := os.Open("testdata/netflix.txt")
	if err != nil {
		t.Fatalf("error opening test data: %s\n", err)
	}

	addrs := amtest.AddrsFromInputFile(orgID, groupID, addrFile, t)

	state := amtest.MockNSState()
	dc := dnsclient.New([]string{"127.0.0.53:53"}, 3)
	ns := ns.New(dc, state)
	ns.Init(nil)

	ctx := context.Background()
	updated := &am.ScanGroupAddress{}
	netflix := &am.ScanGroupAddress{}
	for _, addr := range addrs {
		if addr.HostAddress == "www.netflix.com" {
			netflix = addr
			updated, _, err = ns.Analyze(ctx, addr)
			if err != nil {
				t.Fatalf("error analyzing: %s\n", err)
			}
		}
	}

	if updated.AddressID != netflix.AddressID {
		t.Fatalf("expected addr id: %d got %d\n", netflix.AddressID, updated.AddressID)
	}

	if updated.IPAddress == "" {
		t.Fatalf("did not get ip address for updated netflix\n")
	}
}

func TestIsHostedAddress(t *testing.T) {
	address := &am.ScanGroupAddress{IsHostedService: false, HostAddress: "ec2-52-35-228-185.us-west-2.compute.amazonaws.com"}
	if !address.IsHostedService && address.HostAddress != "" {
		address.IsHostedService = ns.IsHostedDomain(address.HostAddress)
	}
	if address.IsHostedService == false {
		t.Fatalf("error host was not hosted service")
	}
}

func TestIsHostedDomain(t *testing.T) {
	var hosts = []string{
		"02395.cloudfront.net",
		"1345.02395.cloudfront.net",
		"asdf.cloudflare.net",
		"asdf.asdf.cloudflare.net",
		"ec2-34-211-85-116.us-west-2.compute.amazonaws.com",
		"test.wordpress.org",
		"x.x.wordpress.org",
		"test.github.io",
		"asdf.asdf.github.io",
		"asdf.fastly.net",
		"adf.asdf.fastly.net",
		"asdf.googleusercontent.com",
		"asdf.asdf.googleusercontent.com",
		"asdf.gstatic.com",
		"asdf.asdf.gstatic.com",
		"asdf.google.com",
		"asdf.compute.google.com",
		"asdf.blogger.com",
		"asdf.asdf.blogger.com",
		"asdf.shopify.com",
		"asdf.asdf.shopify.com",
		"asdf.adobeevents.com",
		"asdf.clouddn.com",
		"asdf.asdf.clouddn.com",
		"asdf.rackcdn.com",
		"asdf.asdf.rackcdn.com",
		"asdf.ampproject.org",
		"asdf.asdf.ampproject.org",
		"asdf.fc2.com",
		"asdf.asdf.fc2.com",
		"asdf.azure.com",
		"asdf.asdf.azure.com",
		"asdf.azurewebsites.net",
		"asdf.asdf.azurewebsites.net",
		"asdf.azurecontainer.io",
		"asdf.asdf.azurecontainer.io",
		"asdf.azure-mobile.net",
		"asdf.asdf.azure-mobile.net",
		"asdf.cloudapp.net",
		"asdf.asdf.cloudapp.net",
		"asdf.herokuapp.com",
		"asdf.asdf.herokuapp.com",
		"asdf.dyndns.org",
		"asdf.asdf.dyndns.org",
	}

	for _, host := range hosts {
		ret := ns.IsHostedDomain(host)
		if ret == false {
			t.Fatalf("should have returned true for %s\n", host)
		}
	}

}
