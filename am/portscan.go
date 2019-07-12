package am

import "context"

// DefaultTCPPorts are the list of default ports for port scanning
var DefaultTCPPorts = []int32{21, 22, 23, 25, 53, 80, 135, 139, 443, 445, 1443, 1723, 3306, 3389, 5432, 5900, 6379, 8000, 8080, 8443, 8500, 9500, 27017}

// DefaultUDPPorts are the list of default udp ports for port scanning
var DefaultUDPPorts = []int32{500, 1194}

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

// TCPChanges reports changes between current and previous
func (p *Ports) TCPChanges() (open []int32, closed []int32, change bool) {
	current := p.Current
	prev := p.Previous
	currentPorts := make(map[int32]struct{})
	previousPorts := make(map[int32]struct{})

	if current != nil && current.TCPPorts != nil {
		for _, port := range current.TCPPorts {
			currentPorts[port] = struct{}{}
		}
	}

	if prev != nil && prev.TCPPorts != nil {
		for _, port := range prev.TCPPorts {
			previousPorts[port] = struct{}{}
		}
	}

	for port := range currentPorts {
		if _, exist := previousPorts[port]; !exist {
			change = true
			open = append(open, port)
		}
	}
	for port := range previousPorts {
		if _, exist := currentPorts[port]; !exist {
			change = true
			closed = append(closed, port)
		}
	}
	return open, closed, change
}

type PortResults struct {
	PortID                   int64  `json:"port_id,omitempty"`
	OrgID                    int    `json:"org_id,omitempty"`
	GroupID                  int    `json:"group_id,omitempty"`
	HostAddress              string `json:"host_address,omitempty"` // could be IP address if hostname is empty from ScanGroupAddress
	Ports                    *Ports `json:"port_data,omitempty"`
	ScannedTimestamp         int64  `json:"scanned_timestamp,omitempty"`
	PreviousScannedTimestamp int64  `json:"previous_scanned_timestamp,omitempty"`
}

type PortScannerService interface {
	AddGroup(ctx context.Context, userContext UserContext, group *ScanGroup) error
	RemoveGroup(ctx context.Context, userContext UserContext, orgID, groupID int) error
	Analyze(ctx context.Context, userContext UserContext, address *ScanGroupAddress) (*ScanGroupAddress, *PortResults, error)
}
