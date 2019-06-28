package portscanner

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
)

type scanRequest struct {
	TargetIP string  `json:"target_ip"`
	PPS      int     `json:"pps"`
	Ports    []int32 `json:"ports"`
}

// SocketClient portscans by calling over domain socket
type SocketClient struct {
	executorClient http.Client
}

// NewSocketClient builds a client to the socket server to issue port scan requests
func NewSocketClient() *SocketClient {
	e := &SocketClient{}
	e.executorClient = http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", SOCK)
			},
		},
	}
	return e
}

// Init the scanner (config not used atm)
func (e *SocketClient) Init(config []byte) error {
	return nil
}

// PortScan by calling the socket server to scan for us, marshal results and return
func (e *SocketClient) PortScan(ctx context.Context, targetIP string, packetsPerSecond int, ports []int32) (*ScanResults, error) {
	req := &scanRequest{
		TargetIP: targetIP,
		PPS:      packetsPerSecond,
		Ports:    ports,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := e.executorClient.Post("http://unix/scan", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	results := &ScanResults{}
	if err := json.Unmarshal(data, results); err != nil {
		return nil, err
	}

	return results, nil
}
