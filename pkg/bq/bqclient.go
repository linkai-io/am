package bq

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/option"

	"cloud.google.com/go/bigquery"
	"github.com/linkai-io/am/am"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

// TODO: prior to enabling full queries, use commonname/like query instead of keying off etld.
const query = `SELECT certhash, ARRAY_AGG(STRUCT(
	time as time,
	server as server,
	index as index,
	serialnumber as serialnumber,
	notbefore as notbefore,
	notafter as notafter,
	country as country,
	organization as organization,
	organizationalunit as organizationalunit,
	lower(commonname) as commonname,
	lower(verifieddnsnames) as verifieddnsnames, 
	lower(unverifieddnsnames) as unverifieddnsnames,
	ipaddresses as ipaddresses,
	lower(emailaddresses) as emailaddresses,
	lower(etld) as etld))[OFFSET(0)] as result from %s.%s
	where Time >= @from and lower(etld)=@etld
	GROUP BY certhash`

var firstRunTime = time.Date(2018, time.May, 0, 0, 0, 0, 0, time.Local)

const subDomainQuery = `select commonname from %s.%s where Time >= @from and lower(commonname) like @commonname`

// oddly, this saves money because if we specify Time, it scans the entire Time column for all rows
const subDomainQueryFirstRun = `select commonname from %s.%s where lower(commonname) like @commonname`

var (
	// ErrConfigInvalid when missing required fields.
	ErrConfigInvalid = errors.New("configuration was missing dataset name or table name")
)

// SubdomainResult is just so we can have bigquery struct tags.
type SubdomainResult struct {
	CommonName string `bigquery:"commonname"`
}

// Result of CT record
type Result struct {
	Time               time.Time `bigquery:"time"`
	Server             string    `bigquery:"server"`
	Index              int64     `bigquery:"index"`
	SerialNumber       string    `bigquery:"serialnumber"`
	NotBefore          time.Time `bigquery:"notbefore"`
	NotAfter           time.Time `bigquery:"notafter"`
	Country            string    `bigquery:"country"`
	Organization       string    `bigquery:"organization"`
	OrganizationalUnit string    `bigquery:"organizationalunit"`
	CommonName         string    `bigquery:"commonname"`
	VerifiedDNSNames   string    `bigquery:"verifieddnsnames"`
	UnverifiedDNSNames string    `bigquery:"unverifieddnsnames"`
	IPAddresses        string    `bigquery:"ipaddresses"`
	EmailAddresses     string    `bigquery:"emailaddresses"`
	ETLD               string    `bigquery:"etld"`
}

// ETLDResult containing ct record
type ETLDResult struct {
	CertHash string  `bigquery:"certhash"`
	Result   *Result `bigquery:"result"`
}

// ClientConfig of required fields
type ClientConfig struct {
	ProjectID   string `json:"project_id"`
	DatasetName string `json:"dataset_name"`
	TableName   string `json:"table_name"`
}

// Client for querying BigQuery
type Client struct {
	bqClient               *bigquery.Client
	config                 *ClientConfig
	credentials            []byte
	query                  string
	subDomainQuery         string
	subDomainQueryFirstRun string
}

// NewClient creates a new BigQuery client
func NewClient() *Client {
	return &Client{config: &ClientConfig{}}
}

// Init the BigQuery client by parsing config and calling initBQClient
func (c *Client) Init(config, credentials []byte) error {
	c.credentials = credentials

	if err := json.Unmarshal(config, c.config); err != nil {
		return err
	}
	if c.config.DatasetName == "" || c.config.TableName == "" || c.credentials == nil || len(c.credentials) == 0 {
		return ErrConfigInvalid
	}

	c.query = fmt.Sprintf(query, c.config.DatasetName, c.config.TableName)
	c.subDomainQuery = fmt.Sprintf(subDomainQuery, c.config.DatasetName, c.config.TableName)
	c.subDomainQueryFirstRun = fmt.Sprintf(subDomainQueryFirstRun, c.config.DatasetName, c.config.TableName)

	return c.initBQClient()
}

// QuerySubdomains just extracts the common name from the ct log data table.
func (c *Client) QuerySubdomains(ctx context.Context, from time.Time, etld string) (map[string]*am.CTSubdomain, error) {
	commonNames := make(map[string]*am.CTSubdomain, 0)
	commonName := fmt.Sprintf("%%.%s", strings.ToLower(etld)) // add . so we only get subdomains

	if etld == "" {
		return commonNames, errors.New("empty etld passed to query subdomains")
	}

	var q *bigquery.Query
	if from == firstRunTime {
		q = c.bqClient.Query(c.subDomainQueryFirstRun)
		q.Parameters = []bigquery.QueryParameter{
			{Name: "commonname", Value: commonName},
		}
	} else {
		q := c.bqClient.Query(c.subDomainQuery)
		q.Parameters = []bigquery.QueryParameter{
			{Name: "from", Value: from},
			{Name: "commonname", Value: commonName},
		}
	}

	it, err := q.Read(ctx)
	if err != nil {
		return commonNames, errors.Wrap(err, "failed to read from bigquery client")
	}

	log.Info().Uint64("total_rows", it.TotalRows).Str("etld", etld).Msg("iterating over results")

	for {
		var r SubdomainResult
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}

		if err != nil {
			log.Error().Err(err).Str("etld", etld).Msg("error iterating data")
			break
		}
		// replace *.example.com -> example.com
		subdomain := strings.Replace(strings.Trim(r.CommonName, " "), "*.", "", -1)
		if !strings.HasSuffix(subdomain, etld) {
			log.Warn().Str("subdomain", subdomain).Str("etld", etld).Msg("subdomain did not contain etld")
			continue
		}

		commonNames[subdomain] = &am.CTSubdomain{ETLD: etld, Subdomain: subdomain}
	}

	return commonNames, nil
}

// QueryETLD and return all records from our bigquery table
func (c *Client) QueryETLD(ctx context.Context, from time.Time, etld string) (map[string]*am.CTRecord, error) {
	ctRecords := make(map[string]*am.CTRecord)

	q := c.bqClient.Query(c.query)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "from", Value: from},
		{Name: "etld", Value: etld},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return ctRecords, errors.Wrap(err, "failed to read from bigquery client")
	}

	for {
		ctRecord := &am.CTRecord{}
		var r ETLDResult
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}

		if err != nil {
			log.Error().Err(err).Int("size", len(ctRecords)).Str("etld", etld).Msg("error iterating data")
			break
		}

		ctRecord = CTBigQueryResultToDomain(r.CertHash, r.Result)
		ctRecords[r.CertHash] = ctRecord
	}

	return ctRecords, nil
}

func (c *Client) initBQClient() error {
	var err error

	ctx := context.Background()

	c.bqClient, err = bigquery.NewClient(ctx,
		c.config.ProjectID,
		option.WithCredentialsJSON([]byte(c.credentials)))

	if err != nil {
		return errors.Wrap(err, "failed to initialize bigquery client")
	}

	return nil
}
