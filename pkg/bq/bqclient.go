package bq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/linkai-io/am/am"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
)

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

var (
	// ErrConfigInvalid when missing required fields.
	ErrConfigInvalid = errors.New("configuration was missing dataset name or table name")
)

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
	bqClient *bigquery.Client
	config   *ClientConfig
	query    string
}

// NewClient creates a new BigQuery client
func NewClient() *Client {
	return &Client{config: &ClientConfig{}}
}

// Init the BigQuery client by parsing config and calling initBQClient
func (c *Client) Init(config []byte) error {
	if err := json.Unmarshal(config, c.config); err != nil {
		return err
	}
	if c.config.DatasetName == "" || c.config.TableName == "" {
		return ErrConfigInvalid
	}

	c.query = fmt.Sprintf(query, c.config.DatasetName, c.config.TableName)

	return c.initBQClient()
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	c.bqClient, err = bigquery.NewClient(ctx, c.config.ProjectID)
	if err != nil {
		return errors.Wrap(err, "failed to initialize bigquery client")
	}
	return nil
}
