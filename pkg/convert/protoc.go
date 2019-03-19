package convert

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/protocservices/prototypes"
	"github.com/linkai-io/am/protocservices/scangroup"
)

// DomainToUser convert domain user type to protobuf user type
func DomainToUser(in *am.User) *prototypes.User {
	return &prototypes.User{
		OrgID:        int32(in.OrgID),
		OrgCID:       in.OrgCID,
		UserCID:      in.UserCID,
		UserID:       int32(in.UserID),
		UserEmail:    in.UserEmail,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		StatusID:     int32(in.StatusID),
		CreationTime: in.CreationTime,
		Deleted:      in.Deleted,
	}
}

// UserToDomain convert protobuf user type to domain user type
func UserToDomain(in *prototypes.User) *am.User {
	return &am.User{
		OrgID:        int(in.OrgID),
		OrgCID:       in.OrgCID,
		UserCID:      in.UserCID,
		UserID:       int(in.UserID),
		UserEmail:    in.UserEmail,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		StatusID:     int(in.StatusID),
		CreationTime: in.CreationTime,
		Deleted:      in.Deleted,
	}
}

func DomainToUserFilter(in *am.UserFilter) *prototypes.UserFilter {
	return &prototypes.UserFilter{
		Start:   int32(in.Start),
		Limit:   int32(in.Limit),
		OrgID:   int32(in.OrgID),
		Filters: DomainToFilterTypes(in.Filters),
	}
}

func UserFilterToDomain(in *prototypes.UserFilter) *am.UserFilter {
	return &am.UserFilter{
		Start:   int(in.Start),
		Limit:   int(in.Limit),
		OrgID:   int(in.OrgID),
		Filters: FilterTypesToDomain(in.Filters),
	}
}

// UserContextToDomain converts from a protoc usercontext to an am.usercontext
func UserContextToDomain(in *prototypes.UserContext) am.UserContext {
	return &am.UserContextData{
		TraceID:        in.TraceID,
		OrgID:          int(in.OrgID),
		OrgCID:         in.OrgCID,
		UserID:         int(in.UserID),
		UserCID:        in.UserCID,
		Roles:          in.Roles,
		IPAddress:      in.IPAddress,
		SubscriptionID: in.SubscriptionID,
	}
}

// DomainToUserContext converts the domain usercontext to protobuf usercontext
func DomainToUserContext(in am.UserContext) *prototypes.UserContext {
	return &prototypes.UserContext{
		TraceID:        in.GetTraceID(),
		OrgID:          int32(in.GetOrgID()),
		OrgCID:         in.GetOrgCID(),
		UserCID:        in.GetUserCID(),
		UserID:         int32(in.GetUserID()),
		Roles:          in.GetRoles(),
		IPAddress:      in.GetIPAddress(),
		SubscriptionID: in.GetSubscriptionID(),
	}
}

// DomainToOrganization converts the domain organization to protobuf organization
func DomainToOrganization(in *am.Organization) *prototypes.Org {
	return &prototypes.Org{
		OrgID:                      int32(in.OrgID),
		OrgCID:                     in.OrgCID,
		OrgName:                    in.OrgName,
		OwnerEmail:                 in.OwnerEmail,
		UserPoolID:                 in.UserPoolID,
		UserPoolAppClientID:        in.UserPoolAppClientID,
		UserPoolAppClientSecret:    in.UserPoolAppClientSecret,
		IdentityPoolID:             in.IdentityPoolID,
		FirstName:                  in.FirstName,
		LastName:                   in.LastName,
		Phone:                      in.Phone,
		Country:                    in.Country,
		StatePrefecture:            in.StatePrefecture,
		Street:                     in.Street,
		Address1:                   in.Address1,
		Address2:                   in.Address2,
		City:                       in.City,
		PostalCode:                 in.PostalCode,
		CreationTime:               in.CreationTime,
		StatusID:                   int32(in.StatusID),
		Deleted:                    in.Deleted,
		SubscriptionID:             in.SubscriptionID,
		UserPoolJWK:                in.UserPoolJWK,
		LimitTLD:                   in.LimitTLD,
		LimitTLDReached:            in.LimitTLDReached,
		LimitHosts:                 in.LimitHosts,
		LimitHostsReached:          in.LimitHostsReached,
		LimitCustomWebFlows:        in.LimitCustomWebFlows,
		LimitCustomWebFlowsReached: in.LimitCustomWebFlowsReached,
	}
}

// OrganizationToDomain converts the protobuf organization to domain organization
func OrganizationToDomain(in *prototypes.Org) *am.Organization {
	return &am.Organization{
		OrgID:                      int(in.OrgID),
		OrgCID:                     in.OrgCID,
		OrgName:                    in.OrgName,
		OwnerEmail:                 in.OwnerEmail,
		UserPoolID:                 in.UserPoolID,
		UserPoolAppClientID:        in.UserPoolAppClientID,
		UserPoolAppClientSecret:    in.UserPoolAppClientSecret,
		IdentityPoolID:             in.IdentityPoolID,
		UserPoolJWK:                in.UserPoolJWK,
		FirstName:                  in.FirstName,
		LastName:                   in.LastName,
		Phone:                      in.Phone,
		Country:                    in.Country,
		StatePrefecture:            in.StatePrefecture,
		Street:                     in.Street,
		Address1:                   in.Address1,
		Address2:                   in.Address2,
		City:                       in.City,
		PostalCode:                 in.PostalCode,
		CreationTime:               in.CreationTime,
		StatusID:                   int(in.StatusID),
		Deleted:                    in.Deleted,
		SubscriptionID:             in.SubscriptionID,
		LimitTLD:                   in.LimitTLD,
		LimitTLDReached:            in.LimitTLDReached,
		LimitHosts:                 in.LimitHosts,
		LimitHostsReached:          in.LimitHostsReached,
		LimitCustomWebFlows:        in.LimitCustomWebFlows,
		LimitCustomWebFlowsReached: in.LimitCustomWebFlowsReached,
	}
}

func DomainToOrgFilter(in *am.OrgFilter) *prototypes.OrgFilter {
	return &prototypes.OrgFilter{
		Start:   int32(in.Start),
		Limit:   int32(in.Limit),
		Filters: DomainToFilterTypes(in.Filters),
	}
}

// OrgFilterToDomain convert org filter protobuf to orgfilter domain
func OrgFilterToDomain(in *prototypes.OrgFilter) *am.OrgFilter {
	return &am.OrgFilter{
		Start:   int(in.Start),
		Limit:   int(in.Limit),
		Filters: FilterTypesToDomain(in.Filters),
	}
}

func AddressToDomain(in *prototypes.AddressData) *am.ScanGroupAddress {
	return &am.ScanGroupAddress{
		AddressID:           in.AddressID,
		OrgID:               int(in.OrgID),
		GroupID:             int(in.GroupID),
		HostAddress:         in.HostAddress,
		IPAddress:           in.IPAddress,
		DiscoveryTime:       in.DiscoveryTime,
		DiscoveredBy:        in.DiscoveredBy,
		LastScannedTime:     in.LastScannedTime,
		LastSeenTime:        in.LastSeenTime,
		ConfidenceScore:     in.ConfidenceScore,
		UserConfidenceScore: in.UserConfidenceScore,
		IsSOA:               in.IsSOA,
		IsWildcardZone:      in.IsWildcardZone,
		IsHostedService:     in.IsHostedService,
		Ignored:             in.Ignored,
		FoundFrom:           in.FoundFrom,
		NSRecord:            in.NSRecord,
		AddressHash:         in.AddressHash,
		Deleted:             in.Deleted,
	}
}

func DomainToAddress(in *am.ScanGroupAddress) *prototypes.AddressData {
	return &prototypes.AddressData{
		OrgID:               int32(in.OrgID),
		AddressID:           in.AddressID,
		GroupID:             int32(in.GroupID),
		HostAddress:         in.HostAddress,
		IPAddress:           in.IPAddress,
		DiscoveryTime:       in.DiscoveryTime,
		DiscoveredBy:        in.DiscoveredBy,
		LastScannedTime:     in.LastScannedTime,
		LastSeenTime:        in.LastSeenTime,
		ConfidenceScore:     in.ConfidenceScore,
		UserConfidenceScore: in.UserConfidenceScore,
		IsSOA:               in.IsSOA,
		IsWildcardZone:      in.IsWildcardZone,
		IsHostedService:     in.IsHostedService,
		Ignored:             in.Ignored,
		Deleted:             in.Deleted,
		FoundFrom:           in.FoundFrom,
		NSRecord:            in.NSRecord,
		AddressHash:         in.AddressHash,
	}
}

func HostListToDomain(in *prototypes.HostListData) *am.ScanGroupHostList {
	return &am.ScanGroupHostList{
		AddressIDs:  in.AddressIDs,
		OrgID:       int(in.OrgID),
		GroupID:     int(in.GroupID),
		ETLD:        in.ETLD,
		HostAddress: in.HostAddress,
		IPAddresses: in.IPAddresses,
	}
}

func DomainToHostList(in *am.ScanGroupHostList) *prototypes.HostListData {
	return &prototypes.HostListData{
		AddressIDs:  in.AddressIDs,
		OrgID:       int32(in.OrgID),
		GroupID:     int32(in.GroupID),
		ETLD:        in.ETLD,
		HostAddress: in.HostAddress,
		IPAddresses: in.IPAddresses,
	}
}

func AddressFilterToDomain(in *prototypes.AddressFilter) *am.ScanGroupAddressFilter {
	return &am.ScanGroupAddressFilter{
		OrgID:   int(in.OrgID),
		GroupID: int(in.GroupID),
		Start:   in.Start,
		Limit:   int(in.Limit),
		Filters: FilterTypesToDomain(in.Filters),
	}
}

func DomainToAddressFilter(in *am.ScanGroupAddressFilter) *prototypes.AddressFilter {
	return &prototypes.AddressFilter{
		OrgID:   int32(in.OrgID),
		GroupID: int32(in.GroupID),
		Start:   in.Start,
		Limit:   int32(in.Limit),
		Filters: DomainToFilterTypes(in.Filters),
	}
}

// ModuleToDomain converts protoc ModuleConfiguration to am.ModuleConfiguration
func ModuleToDomain(in *scangroup.ModuleConfiguration) *am.ModuleConfiguration {
	return &am.ModuleConfiguration{
		NSModule: &am.NSModuleConfig{
			RequestsPerSecond: in.NSConfig.RequestsPerSecond,
		},
		BruteModule: &am.BruteModuleConfig{
			RequestsPerSecond: in.BruteConfig.RequestsPerSecond,
			CustomSubNames:    in.BruteConfig.CustomSubNames,
			MaxDepth:          in.BruteConfig.MaxDepth,
		},
		PortModule: &am.PortModuleConfig{
			RequestsPerSecond: in.PortConfig.RequestsPerSecond,
			CustomPorts:       in.PortConfig.CustomPorts,
		},
		WebModule: &am.WebModuleConfig{
			RequestsPerSecond:     in.WebModuleConfig.RequestsPerSecond,
			TakeScreenShots:       in.WebModuleConfig.TakeScreenShots,
			MaxLinks:              in.WebModuleConfig.MaxLinks,
			ExtractJS:             in.WebModuleConfig.ExtractJS,
			FingerprintFrameworks: in.WebModuleConfig.FingerprintFrameworks,
		},
		KeywordModule: &am.KeywordModuleConfig{
			Keywords: in.KeywordModuleConfig.Keywords,
		},
	}
}

func DomainToModule(in *am.ModuleConfiguration) *scangroup.ModuleConfiguration {
	return &scangroup.ModuleConfiguration{
		NSConfig: &scangroup.NSModuleConfig{
			RequestsPerSecond: in.NSModule.RequestsPerSecond,
		},
		BruteConfig: &scangroup.BruteModuleConfig{
			RequestsPerSecond: in.BruteModule.RequestsPerSecond,
			CustomSubNames:    in.BruteModule.CustomSubNames,
			MaxDepth:          in.BruteModule.MaxDepth,
		},
		PortConfig: &scangroup.PortModuleConfig{
			RequestsPerSecond: in.PortModule.RequestsPerSecond,
			CustomPorts:       in.PortModule.CustomPorts,
		},
		WebModuleConfig: &scangroup.WebModuleConfig{
			RequestsPerSecond:     in.WebModule.RequestsPerSecond,
			TakeScreenShots:       in.WebModule.TakeScreenShots,
			MaxLinks:              in.WebModule.MaxLinks,
			ExtractJS:             in.WebModule.ExtractJS,
			FingerprintFrameworks: in.WebModule.FingerprintFrameworks,
		},
		KeywordModuleConfig: &scangroup.KeywordModuleConfig{
			Keywords: in.KeywordModule.Keywords,
		},
	}
}

// ScanGroupToDomain convert protoc group to domain type ScanGroup
func ScanGroupToDomain(in *scangroup.Group) *am.ScanGroup {
	return &am.ScanGroup{
		OrgID:                int(in.OrgID),
		GroupID:              int(in.GroupID),
		GroupName:            in.GroupName,
		CreationTime:         in.CreationTime,
		CreatedBy:            in.CreatedBy,
		CreatedByID:          int(in.CreatedByID),
		OriginalInputS3URL:   in.OriginalInputS3URL,
		ModifiedBy:           in.ModifiedBy,
		ModifiedByID:         int(in.ModifiedByID),
		ModifiedTime:         in.ModifiedTime,
		ModuleConfigurations: ModuleToDomain(in.ModuleConfiguration),
		Paused:               in.Paused,
		Deleted:              in.Deleted,
	}
}

// DomainToScanGroup convert domain type SdcanGroup to protoc Group
func DomainToScanGroup(in *am.ScanGroup) *scangroup.Group {
	return &scangroup.Group{
		OrgID:               int32(in.OrgID),
		GroupID:             int32(in.GroupID),
		GroupName:           in.GroupName,
		CreationTime:        in.CreationTime,
		CreatedBy:           in.CreatedBy,
		CreatedByID:         int32(in.CreatedByID),
		OriginalInputS3URL:  in.OriginalInputS3URL,
		ModifiedBy:          in.ModifiedBy,
		ModifiedByID:        int32(in.ModifiedByID),
		ModifiedTime:        in.ModifiedTime,
		ModuleConfiguration: DomainToModule(in.ModuleConfigurations),
		Paused:              in.Paused,
		Deleted:             in.Deleted,
	}
}

func DomainToScanGroupFilter(in *am.ScanGroupFilter) *scangroup.ScanGroupFilter {
	return &scangroup.ScanGroupFilter{
		Filters: DomainToFilterTypes(in.Filters),
	}
}

func ScanGroupFilterToDomain(in *scangroup.ScanGroupFilter) *am.ScanGroupFilter {
	return &am.ScanGroupFilter{
		Filters: FilterTypesToDomain(in.Filters),
	}
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

func DomainToGroupStats(in *am.GroupStats) *scangroup.GroupStats {
	return &scangroup.GroupStats{
		OrgID:           int32(in.OrgID),
		GroupID:         int32(in.GroupID),
		ActiveAddresses: in.ActiveAddresses,
		BatchSize:       in.BatchSize,
		LastUpdated:     in.LastUpdated,
		BatchStart:      in.BatchStart,
		BatchEnd:        in.BatchEnd,
	}
}

func DomainToGroupsStats(in map[int]*am.GroupStats) map[int32]*scangroup.GroupStats {
	stats := make(map[int32]*scangroup.GroupStats, len(in))
	for groupID, stat := range in {
		stats[int32(groupID)] = DomainToGroupStats(stat)
	}
	return stats
}

func GroupStatsToDomain(in *scangroup.GroupStats) *am.GroupStats {
	return &am.GroupStats{
		OrgID:           int(in.OrgID),
		GroupID:         int(in.GroupID),
		ActiveAddresses: in.ActiveAddresses,
		BatchSize:       in.BatchSize,
		LastUpdated:     in.LastUpdated,
		BatchStart:      in.BatchStart,
		BatchEnd:        in.BatchEnd,
	}
}

func GroupsStatsToDomain(in map[int32]*scangroup.GroupStats) map[int]*am.GroupStats {
	stats := make(map[int]*am.GroupStats, len(in))
	for groupID, stat := range in {
		stats[int(groupID)] = GroupStatsToDomain(stat)
	}

	return stats
}

func DomainToFilterTypes(in *am.FilterType) *prototypes.FilterType {
	filter := &prototypes.FilterType{}
	if in.BoolFilters != nil {
		filter.BoolFilters = make(map[string]*prototypes.BoolFilter)
		for k, v := range in.BoolFilters {
			if v == nil {
				continue
			}
			filter.BoolFilters[k] = &prototypes.BoolFilter{Value: v}
		}
	}
	if in.Int32Filters != nil {
		filter.Int32Filters = make(map[string]*prototypes.Int32Filter)
		for k, v := range in.Int32Filters {
			if v == nil {
				continue
			}
			filter.Int32Filters[k] = &prototypes.Int32Filter{Value: v}
		}
	}
	if in.Int64Filters != nil {
		filter.Int64Filters = make(map[string]*prototypes.Int64Filter)
		for k, v := range in.Int64Filters {
			if v == nil {
				continue
			}
			filter.Int64Filters[k] = &prototypes.Int64Filter{Value: v}
		}
	}
	if in.Float32Filters != nil {
		filter.FloatFilters = make(map[string]*prototypes.FloatFilter)
		for k, v := range in.Float32Filters {
			if v == nil {
				continue
			}
			filter.FloatFilters[k] = &prototypes.FloatFilter{Value: v}
		}
	}

	if in.StringFilters != nil {
		filter.StringFilters = make(map[string]*prototypes.StringFilter)
		for k, v := range in.StringFilters {
			if v == nil {
				continue
			}
			filter.StringFilters[k] = &prototypes.StringFilter{Value: v}
		}
	}
	return filter
}

func FilterTypesToDomain(in *prototypes.FilterType) *am.FilterType {
	filter := &am.FilterType{}
	if in.BoolFilters != nil {
		filter.BoolFilters = make(map[string][]bool)
		for k, v := range in.BoolFilters {
			if v == nil {
				continue
			}
			filter.BoolFilters[k] = v.Value
		}
	}
	if in.Int32Filters != nil {
		filter.Int32Filters = make(map[string][]int32)
		for k, v := range in.Int32Filters {
			if v == nil {
				continue
			}
			filter.Int32Filters[k] = v.Value
		}
	}
	if in.Int64Filters != nil {
		filter.Int64Filters = make(map[string][]int64)
		for k, v := range in.Int64Filters {
			if v == nil {
				continue
			}
			filter.Int64Filters[k] = v.Value
		}
	}
	if in.FloatFilters != nil {
		filter.Float32Filters = make(map[string][]float32)
		for k, v := range in.FloatFilters {
			if v == nil {
				continue
			}
			filter.Float32Filters[k] = v.Value
		}
	}
	if in.StringFilters != nil {
		filter.StringFilters = make(map[string][]string)
		for k, v := range in.StringFilters {
			if v == nil {
				continue
			}
			filter.StringFilters[k] = v.Value
		}
	}
	return filter
}

func DomainToScanGroupAggregates(in map[string]*am.ScanGroupAggregates) map[string]*prototypes.ScanGroupAggregates {
	if in == nil {
		return nil
	}
	agg := make(map[string]*prototypes.ScanGroupAggregates, len(in))
	for k, v := range in {
		agg[k] = &prototypes.ScanGroupAggregates{Time: v.Time, Count: v.Count}
	}
	return agg
}

func DomainToScanGroupAddressStats(in *am.ScanGroupAddressStats) *prototypes.ScanGroupAddressStats {
	return &prototypes.ScanGroupAddressStats{
		OrgID:          int32(in.OrgID),
		GroupID:        int32(in.GroupID),
		DiscoveredBy:   in.DiscoveredBy,
		Aggregates:     DomainToScanGroupAggregates(in.Aggregates),
		Total:          in.Total,
		ConfidentTotal: in.ConfidentTotal,
	}
}

func DomainToScanGroupsAddressStats(in []*am.ScanGroupAddressStats) []*prototypes.ScanGroupAddressStats {
	stats := make([]*prototypes.ScanGroupAddressStats, 0)
	for _, v := range in {
		stats = append(stats, DomainToScanGroupAddressStats(v))
	}
	return stats
}

func ScanGroupAggregatesToDomain(in map[string]*prototypes.ScanGroupAggregates) map[string]*am.ScanGroupAggregates {
	if in == nil {
		return nil
	}
	agg := make(map[string]*am.ScanGroupAggregates, len(in))
	for k, v := range in {
		agg[k] = &am.ScanGroupAggregates{Time: v.Time, Count: v.Count}
	}
	return agg
}

func ScanGroupAddressStatsToDomain(in *prototypes.ScanGroupAddressStats) *am.ScanGroupAddressStats {
	return &am.ScanGroupAddressStats{
		OrgID:          int(in.OrgID),
		GroupID:        int(in.GroupID),
		DiscoveredBy:   in.DiscoveredBy,
		Aggregates:     ScanGroupAggregatesToDomain(in.Aggregates),
		Total:          in.Total,
		ConfidentTotal: in.ConfidentTotal,
	}
}

func ScanGroupsAddressStatsToDomain(in []*prototypes.ScanGroupAddressStats) []*am.ScanGroupAddressStats {
	stats := make([]*am.ScanGroupAddressStats, 0)
	for _, v := range in {
		stats = append(stats, ScanGroupAddressStatsToDomain(v))
	}
	return stats
}
