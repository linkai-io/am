package bigdata

import "fmt"

const (
	commonColumns = `inserted_timestamp,
	etld,
	cert_hash,
	serial_number,
	not_before,
	not_after,
	country,
	organization,
	organizational_unit,
	common_name,
	verified_dns_names,
	unverified_dns_names,
	ip_addresses, 
	email_addresses`
)

var queryMap = map[string]string{
	"getCertificates": fmt.Sprintf(`select query_timestamp,certificate_id,%s from am.certificates as certs
		inner join am.certificate_queries as queries on certs.etld=queries.etld where etld=$1`, commonColumns),
	"insertQuery": `insert into am.certificate_queries (etld, query_timestamp) values ($1, $2) on conflict 
	(etld) do update set query_timestamp=EXCLUDED.query_timestamp`,
}

var (
	AddCTTempTableKey     = "cert_add_temp"
	AddCTTempTableColumns = []string{"inserted_timestamp", "etld", "cert_hash", "serial_number",
		"not_before", "not_after", "country", "organization", "organizational_unit", "common_name",
		"verified_dns_names", "unverified_dns_names", "ip_addresses", "email_addresses"}

	AddCTTempTable = `create temporary table cert_add_temp (
			inserted_timestamp bigint,
			etld varchar(512) not null,
			cert_hash varchar(256) not null unique,
			serial_number varchar(256),
			not_before timestamptz,
			not_after timestamptz,
			country varchar(256),
			organization text,
			organizational_unit text,
			common_name text,
			verified_dns_names text,
			unverified_dns_names text,
			ip_addresses text, 
			email_addresses text
		) on commit drop;`

	AddTempToCT = fmt.Sprintf(`insert into am.certificates as cert (
			%s
		)
		select
			temp.inserted_timestamp,
			temp.etld,
			temp.cert_hash,
			temp.serial_number,
			temp.not_before,
			temp.not_after,
			temp.country,
			temp.organization,
			temp.organizational_unit,
			temp.common_name,
			temp.verified_dns_names,
			temp.unverified_dns_names,
			temp.ip_addresses, 
			temp.email_addresses
		from cert_add_temp as temp on conflict do nothing`, commonColumns)
)
