package bigdata

import "fmt"

const (
	commonColumns = `inserted_timestamp,
	server_name,
	server_index,
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
	"getCertificates": `select queries.query_timestamp,
			certs.certificate_id,
			certs.inserted_timestamp,
			certs.server_name,
			certs.server_index,
			certs.etld,
			certs.cert_hash,
			certs.serial_number,
			certs.not_before,
			certs.not_after,
			certs.country,
			certs.organization,
			certs.organizational_unit,
			certs.common_name,
			certs.verified_dns_names,
			certs.unverified_dns_names,
			certs.ip_addresses, 
			certs.email_addresses from am.certificates as certs
		inner join am.certificate_queries as queries on 
			certs.etld=queries.etld where certs.etld=$1`,

	"getSubdomains": `select queries.query_timestamp,
			subdomains.subdomain_id,
			queries.etld,
			subdomains.inserted_timestamp,
			subdomains.common_name from am.certificate_subdomains as subdomains
		inner join am.certificate_queries_subdomains as queries on subdomains.etld_id=queries.etld_id where queries.etld=$1`,

	"insertQuery": `insert into am.certificate_queries (etld, query_timestamp) values ($1, $2) on conflict 
	(etld) do update set query_timestamp=EXCLUDED.query_timestamp`,

	"insertSubDomainsQuery": `insert into am.certificate_queries_subdomains (etld, query_timestamp) values ($1, $2) on conflict
	(etld) do update set query_timestamp=EXCLUDED.query_timestamp returning etld_id`,

	"deleteQuery": "delete from am.certificate_queries where etld=$1",
	"deleteETLD":  "delete from am.certificates where etld=$1",

	"deleteSubdomains": "delete from am.certificate_queries_subdomains where etld=$1",
}

var (
	// Subdomain table data
	AddCTSubDomainTempTableKey     = "cert_subdomain_temp"
	AddCTSubDomainTempTableColumns = []string{"inserted_timestamp", "etld_id", "common_name"}
	AddCTSubDomainTempTable        = `create temporary table cert_subdomain_temp (
		inserted_timestamp bigint,
		etld_id integer not null,
		common_name text) on commit drop;`
	AddTempSubDomainToCTSubDomain = `insert into am.certificate_subdomains as cert (
			inserted_timestamp,
			etld_id,
			common_name
		)
		select 
			temp.inserted_timestamp,
			temp.etld_id,
			temp.common_name from cert_subdomain_temp as temp on conflict do nothing;`

	// Full CT data
	AddCTTempTableKey     = "cert_add_temp"
	AddCTTempTableColumns = []string{"inserted_timestamp", "server_name", "server_index", "etld", "cert_hash", "serial_number",
		"not_before", "not_after", "country", "organization", "organizational_unit", "common_name",
		"verified_dns_names", "unverified_dns_names", "ip_addresses", "email_addresses"}

	AddCTTempTable = `create temporary table cert_add_temp (
			inserted_timestamp bigint,
			server_name varchar(512) not null,
			server_index bigint not null,
			etld varchar(512) not null,
			cert_hash varchar(256) not null unique,
			serial_number varchar(256),
			not_before bigint,
			not_after bigint,
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
			temp.server_name,
			temp.server_index,
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
