package portscan_test

import (
	"testing"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/dnsclient"
	"github.com/linkai-io/am/services/module/portscan"
)

func TestInit(t *testing.T) {
	dnsClient := dnsclient.New([]string{"1.1.1.1"}, 2)
	module := portscan.New(dnsClient)
	if err := module.Init(nil); err != nil {
		t.Fatalf("error init module: %v\n", err)
	}

	group := &am.ScanGroup{
		OrgID:              1,
		GroupID:            1,
		GroupName:          "",
		CreationTime:       0,
		CreatedBy:          "",
		CreatedByID:        0,
		ModifiedBy:         "",
		ModifiedByID:       0,
		ModifiedTime:       0,
		OriginalInputS3URL: "",
		ModuleConfigurations: &am.ModuleConfiguration{
			NSModule: &am.NSModuleConfig{
				RequestsPerSecond: 0,
			},
			BruteModule: &am.BruteModuleConfig{
				CustomSubNames:    nil,
				RequestsPerSecond: 0,
				MaxDepth:          0,
			},
			PortModule: &am.PortModuleConfig{
				RequestsPerSecond: 0,
				CustomPorts:       nil,
			},
			WebModule: &am.WebModuleConfig{
				TakeScreenShots:       false,
				RequestsPerSecond:     0,
				MaxLinks:              0,
				ExtractJS:             false,
				FingerprintFrameworks: false,
			},
			KeywordModule: &am.KeywordModuleConfig{
				Keywords: nil,
			},
		},
		Paused:           false,
		Deleted:          false,
		LastPausedTime:   0,
		ArchiveAfterDays: 0,
	}
	module.AddGroup(group)
	module.RemoveGroup(group.OrgID, group.GroupID)
}
