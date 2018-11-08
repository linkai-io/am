package filestorage_test

import (
	"context"
	"os"
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"

	"github.com/linkai-io/am/pkg/secrets"

	"github.com/linkai-io/am/pkg/filestorage"
)

func TestS3Storage(t *testing.T) {

	env := "local"
	region := "us-east-1"

	os.Setenv("_am_local_webfilepath", "test-am-webdata")

	s := filestorage.NewS3Storage(env, region)

	cache := secrets.NewSecretsCache(env, region)
	if err := s.Init(cache); err != nil {
		t.Fatalf("failed to initialize s3 storage: %v\n", err)
	}

	addr := &am.ScanGroupAddress{
		AddressID:           1,
		OrgID:               1,
		GroupID:             1,
		HostAddress:         "example.com",
		IPAddress:           "192.168.1.1",
		DiscoveryTime:       0,
		DiscoveredBy:        "",
		LastScannedTime:     0,
		LastSeenTime:        0,
		ConfidenceScore:     0.0,
		UserConfidenceScore: 0.0,
		IsSOA:               false,
		IsWildcardZone:      false,
		IsHostedService:     false,
		Ignored:             false,
		FoundFrom:           "",
		NSRecord:            0,
		AddressHash:         convert.HashAddress("192.168.1.1", "example.com"),
	}

	hash, link, err := s.Write(context.Background(), addr, []byte("hello"))
	if err != nil {
		t.Fatalf("error writing file to s3: %#v\n", err)
	}
	t.Logf("link: %v, hash: %v\n", link, hash)
}
