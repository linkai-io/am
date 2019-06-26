package am

import "context"

type PortData struct {
	IPAddress  string   `json:"ip_address"`
	TCPPorts   []int32  `json:"tcp_ports"`
	UDPPorts   []int32  `json:"udp_ports"`
	TCPBanners []string `json:"tcp_banners,omitempty"`
	UDPBanners []string `json:"udp_banners,omitempty"`
}

type Ports struct {
	Current  *PortData `json:"current"`
	Previous *PortData `json:"previous,omitempty"`
}

type PortResults struct {
	PortID                   int64  `json:"port_id"`
	OrgID                    int    `json:"org_id"`
	GroupID                  int    `json:"group_id"`
	HostAddress              string `json:"host_address"` // could be IP address if hostname is empty from ScanGroupAddress
	Ports                    *Ports `json:"port_data"`
	ScannedTimestamp         int64  `json:"scanned_timestamp"`
	PreviousScannedTimestamp int64  `json:"previous_scanned_timestamp"`
}

type PortScannerService interface {
	AddGroup(ctx context.Context, userContext UserContext, group *ScanGroup) error
	RemoveGroup(ctx context.Context, userContext UserContext, orgID, groupID int) error
	Analyze(ctx context.Context, userContext UserContext, address *ScanGroupAddress) (*ScanGroupAddress, *PortResults, error)
}
