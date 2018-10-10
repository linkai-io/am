package certworker

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/linkai-io/am/pkg/retrier"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/parsers"

	"cloud.google.com/go/bigquery"
	"github.com/google/certificate-transparency-go"
	"github.com/google/certificate-transparency-go/jsonclient"
	"github.com/google/certificate-transparency-go/x509"

	"github.com/google/certificate-transparency-go/client"
)

// Result of parsing certificates
type Result struct {
	CertHash           string
	SerialNumber       string
	NotBefore          time.Time
	NotAfter           time.Time
	Country            string
	Organization       string
	OrganizationalUnit string `bigquery:"organizationalunit"`
	CommonName         string
	VerifiedDNSNames   string
	UnverifiedDNSNames string
	IPAddresses        string
	EmailAddresses     string
	ETLD               string
}

// Save method for bigquery
func (r *Result) Save() (map[string]bigquery.Value, string, error) {
	return map[string]bigquery.Value{
		"CertHash":           r.CertHash,
		"SerialNumber":       r.SerialNumber,
		"NotBefore":          r.NotBefore,
		"NotAfter":           r.NotAfter,
		"Country":            r.Country,
		"Organization":       r.Organization,
		"OrganizationalUnit": r.OrganizationalUnit,
		"CommonName":         r.CommonName,
		"VerifiedDNSNames":   r.VerifiedDNSNames,
		"UnverifiedDNSNames": r.UnverifiedDNSNames,
		"IPAddresses":        r.IPAddresses,
		"EmailAddresses":     r.EmailAddresses,
		"ETLD":               r.ETLD,
	}, "", nil
}

// Extractor of CT certificates
type Extractor struct {
	wg            sync.WaitGroup
	maxExtractors int
	client        *client.LogClient
	server        *am.CTServer
	start         int64
	doneCh        chan struct{}
	next          chan int64
	certs         chan *x509.Certificate
	step          int
	maxSteps      int
	uploader      Uploader
}

// NewExtractor initializes our extractor
func NewExtractor(uploader Uploader, server *am.CTServer, extractors int) *Extractor {

	if strings.Contains(server.URL, "google") {
		extractors = extractors * 2
	}

	return &Extractor{
		server:        server,
		uploader:      uploader,
		maxExtractors: extractors,
		start:         server.Index,
		step:          server.Step,
		next:          make(chan int64, extractors),
	}
}

func (e *Extractor) handleClientError(err error) {
	switch e := err.(type) {
	case client.RspError:
		log.Warn().Err(e).Int("status_code", e.StatusCode).Msg("response error")
	case *url.Error:
		log.Warn().Str("op", e.Op).Err(e.Err).Msg("url error")
	}
}

// Run the extraction
func (e *Extractor) Run(ctx context.Context) (*am.CTServer, error) {
	var err error
	e.client, err = client.New("https://"+e.server.URL, e.getClient(), jsonclient.Options{})
	if err != nil {
		return nil, errors.Wrap(err, "error creating client")
	}
	// start extractors
	for i := 0; i < e.maxExtractors; i++ {
		e.wg.Add(1)
		go e.Extract(ctx)
	}

	// spam
	for i := 0; i < e.maxExtractors; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case e.next <- e.start:
			e.start += int64(e.step)
		}
	}
	e.wg.Wait()
	e.server.Index = e.start
	e.server.IndexUpdated = time.Now().UnixNano()
	return e.server, nil

}

// Extract listens on e.next chan for where to start and extracts certs
func (e *Extractor) Extract(ctx context.Context) {
	var err error
	var entries []ct.LogEntry

	logger := log.With().Str("server", e.server.URL).Logger()

	start := <-e.next

	logger.Info().Int64("start", start).Int64("end", start+int64(e.step)).Msg("processing")

	err = retrier.Retry(func() error {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		entries, err = e.client.GetEntries(timeoutCtx, start, start+int64(e.step))
		if err != nil {
			if strings.Contains(err.Error(), "failed to parse") {
				return errors.Wrap(err, "failed to parse certificates")
			}

			if httpErr, ok := err.(client.RspError); ok {
				if httpErr.StatusCode == 400 {
					return errors.Wrap(httpErr, "bad request")
				}
			}

			if len(entries) == 0 {
				return errors.New("emptry entries")
			}

			return errors.Wrap(err, "error getting entries")
		}

		return nil
	})

	for _, entry := range entries {
		if entry.Precert != nil && entry.Precert.TBSCertificate != nil {
			e.ParseCertificate(entry.Precert.TBSCertificate)
		}
		if entry.X509Cert != nil {
			e.ParseCertificate(entry.X509Cert)
		}
	}
	e.wg.Done()
}

// ParseCertificate and extract out what we want.
func (e *Extractor) ParseCertificate(cert *x509.Certificate) {
	raw := cert.Raw

	if raw == nil {
		raw = make([]byte, 0)
	}

	verifiedDomains, unverifiedDomains, ips, emails := e.ExtractAllAddresses(cert)

	etld, err := parsers.GetETLD(cert.Subject.CommonName)
	if err != nil {
		log.Info().Msg("unable to get ETLD of cert.Subject.CommonName, using common name")
		etld = cert.Subject.CommonName
	}

	result := &Result{
		CertHash:           e.Hash(raw),
		SerialNumber:       cert.SerialNumber.String(),
		NotBefore:          cert.NotBefore,
		NotAfter:           cert.NotAfter,
		Country:            strings.Join(cert.Subject.Country, " "),
		CommonName:         cert.Subject.CommonName,
		Organization:       strings.Join(cert.Subject.Organization, " "),
		OrganizationalUnit: strings.Join(cert.Subject.OrganizationalUnit, " "),
		VerifiedDNSNames:   verifiedDomains,
		UnverifiedDNSNames: unverifiedDomains,
		IPAddresses:        ips,
		EmailAddresses:     emails,
		ETLD:               etld,
	}

	e.uploader.Add(result)
}

// ExtractAllAddresses verified domains, unverified, ip addresses and emails.
func (e *Extractor) ExtractAllAddresses(cert *x509.Certificate) (string, string, string, string) {
	verifiedDomains := make([]string, 0)
	unverifiedDomains := make([]string, 0)
	ips := make([]string, 0)
	emails := make([]string, 0)

	// verified domains
	if cert.DNSNames != nil {
		verifiedDomains = append(verifiedDomains, cert.DNSNames...)
	}

	// unverified domains
	if cert.ExcludedDNSDomains != nil {
		unverifiedDomains = append(unverifiedDomains, cert.ExcludedDNSDomains...)
	}

	if cert.PermittedDNSDomains != nil {
		unverifiedDomains = append(unverifiedDomains, cert.PermittedDNSDomains...)
	}

	if cert.URIs != nil {
		for _, uri := range cert.URIs {
			unverifiedDomains = append(unverifiedDomains, uri.Hostname())
		}
	}

	if cert.ExcludedURIDomains != nil {
		unverifiedDomains = append(unverifiedDomains, cert.ExcludedURIDomains...)
	}

	if cert.PermittedURIDomains != nil {
		unverifiedDomains = append(unverifiedDomains, cert.PermittedURIDomains...)
	}

	// emails
	if cert.EmailAddresses != nil {
		emails = append(emails, cert.EmailAddresses...)
	}

	if cert.ExcludedEmailAddresses != nil {
		emails = append(emails, cert.EmailAddresses...)
	}

	if cert.PermittedEmailAddresses != nil {
		emails = append(emails, cert.EmailAddresses...)
	}

	// ips
	if cert.IPAddresses != nil {
		for _, ip := range cert.IPAddresses {
			ips = append(ips, ip.String())
		}
	}

	if cert.PermittedIPRanges != nil {
		for _, ip := range cert.PermittedIPRanges {
			ips = append(ips, ip.String())
		}
	}

	if cert.ExcludedIPRanges != nil {
		for _, ip := range cert.ExcludedIPRanges {
			ips = append(ips, ip.String())
		}
	}

	return strings.Join(verifiedDomains, " "),
		strings.Join(unverifiedDomains, " "),
		strings.Join(ips, " "),
		strings.Join(emails, " ")
}

// Hash the data
func (e *Extractor) Hash(data []byte) string {
	sum := sha1.Sum(data)
	return base64.StdEncoding.EncodeToString(sum[:])
}

func (e *Extractor) getClient() *http.Client {
	transport := &http.Transport{
		TLSHandshakeTimeout:   50 * time.Second,
		DisableKeepAlives:     false,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}
