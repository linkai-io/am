package portscanner

import "context"

type LocalClient struct {
	scanner *Scanner
}

func NewLocalClient() *LocalClient {
	e := &LocalClient{scanner: New()}
	return e
}

// Init the scanner (config not used atm)
func (e *LocalClient) Init(config []byte) error {
	if err := e.scanner.Init(); err != nil {
		return err
	}
	return nil
}

func (e *LocalClient) PortScan(ctx context.Context, targetIP string, packetsPerSecond int, ports []int32) (*ScanResults, error) {
	return e.scanner.ScanIPv4(ctx, targetIP, packetsPerSecond, ports)
}
