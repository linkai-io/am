package dnsclient

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/miekg/dns"
)

var (
	// ErrNoResponse Returned when there was no response
	ErrNoResponse = errors.New("error getting valid response")
	// ErrEmptyRecords returned when there were no records returned
	ErrEmptyRecords = errors.New("error empty record returned")
	// ErrRcode when Rcode != dns.RcodeSuccess
	ErrRcode = errors.New("bad Rcode returned by server")
	// ErrInvalidIP when IP is not properly formed
	ErrInvalidIP = errors.New("invalid format for IP address")
)

const (
	ipv4arpafmt = "%d.%d.%d.%d.in-addr.arpa"
	ipv6arpafmt = "%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.%c.ip6.arpa"
)

// Client to resolve hosts and ip addresses
type Client struct {
	client  *dns.Client
	servers []string
	retry   int
}

// New returns a new DNS client
func New(servers []string, retry int) *Client {
	c := &Client{}
	c.servers = servers
	c.retry = retry
	c.client = &dns.Client{Timeout: 5 * time.Second}
	return c
}

// IsWildcard tests if a domain is a wildcard domain
func (c *Client) IsWildcard(domain string) bool {
	return false
}

// ResolveName attempts to resolve a name to ip addresses. It will
// attempt to resolve both IPv4 and IPv6 addresses to the name
func (c *Client) ResolveName(name string) ([]*Results, error) {
	results := make([]*Results, 0)
	resultErrors := make(chan *resultError, 2)

	go c.queryA(name, resultErrors)
	go c.queryAAAA(name, resultErrors)

	recvd := 0
	for r := range resultErrors {
		if r.Error == nil {
			results = append(results, r.Result)
		}
		recvd++
		if recvd == 2 {
			goto DONE
		}
	}
DONE:
	close(resultErrors)
	return results, nil
}

func (c *Client) queryA(name string, rc chan *resultError) {
	result, err := c.exchange(name, dns.TypeA)
	if err != nil {
		rc <- &resultError{Result: nil, Error: err}
		return
	}

	results := &Results{RequestType: dns.TypeA}
	for _, answer := range result.Answer {
		if a, ok := answer.(*dns.A); ok {
			results.RecordType = dns.TypeA
			results.IPs = append(results.IPs, a.A.String())
			results.Hosts = append(results.Hosts, name)
		}
	}
	rc <- &resultError{Result: results, Error: nil}
}

func (c *Client) queryAAAA(name string, rc chan *resultError) {
	result, err := c.exchange(name, dns.TypeAAAA)
	if err != nil {
		rc <- &resultError{Result: nil, Error: err}
		return
	}

	results := &Results{RequestType: dns.TypeAAAA}
	for _, answer := range result.Answer {
		if a, ok := answer.(*dns.AAAA); ok {
			results.RecordType = dns.TypeAAAA
			results.IPs = append(results.IPs, a.AAAA.String())
			results.Hosts = append(results.Hosts, name)
		}
	}
	rc <- &resultError{Result: results, Error: nil}
}

// ResolveIP returns RRs for an IP address by parsing IP type and
// calling ipv4 or ipv6
func (c *Client) ResolveIP(ip string) (*Results, error) {
	name, err := dns.ReverseAddr(ip)
	if err != nil {
		return nil, ErrInvalidIP
	}

	result, err := c.exchange(name, dns.TypePTR)
	if err != nil {
		return nil, err
	}

	results := &Results{RequestType: dns.TypePTR}
	for _, answer := range result.Answer {
		if a, ok := answer.(*dns.PTR); ok {
			results.IPs = append(results.IPs, ip)
			results.RecordType = dns.TypePTR
			// TODO: check if inarpa check?
			results.Hosts = append(results.Hosts, strings.TrimRight(a.Ptr, "."))
		}
	}
	return results, nil
}

// LookupNS returns NS RRs for a zone
func (c *Client) LookupNS(zone string) (*Results, error) {
	result, err := c.exchange(zone, dns.TypeNS)
	if err != nil {
		return nil, err
	}

	results := &Results{RequestType: dns.TypeNS}
	for _, answer := range result.Answer {
		if a, ok := answer.(*dns.NS); ok {
			results.RecordType = dns.TypeNS
			results.Hosts = append(results.Hosts, a.Ns)
		}
	}
	return results, nil
}

// LookupMX returns MX RRs for a zone
func (c *Client) LookupMX(zone string) (*Results, error) {
	result, err := c.exchange(zone, dns.TypeMX)
	if err != nil {
		return nil, err
	}

	results := &Results{RequestType: dns.TypeMX}
	for _, answer := range result.Answer {
		if a, ok := answer.(*dns.MX); ok {
			results.RecordType = dns.TypeMX
			results.Hosts = append(results.Hosts, a.Mx)
		}
	}
	return results, nil
}

// LookupSRV returns SRV RRs for a zone
func (c *Client) LookupSRV(zone string) (*Results, error) {
	result, err := c.exchange(zone, dns.TypeSRV)
	if err != nil {
		return nil, err
	}

	results := &Results{RequestType: dns.TypeSRV}
	for _, answer := range result.Answer {
		if a, ok := answer.(*dns.SRV); ok {
			results.RecordType = dns.TypeSRV
			results.Hosts = append(results.Hosts, a.Target)
		}
	}
	return results, nil
}

// DoAXFR attempts to execute an AXFR against a domain by first
// getting the domains NS records, and attempts an AXFR against
// each server. We use a waitgroup for NS servers and a workerpool
// for doing resolution on various records returned by AXFR.
func (c *Client) DoAXFR(name string) (map[string][]*Results, error) {
	nsAddrs, err := c.LookupNS(name)
	if err != nil {
		return nil, err
	}

	results := make(map[string][]*Results)

	msg := new(dns.Msg)
	msg.SetAxfr(dns.Fqdn(name))

	rc := make(chan *axfrResultError)
	wg := &sync.WaitGroup{}
	pool := workerpool.New(100)

	wg.Add(len(nsAddrs.Hosts))

	for _, nameserver := range nsAddrs.Hosts {

		if dns.IsFqdn(nameserver) {
			nameserver = nameserver[:len(nameserver)-1]
		}
		go c.doAXFR(msg, nameserver, rc, pool, wg)
	}

	go func() {
		wg.Wait()
		close(rc)
		pool.Stop()
	}()

	for axfrResults := range rc {
		results[axfrResults.NSAddress] = axfrResults.Result
	}
	return results, nil
}

// doAXFR writes results to the rc channel and creates a workerpool waitgroup for each time doAXFR
// is spawned by testing a specific nameserver. The workerpool waitgroup is used to signal each
// RR record has been processed and wait for all records to be resolved. Each RR from an AXFR
// is submitted to the workerpool task as a closure where it can process, then write the results
// to the out chan.
func (c *Client) doAXFR(msg *dns.Msg, nameserver string, rc chan<- *axfrResultError, pool *workerpool.WorkerPool, nswg *sync.WaitGroup) {
	defer nswg.Done()

	results := &axfrResultError{NSAddress: nameserver, Result: make([]*Results, 0)}
	out := make(chan *Results)

	tr := &dns.Transfer{
		DialTimeout: 3 * time.Second,
		ReadTimeout: 3 * time.Second,
	}

	log.Printf("testing nameserver: %s\n", nameserver)

	envelope, err := tr.In(msg, nameserver+":53")
	if err != nil {
		log.Printf("nameserver: %s returned err: %s\n", nameserver, err)
		return
	}
	// workerpoool waitgroup
	wpwg := &sync.WaitGroup{}
	for answer := range envelope {
		if answer.Error != nil {
			continue
		}

		for _, rr := range answer.RR {
			wpwg.Add(1)
			task := func(rr dns.RR, wpwg *sync.WaitGroup, out chan<- *Results) func() {
				return func() {
					if r := c.processAXFRRR(rr); r != nil {
						out <- r
					}
					wpwg.Done()
				}
			}
			pool.Submit(task(rr, wpwg, out))
		}
	}

	go func() {
		wpwg.Wait()
		close(out)
	}()

	// will not exit range until wpwg.Wait() / out is closed.
	for axfrRR := range out {
		results.Result = append(results.Result, axfrRR)
	}

	rc <- results
}

func (c *Client) processAXFRRR(rr dns.RR) *Results {
	axfrResult := &Results{}
	axfrResult.RequestType = dns.TypeAXFR
	axfrResult.RecordType = rr.Header().Rrtype

	switch value := rr.(type) {
	case *dns.NS:
		ips, err := c.ResolveName(value.Ns)
		axfrResult.Hosts = append(axfrResult.Hosts, fqdnTrim(value.Ns))
		if err != nil {
			log.Printf("error resolving NS server: %s\n", err)
			return nil
		}
		for _, resolved := range ips {
			axfrResult.IPs = append(axfrResult.IPs, resolved.IPs...)
		}

	case *dns.CNAME:
		ips, err := c.ResolveName(value.Target)
		if err != nil {
			log.Printf("error resolving CNAME: %s\n", err)
			return nil
		}
		axfrResult.Hosts = append(axfrResult.Hosts, fqdnTrim(value.Hdr.Name))
		for _, resolved := range ips {
			axfrResult.IPs = append(axfrResult.IPs, resolved.IPs...)
		}
	case *dns.SRV:
		ips, err := c.ResolveName(value.Target)
		if err != nil {
			log.Printf("error resolving SRV: %s\n", err)
			return nil
		}
		axfrResult.Hosts = append(axfrResult.Hosts, fqdnTrim(value.Hdr.Name))
		for _, resolved := range ips {
			axfrResult.IPs = append(axfrResult.IPs, resolved.IPs...)
		}
	case *dns.MX:
		ips, err := c.ResolveName(value.Mx)
		if err != nil {
			log.Printf("error resolving MX: %s\n", err)
			return nil
		}
		axfrResult.Hosts = append(axfrResult.Hosts, fqdnTrim(value.Hdr.Name))
		for _, resolved := range ips {
			axfrResult.IPs = append(axfrResult.IPs, resolved.IPs...)
		}
	case *dns.A:
		axfrResult.IPs = append(axfrResult.IPs, value.A.String())
		axfrResult.Hosts = append(axfrResult.Hosts, fqdnTrim(value.Hdr.Name))
	case *dns.AAAA:
		axfrResult.IPs = append(axfrResult.IPs, value.AAAA.String())
		axfrResult.Hosts = append(axfrResult.Hosts, fqdnTrim(value.Hdr.Name))
	case *dns.PTR:
		axfrResult.IPs = append(axfrResult.IPs, value.Hdr.Name)
		axfrResult.Hosts = append(axfrResult.Hosts, fqdnTrim(value.Ptr))
	default:
		log.Printf("unknown type: %s\n", value.String())
		return nil
	}
	return axfrResult
}

// Initiates an exchange with a dns resolver
func (c *Client) exchange(name string, query uint16) (*dns.Msg, error) {
	var result *dns.Msg
	var err error

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(name), query)
	// randomize dns resolver for requests
	server := c.servers[rand.Intn(len(c.servers))]
	for i := 0; i < c.retry; i++ {
		result, _, err = c.client.Exchange(msg, server)
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	// TODO: determine if causes FNs
	//if result.Rcode != dns.RcodeSuccess {
	//	return nil, ErrRcode
	//}

	if len(result.Answer) < 1 {
		return nil, ErrEmptyRecords
	}
	return result, nil
}

// ParseArpa parses an in-addr.arpa or ip6.arpa name to IP address.
func ParseArpa(arpa string) (string, bool) {
	arpa = fqdnTrim(arpa)
	// IPv4
	if strings.LastIndex(arpa, "in-addr.arpa") != -1 {
		return parseIPv4Arpa(arpa)
	} else if strings.LastIndex(arpa, "ip6.arpa") != -1 {
		return parseIPv6Arpa(arpa)
	}
	return "", false
}

// parseIPv4Arpa uses sscanf to ensure we only get integer values for the in-addr.arpa string.
func parseIPv4Arpa(ipv4arpa string) (string, bool) {
	bytes := make([]int, 4)
	n, err := fmt.Sscanf(ipv4arpa, ipv4arpafmt, &bytes[3], &bytes[2], &bytes[1], &bytes[0])
	if err != nil || n != 4 {
		return "", false
	}
	return fmt.Sprintf("%d.%d.%d.%d", bytes[0], bytes[1], bytes[2], bytes[3]), true
}

// parseIPv6Arpa uses sscanf to ensure we only get integer values for the in-addr.arpa string.
func parseIPv6Arpa(ipv4arpa string) (string, bool) {
	bytes := make([]byte, 32)
	n, err := fmt.Sscanf(ipv4arpa, ipv6arpafmt, &bytes[31], &bytes[30], &bytes[29], &bytes[28], &bytes[27], &bytes[26], &bytes[25],
		&bytes[24], &bytes[23], &bytes[22], &bytes[21], &bytes[20], &bytes[19], &bytes[18], &bytes[17], &bytes[16], &bytes[15],
		&bytes[14], &bytes[13], &bytes[12], &bytes[11], &bytes[10], &bytes[9], &bytes[8], &bytes[7], &bytes[6], &bytes[5], &bytes[4],
		&bytes[3], &bytes[2], &bytes[1], &bytes[0])
	if err != nil || n != 32 {
		log.Printf("%d err: %s\n", n, err)
		return "", false
	}
	return toIPv6(bytes), true
}

func toIPv6(in []byte) string {
	out := make([]string, len(in)+8) // 7 : characters
	for i, j := 0, 0; i < len(in); i, j = i+1, j+1 {
		out[j] = string(in[i])
		if i != len(in)-1 && (i+1)%4 == 0 {
			j++
			out[j] = ":"
		}
	}
	return strings.Join(out, "")
}

func fqdnTrim(name string) string {
	if dns.IsFqdn(name) {
		return strings.TrimRight(name, ".")
	}
	return name
}
