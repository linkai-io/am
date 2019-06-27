package portscanner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/linkai-io/am/pkg/retrier"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type ScanResults struct {
	Open   []int32 `json:"open"`
	Closed []int32 `json:"closed"`
}

type portResult struct {
	Port   int32
	Closed bool
}

type Scanner struct {
	device string
	srcIP  string
	srcMac net.HardwareAddr
	dstMac net.HardwareAddr

	// settings
	timeout time.Duration

	// destination, gateway (if applicable), and source IP addresses to use.
	dst, gw, src net.IP
	handle       *pcap.Handle
	// opts and buf allow us to easily serialize packets in the send()
	// method.
	opts gopacket.SerializeOptions
	buf  gopacket.SerializeBuffer
}

// New returns a new scanner with a default timeout of 1 minute after send completion
func New() *Scanner {
	return &Scanner{timeout: time.Minute}
}

// NewWithMAC builds a new scanner with provided src/dst macs for routing with a default
// timeout of 1 minute after send completion
func NewWithMAC(srcMac, dstMac net.HardwareAddr, srcIP string) *Scanner {
	return &Scanner{timeout: time.Minute, srcMac: srcMac, dstMac: dstMac, srcIP: srcIP}
}

// SetTimeout for when to shutdown the pcap handle after send completes (default is 1 minute or if all ports respond)
func (s *Scanner) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// Init the scanner with the provided device
func (s *Scanner) Init(device string) error {
	s.device = device
	if device == "" {
		return errors.New("error device must be specified if mac addresses are not set")
	}

	s.opts = gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	if s.srcMac == nil || s.dstMac == nil {
		return s.initDevice(device)
	}

	return nil
}

// initDevice automatically detects src/dst mac addresses and source ip by sending an
// icmp echo with a ttl of 1, listens for responses and extracts necessary information.
func (s *Scanner) initDevice(device string) error {
	var err error
	var snapshotLen int32 = 60
	var promiscuous bool
	var timeout = 2 * time.Second
	var handle *pcap.Handle
	var srcMac string
	var dstMac string

	// Open device
	handle, err = pcap.OpenLive(device, snapshotLen, promiscuous, timeout)
	if err != nil {
		return err
	}
	defer handle.Close()

	go ipv4GetDstCheck()

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		if packet.NetworkLayer() != nil {
			flow := packet.NetworkLayer().NetworkFlow()

			if flow.Dst().String() == "1.1.1.1" {
				srcMac = packet.LinkLayer().LinkFlow().Src().String()
				dstMac = packet.LinkLayer().LinkFlow().Dst().String()
				s.srcIP = packet.NetworkLayer().NetworkFlow().Src().String()
				log.Info().Msgf("got src mac: %s (%s) and dst mac: %s", srcMac, s.srcIP, dstMac)
				break
			}
		}
	}
	handle.Close()

	s.srcMac, err = net.ParseMAC(srcMac)
	if err != nil {
		return err
	}
	s.dstMac, err = net.ParseMAC(dstMac)
	if err != nil {
		return err
	}
	return nil
}

// ScanIPv4 scans an IPv4 address
func (s *Scanner) ScanIPv4(ctx context.Context, targetIP string, packetsPerSecond int, ports []int32) (*ScanResults, error) {
	log.Info().Msgf("Scanning with smac: %v sip: %v dmac: %v target: %v", s.srcMac, s.srcIP, s.dstMac, targetIP)
	// Construct all the network layers we need.
	eth := layers.Ethernet{
		SrcMAC:       s.srcMac,
		DstMAC:       s.dstMac,
		EthernetType: layers.EthernetTypeIPv4,
	}
	ip4 := layers.IPv4{
		SrcIP:    net.ParseIP(s.srcIP),
		DstIP:    net.ParseIP(targetIP),
		Version:  4,
		TTL:      160,
		Protocol: layers.IPProtocolTCP,
	}

	rawPort, err := getFreePort()
	if err != nil {
		return nil, err
	}

	tcp := layers.TCP{
		SrcPort: layers.TCPPort(rawPort),
		DstPort: 0,
		SYN:     true,
	}

	tcp.SetNetworkLayerForChecksum(&ip4)

	handle, err := pcap.OpenLive(s.device, 512, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	resultCh := make(chan *portResult, len(ports))

	ipFlow := gopacket.NewFlow(layers.EndpointIPv4, ip4.DstIP, ip4.SrcIP)
	go s.listen(ctx, handle, resultCh, ipFlow, len(ports), rawPort)

	r := rate.Limit(packetsPerSecond)
	lim := rate.NewLimiter(r, 1)
	for _, port := range ports {
		tcp.DstPort = layers.TCPPort(port)
		if err := s.send(handle, &eth, &ip4, &tcp); err != nil {
			return nil, err
		}
		lim.Wait(ctx)
	}

	timer := time.AfterFunc(s.timeout, func() { handle.Close() })
	defer timer.Stop()

	// use a map *just in case* we get duplicate responses
	openResults := make(map[int32]struct{})
	closedResults := make(map[int32]struct{})

	for result := range resultCh {
		if result.Closed {
			closedResults[result.Port] = struct{}{}
		} else {
			openResults[result.Port] = struct{}{}
		}
	}
	results := &ScanResults{
		Open:   make([]int32, len(openResults)),
		Closed: make([]int32, len(closedResults)),
	}

	i := 0
	for port := range openResults {
		results.Open[i] = port
		i++
	}

	i = 0
	for port := range closedResults {
		results.Closed[i] = port
		i++
	}
	return results, nil
}

// listen on the pcap handle and decode eth/ip4/tcp packet responses.
func (s *Scanner) listen(ctx context.Context, handle *pcap.Handle, resultCh chan<- *portResult, ipFlow gopacket.Flow, total, rawPort int) {
	decodeETH := &layers.Ethernet{}
	decodeIP4 := &layers.IPv4{}
	decodeTCP := &layers.TCP{}
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, decodeETH, decodeIP4, decodeTCP)

	for {

		select {
		case <-ctx.Done():
			break
		default:
		}

		// Read in the next packet.
		data, _, err := handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			log.Info().Msg("NextErrorTimeoutExpired")
			break
		} else if err == io.EOF {
			break
		} else if err != nil {
			// connection closed
			log.Error().Err(err).Msg("Packet read error")
			continue
		}

		decoded := []gopacket.LayerType{}
		if err := parser.DecodeLayers(data, &decoded); err != nil {
			continue
		}
		for _, layerType := range decoded {
			switch layerType {
			case layers.LayerTypeIPv4:
				if decodeIP4.NetworkFlow() != ipFlow {
					continue
				}
			case layers.LayerTypeTCP:
				if decodeTCP.DstPort != layers.TCPPort(rawPort) {
					continue
				} else if decodeTCP.SYN && decodeTCP.ACK {
					resultCh <- &portResult{Port: int32(decodeTCP.SrcPort)}
					fmt.Printf("open: %d\n", decodeTCP.SrcPort)
					total--
				} else if decodeTCP.RST {
					fmt.Printf("closed: %d\n", decodeTCP.SrcPort)
					resultCh <- &portResult{Port: int32(decodeTCP.SrcPort), Closed: true}
					total--
				}
			}
		}

		if total == 0 {
			break
		}
	}
	close(resultCh)
}

func (s *Scanner) send(handle *pcap.Handle, l ...gopacket.SerializableLayer) error {
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, s.opts, l...); err != nil {
		return err
	}
	return handle.WritePacketData(buf.Bytes())
}

// getFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (int, error) {
	var rawPort int

	retryErr := retrier.RetryAttempts(func() error {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}

		defer l.Close()
		rawPort = l.Addr().(*net.TCPAddr).Port
		return nil
	}, 4)
	return rawPort, retryErr
}

// ipv4GetDstCheck send an icmp echo with a ttl of 1 so we can identify what our
// target mac addrs are for routing. Shout out to hdmoore for this trick
func ipv4GetDstCheck() {
	c, err := net.ListenPacket("ip4:1", "0.0.0.0") // ICMP for IPv4
	if err != nil {
		log.Fatal().Err(err).Msg("error during ipv4 get destination check")
	}
	p := ipv4.NewPacketConn(c)
	p.SetTTL(1)

	wm := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID:   0xffff,
			Data: []byte("ROUTE-TEST"),
		},
	}

	wb, err := wm.Marshal(nil)
	if err != nil {
		log.Fatal().Err(err).Msg("error during ipv4 get destination check marshal icmp message")
	}

	dst := net.IPAddr{IP: net.IP([]byte{1, 1, 1, 1})}
	for i := 0; i < 5; i++ {
		if _, err := p.WriteTo(wb, nil, &dst); err != nil {
			log.Fatal().Err(err).Msg("error during ipv4 get destination check write")
		}
		time.Sleep(500 * time.Millisecond)
	}
}
