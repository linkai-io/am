package portscanner

import "context"

type Executor interface {
	Init(config []byte) error
	PortScan(ctx context.Context, targetIP string, packetsPerSecond int, ports []int32) (*ScanResults, error)
}
