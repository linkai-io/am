package main

import (
	"context"
	"time"

	"github.com/linkai-io/am/amtest"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/initializers"
)

func main() {
	cli := initializers.PortScanModule("scanner1.linkai.io:50053", "testtoken")
	userContext := &am.UserContextData{
		TraceID:        "test",
		OrgID:          1,
		OrgCID:         "test",
		UserID:         1,
		UserCID:        "test",
		Roles:          []string{"owner"},
		IPAddress:      "1.1.1.1",
		SubscriptionID: 1000,
	}
	group := amtest.CreateScanGroupOnly(1, 1)
	ctx := context.Background()

	if err := cli.AddGroup(ctx, userContext, group); err != nil {
		log.Fatal().Err(err).Msg("failed to add group")
	}
	log.Info().Msg("connected to service and sent group")
	address := &am.ScanGroupAddress{
		AddressID:           1,
		OrgID:               userContext.GetOrgID(),
		GroupID:             group.GroupID,
		HostAddress:         "test.linkai.io",
		IPAddress:           "209.126.252.34",
		DiscoveryTime:       time.Now().UnixNano(),
		DiscoveredBy:        "input_list",
		LastScannedTime:     time.Now().UnixNano(),
		LastSeenTime:        0,
		ConfidenceScore:     100,
		UserConfidenceScore: 0.0,
		IsSOA:               false,
		IsWildcardZone:      false,
		IsHostedService:     false,
		Ignored:             false,
		FoundFrom:           "abc",
		NSRecord:            0,
		AddressHash:         convert.HashAddress("209.126.252.34", ""),
		Deleted:             false,
	}

	addr, results, err := cli.Analyze(ctx, userContext, address)
	if err != nil {
		log.Info().Err(err).Msgf("failed to analyze: %v", err)
	}
	log.Info().Msgf("address returned: %#v", addr)

	if results != nil && results.Ports != nil && results.Ports.Current != nil {
		log.Info().Msgf("%#v\n", results.Ports.Current)
	}
	if err := cli.RemoveGroup(ctx, userContext, userContext.OrgID, group.GroupID); err != nil {
		log.Fatal().Err(err).Msg("failed to add group")
	}
}
