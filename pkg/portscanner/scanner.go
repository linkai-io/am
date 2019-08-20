package portscanner

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
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
	device     string
	srcIP      string
	srcMac     string
	dstMac     string
	srcMacAddr net.HardwareAddr
	dstMacAddr net.HardwareAddr
	macLock    *sync.RWMutex
	// settings
	timeout time.Duration
	retry   int
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
	return &Scanner{timeout: time.Minute, macLock: &sync.RWMutex{}, retry: 3}
}

// NewWithMAC builds a new scanner with provided src/dst macs for routing with a default
// timeout of 1 minute after send completion
func NewWithMAC(srcMac, dstMac net.HardwareAddr, device, srcIP string) *Scanner {
	return &Scanner{
		device:     device,
		srcIP:      srcIP,
		srcMacAddr: srcMac,
		dstMacAddr: dstMac,
		macLock:    &sync.RWMutex{},
		timeout:    time.Minute,
		retry:      3,
	}
}

// SetTimeout for when to shutdown the pcap handle after send completes (default is 1 minute or if all ports respond)
func (s *Scanner) SetTimeout(timeout time.Duration) {
	s.timeout = timeout
}

// Init the scanner with the provided device
func (s *Scanner) Init() error {
	s.opts = gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	if s.srcMacAddr == nil || s.dstMacAddr == nil {
		return s.initDevice()
	}

	return nil
}

// initDevice automatically detects device and src/dst mac addresses and source ip by sending an
// icmp echo with a ttl of 1 on all devices, and listens for responses and extracts necessary information.
func (s *Scanner) initDevice() error {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ifaceNames, err := getInterfaces()
	if err != nil {
		return err
	}

	foundCh := make(chan struct{})
	// Open devices
	for _, iface := range ifaceNames {
		if iface == "lo" {
			continue
		}
		i := iface
		go s.listenInterfaces(i, foundCh)
	}

	select {
	case <-ctx.Done():
		return errors.New("timeout waiting for interface detection")
	case <-foundCh:
	}

	s.srcMacAddr, err = net.ParseMAC(s.srcMac)
	if err != nil {
		return err
	}
	s.dstMacAddr, err = net.ParseMAC(s.dstMac)
	if err != nil {
		return err
	}
	return nil
}

func (s *Scanner) listenInterfaces(iface string, foundCh chan struct{}) {
	var err error
	var snapshotLen int32 = 60
	var promiscuous bool
	var timeout = 2 * time.Second
	var handle *pcap.Handle
	handle, err = pcap.OpenLive(iface, snapshotLen, promiscuous, timeout)
	if err != nil {
		return
	}
	defer handle.Close()

	go ipv4GetDstCheck()

	// Use the handle as a packet source to process all packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for packet := range packetSource.Packets() {
		if packet.NetworkLayer() != nil {
			flow := packet.NetworkLayer().NetworkFlow()

			if flow.Dst().String() == "1.1.1.1" {
				s.macLock.Lock()
				s.srcMac = packet.LinkLayer().LinkFlow().Src().String()
				s.dstMac = packet.LinkLayer().LinkFlow().Dst().String()
				s.srcIP = packet.NetworkLayer().NetworkFlow().Src().String()
				s.device = iface
				log.Info().Msgf("got src mac: %s (%s) and dst mac: %s for iface: %s", s.srcMac, s.srcIP, s.dstMac, s.device)
				s.macLock.Unlock()
				break
			}
		}
	}
	handle.Close()
	foundCh <- struct{}{}
}

// ScanIPv4 scans an IPv4 address
func (s *Scanner) ScanIPv4(ctx context.Context, targetIP string, packetsPerSecond int, ports []int32) (*ScanResults, error) {
	log.Info().Msgf("Scanning with smac: %v sip: %v dmac: %v target: %v", s.srcMac, s.srcIP, s.dstMac, targetIP)
	// Construct all the network layers we need.
	eth := layers.Ethernet{
		SrcMAC:       s.srcMacAddr,
		DstMAC:       s.dstMacAddr,
		EthernetType: layers.EthernetTypeIPv4,
	}

	destIP := net.ParseIP(targetIP)
	ip4 := layers.IPv4{
		SrcIP:    net.ParseIP(s.srcIP),
		DstIP:    destIP,
		Version:  4,
		TTL:      160,
		Protocol: layers.IPProtocolTCP,
	}

	rawPort, err := getFreePort()
	if err != nil {
		return nil, err
	}

	handle, err := pcap.OpenLive(s.device, 512, true, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	resultCh := make(chan *portResult, len(ports)*s.retry)

	ipFlow := gopacket.NewFlow(layers.EndpointIPv4, ip4.DstIP, ip4.SrcIP)
	go s.listen(ctx, handle, resultCh, destIP, ipFlow, len(ports), rawPort)

	r := rate.Limit(packetsPerSecond)
	lim := rate.NewLimiter(r, 1)

	for _, port := range ports {

		tcp := layers.TCP{
			SrcPort: layers.TCPPort(rawPort),
			DstPort: layers.TCPPort(port),
			SYN:     true,
			Seq:     CreateSequence(destIP, uint32(port)),
		}
		tcp.SetNetworkLayerForChecksum(&ip4)

		if err := s.send(handle, &eth, &ip4, &tcp); err != nil {
			log.Error().Err(err).Msg("failed to send packet")
			continue
		}
		lim.Wait(ctx)
	}

	timer := time.AfterFunc(s.timeout, func() { handle.Close() })
	defer timer.Stop()

	// use a map *just in case* we get duplicate responses
	openResults := make(map[int32]int)
	closedResults := make(map[int32]int)

	for result := range resultCh {
		if result.Closed {
			closedResults[result.Port]++
		} else {
			openResults[result.Port]++
		}
	}
	results := &ScanResults{
		Open:   make([]int32, 0),
		Closed: make([]int32, 0),
	}

	for port := range openResults {
		results.Open = append(results.Open, port)
	}

	for port := range closedResults {
		results.Closed = append(results.Closed, port)
	}

	return results, nil
}

// listen on the pcap handle and decode eth/ip4/tcp packet responses.
func (s *Scanner) listen(ctx context.Context, handle *pcap.Handle, resultCh chan<- *portResult, destIP net.IP, ipFlow gopacket.Flow, total, rawPort int) {
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
					expectedAck := CreateSequence(destIP, uint32(decodeTCP.SrcPort)) + 1 //add one for seq+1
					if expectedAck != decodeTCP.Ack {
						log.Warn().IPAddr("dest", destIP).Int32("port", int32(decodeTCP.SrcPort)).Uint32("expected", expectedAck).Uint32("ack", decodeTCP.Ack).Msg("expected ack did not match returned ack")
						continue
					}
					resultCh <- &portResult{Port: int32(decodeTCP.SrcPort)}
					total--
				} else if decodeTCP.RST {
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

func getInterfaces() ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(ifaces))

	for i, iface := range ifaces {
		names[i] = iface.Name
	}
	return names, nil
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
		log.Info().Err(err).Msg("error during ipv4 get destination check")
		return
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
		log.Info().Err(err).Msg("error during ipv4 get destination check marshal icmp message")
		return
	}

	dst := net.IPAddr{IP: net.IP([]byte{1, 1, 1, 1})}
	for i := 0; i < 5; i++ {
		if _, err := p.WriteTo(wb, nil, &dst); err != nil {
			log.Info().Err(err).Msg("error during ipv4 get destination check write")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func CreateSequence(ip net.IP, port uint32) uint32 {
	return binary.BigEndian.Uint32(ip.To4()) + port
}
