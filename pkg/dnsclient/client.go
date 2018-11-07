package dnsclient

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/linkai-io/am/pkg/generators"

	"github.com/gammazero/workerpool"
	"github.com/linkai-io/am/pkg/parsers"
	"github.com/linkai-io/am/pkg/retrier"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
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

// Client to resolve hosts and ip addresses
type Client struct {
	client  *dns.Client
	servers []string
	retry   uint
}

// New returns a new DNS client
func New(servers []string, retry int) *Client {
	c := &Client{}
	c.servers = servers
	c.retry = uint(retry)
	c.client = &dns.Client{Timeout: 5 * time.Second}
	return c
}

// IsWildcard tests if a domain is a wildcard domain, attempt 10 A and 10 AAAA queries
// of randomly generated names, if we get even a single response it's probably a wildcard
func (c *Client) IsWildcard(ctx context.Context, domain string) bool {
	attempts := 10
	recvd := 0
	resultErrors := make(chan *resultError, attempts)

	ctxDeadline, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	for i := 0; i < attempts/2; i++ {
		randomSubDomain := generators.InsecureAlphabetString(8)
		go c.queryA(ctxDeadline, randomSubDomain+"."+domain, resultErrors)
		go c.queryAAAA(ctxDeadline, randomSubDomain+"."+domain, resultErrors)
	}

	for {
		select {
		case <-ctxDeadline.Done():
			log.Ctx(ctx).Info().Msg("wildcard domain test timeout")
			break
		case r := <-resultErrors:
			recvd++
			if r.Result != nil {
				log.Ctx(ctx).Info().Strs("results", r.Result.IPs).Msg("got wildcard domain result")
				return true
			}

			if recvd == attempts {
				close(resultErrors)
				return false
			}
		}
	}
}

// ResolveName attempts to resolve a name to ip addresses. It will
// attempt to resolve both IPv4 and IPv6 addresses to the name
func (c *Client) ResolveName(ctx context.Context, name string) ([]*Results, error) {
	recvd := 0
	results := make([]*Results, 0)
	resultErrors := make(chan *resultError, 2)

	ctxDeadline, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	go c.queryA(ctxDeadline, name, resultErrors)
	go c.queryAAAA(ctxDeadline, name, resultErrors)

	for {
		select {
		case <-ctxDeadline.Done():
			return results, ctxDeadline.Err()
		case r := <-resultErrors:
			recvd++
			if r.Error == nil {
				results = append(results, r.Result)
			}
			if recvd == 2 {
				close(resultErrors)
				return results, nil
			}
		}
	}
}

func (c *Client) queryA(ctx context.Context, name string, rc chan *resultError) {
	result, err := c.exchange(ctx, name, dns.TypeA)
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

func (c *Client) queryAAAA(ctx context.Context, name string, rc chan *resultError) {
	result, err := c.exchange(ctx, name, dns.TypeAAAA)
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
func (c *Client) ResolveIP(ctx context.Context, ip string) (*Results, error) {
	name, err := dns.ReverseAddr(ip)
	if err != nil {
		return nil, ErrInvalidIP
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	result, err := c.exchange(ctxDeadline, name, dns.TypePTR)
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
func (c *Client) LookupNS(ctx context.Context, zone string) (*Results, error) {
	ctxDeadline, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	result, err := c.exchange(ctxDeadline, zone, dns.TypeNS)
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
func (c *Client) LookupMX(ctx context.Context, zone string) (*Results, error) {
	ctxDeadline, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	result, err := c.exchange(ctxDeadline, zone, dns.TypeMX)
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
func (c *Client) LookupSRV(ctx context.Context, zone string) (*Results, error) {
	ctxDeadline, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	result, err := c.exchange(ctxDeadline, zone, dns.TypeSRV)
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
func (c *Client) DoAXFR(ctx context.Context, name string) (map[string][]*Results, error) {
	nsAddrs, err := c.LookupNS(ctx, name)
	if err != nil {
		return nil, err
	}

	ctxDeadline, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

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
		go c.doAXFR(ctxDeadline, msg, nameserver, rc, pool, wg)
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
func (c *Client) doAXFR(ctx context.Context, msg *dns.Msg, nameserver string, rc chan<- *axfrResultError, pool *workerpool.WorkerPool, nswg *sync.WaitGroup) {
	defer nswg.Done()

	results := &axfrResultError{NSAddress: nameserver, Result: make([]*Results, 0)}
	out := make(chan *Results)

	tr := &dns.Transfer{
		DialTimeout: 3 * time.Second,
		ReadTimeout: 3 * time.Second,
	}

	envelope, err := tr.In(msg, nameserver+":53")
	if err != nil {
		log.Error().Err(err).Str("nameserver", nameserver).Msg("nameserver returned error")
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
			task := func(resourceRecord dns.RR, wpwg *sync.WaitGroup, out chan<- *Results) func() {
				return func() {
					if rec := c.processAXFRRR(ctx, resourceRecord); rec != nil {
						out <- rec
					}
					wpwg.Done()
				}
			}
			r := rr // capture since we are providing it to a closure.
			pool.Submit(task(r, wpwg, out))
		}
	}

	go func() {
		wpwg.Wait()
		close(out)
		pool.Stop()
	}()

	// will not exit range until wpwg.Wait() / out is closed.
	for axfrRR := range out {
		results.Result = append(results.Result, axfrRR)
	}

	rc <- results
}

func (c *Client) processAXFRRR(ctx context.Context, rr dns.RR) *Results {
	axfrResult := &Results{}
	axfrResult.RequestType = dns.TypeAXFR
	axfrResult.RecordType = rr.Header().Rrtype

	switch value := rr.(type) {
	case *dns.NS:
		ips, err := c.ResolveName(ctx, value.Ns)
		axfrResult.Hosts = append(axfrResult.Hosts, parsers.FQDNTrim(value.Ns))
		if err != nil {
			log.Error().Err(err).Msg("error resolving NS server")
			return nil
		}
		for _, resolved := range ips {
			axfrResult.IPs = append(axfrResult.IPs, resolved.IPs...)
		}

	case *dns.CNAME:
		ips, err := c.ResolveName(ctx, value.Target)
		if err != nil {
			log.Error().Err(err).Msg("error resolving CNAME")
			return nil
		}
		axfrResult.Hosts = append(axfrResult.Hosts, parsers.FQDNTrim(value.Hdr.Name))
		for _, resolved := range ips {
			axfrResult.IPs = append(axfrResult.IPs, resolved.IPs...)
		}
	case *dns.SRV:
		ips, err := c.ResolveName(ctx, value.Target)
		if err != nil {
			log.Error().Err(err).Msg("error resolving SRV")
			return nil
		}
		axfrResult.Hosts = append(axfrResult.Hosts, parsers.FQDNTrim(value.Hdr.Name))
		for _, resolved := range ips {
			axfrResult.IPs = append(axfrResult.IPs, resolved.IPs...)
		}
	case *dns.MX:
		ips, err := c.ResolveName(ctx, value.Mx)
		if err != nil {
			log.Error().Err(err).Msg("error resolving MX")
			return nil
		}
		axfrResult.Hosts = append(axfrResult.Hosts, parsers.FQDNTrim(value.Hdr.Name))
		for _, resolved := range ips {
			axfrResult.IPs = append(axfrResult.IPs, resolved.IPs...)
		}
	case *dns.A:
		axfrResult.IPs = append(axfrResult.IPs, value.A.String())
		axfrResult.Hosts = append(axfrResult.Hosts, parsers.FQDNTrim(value.Hdr.Name))
	case *dns.AAAA:
		axfrResult.IPs = append(axfrResult.IPs, value.AAAA.String())
		axfrResult.Hosts = append(axfrResult.Hosts, parsers.FQDNTrim(value.Hdr.Name))
	case *dns.PTR:
		axfrResult.IPs = append(axfrResult.IPs, value.Hdr.Name)
		axfrResult.Hosts = append(axfrResult.Hosts, parsers.FQDNTrim(value.Ptr))
	default:
		log.Warn().Str("unknown_type", value.String()).Msg("unable to resolve")
		return nil
	}
	return axfrResult
}

// Initiates an exchange with a dns resolver
func (c *Client) exchange(ctx context.Context, name string, query uint16) (*dns.Msg, error) {
	var result *dns.Msg
	var err error

	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(name), query)
	// randomize dns resolver for requests
	server := c.servers[rand.Intn(len(c.servers))]
	err = retrier.RetryAttempts(func() error {
		result, _, err = c.client.ExchangeContext(ctx, msg, server)
		return err
	}, c.retry)

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
