package webdata

import (
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/linkai-io/am/am"
	"github.com/rs/zerolog/log"
)

func buildSnapshotQuery(userContext am.UserContext, filter *am.WebSnapshotFilter) (string, []interface{}, error) {
	columns := strings.Replace(snapshotColumns, "\n\t\t", "", -1)
	p := sq.Select().Columns(strings.Split(columns, ",")...).From("am.web_snapshots").
		Where(sq.Eq{"organization_id": filter.OrgID}).
		Where(sq.Eq{"scan_group_id": filter.GroupID}).
		Where(sq.Gt{"snapshot_id": filter.Start})

	if val, ok := filter.Filters.Int64("after_response_time"); ok && val != 0 {
		p = p.Where(sq.Gt{"response_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.String("host_address"); ok && val != "" {
		p = p.Where(sq.Eq{"host_address": val})
	}

	p = p.Limit(uint64(filter.Limit)).PlaceholderFormat(sq.Dollar)

	return p.ToSql()
}

func buildWebFilterQuery(userContext am.UserContext, filter *am.WebResponseFilter) (string, []interface{}, error) {
	columns := strings.Replace(responseColumns, "\n\t\t", "", -1)
	p := sq.Select().Columns(strings.Split(columns, ",")...)

	if latestOnly, _ := filter.Filters.Bool("latest_only"); latestOnly {
		return latestWebResponseFilter(p, userContext, filter)
	}

	p = p.From("am.web_responses as wb").
		Join("am.web_status_text as wst on wb.status_text_id = wst.status_text_id").
		Join("am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id")

	p = p.Where(sq.Eq{"organization_id": userContext.GetOrgID()}).
		Where(sq.Eq{"scan_group_id": filter.GroupID}).
		Where(sq.Gt{"response_id": filter.Start})

	p = webResponseFilterClauses(p, userContext, filter)

	p = p.Limit(uint64(filter.Limit)).PlaceholderFormat(sq.Dollar)
	return p.ToSql()
}

func latestWebResponseFilter(p sq.SelectBuilder, userContext am.UserContext, filter *am.WebResponseFilter) (string, []interface{}, error) {
	sub := sq.Select("web_responses.url").
		Column(sq.Alias(sq.Expr("max(web_responses.url_request_timestamp)"), "url_request_timestamp")).
		From("am.web_responses").
		Where(sq.Eq{"organization_id": userContext.GetOrgID()}).
		Where(sq.Eq{"scan_group_id": filter.GroupID}).
		Where(sq.Gt{"response_id": filter.Start}).GroupBy("url")

	p = webResponseFilterClauses(p, userContext, filter)
	p = p.FromSelect(sub, "latest").
		Join("am.web_responses as wb on wb.url_request_timestamp=latest.url_request_timestamp and wb.url=latest.url").
		Join("am.web_status_text as wst on wb.status_text_id = wst.status_text_id").
		Join("am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id").
		OrderBy("response_id").
		Limit(uint64(filter.Limit)).
		PlaceholderFormat(sq.Dollar)
	return p.ToSql()
}

func webResponseFilterClauses(p sq.SelectBuilder, userContext am.UserContext, filter *am.WebResponseFilter) sq.SelectBuilder {
	if vals, ok := filter.Filters.Strings("mime_type"); ok && len(vals) > 0 {
		args := make([]interface{}, len(vals))
		for i, v := range vals {
			args[i] = v
		}
		p = p.Where("mime_type IN ("+sq.Placeholders(len(vals))+")", args...)
	}

	if vals, ok := filter.Filters.Strings("header_names"); ok && len(vals) > 0 {
		for _, v := range vals {
			p = p.Where("headers ?? "+sq.Placeholders(1), v)
		}
	}

	if vals, ok := filter.Filters.Strings("not_header_names"); ok && len(vals) > 0 {
		for _, v := range vals {
			p = p.Where("not(headers ?? "+sq.Placeholders(1)+")", v)
		}
	}

	if nameValues, ok := filter.Filters.Strings("header_pair_names"); ok && len(nameValues) > 0 {
		if headerValues, ok := filter.Filters.Strings("header_pair_values"); ok && len(headerValues) == len(nameValues) {
			for i := 0; i < len(nameValues); i++ {
				log.Info().Msg("ADDING HEADER")
				p = p.Where("headers->>"+sq.Placeholders(1)+"="+sq.Placeholders(1), nameValues[i], headerValues[i])
			}
		}
	}

	if val, ok := filter.Filters.Int64("after_request_time"); ok && val != 0 {
		p = p.Where(sq.Gt{"wb.url_request_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64("before_request_time"); ok && val != 0 {
		p = p.Where(sq.Lt{"wb.url_request_timestamp": time.Unix(0, val)})
	}

	if vals, ok := filter.Filters.Strings("ip_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Eq{"ip_address": val})
		}
	}

	if vals, ok := filter.Filters.Strings("host_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Eq{"wb.host_address": val})
		}
	}

	if vals, ok := filter.Filters.Strings("ends_host_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Like{"wb.host_address": fmt.Sprintf("%%%s", val)})
		}
	}

	if vals, ok := filter.Filters.Strings("starts_host_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Like{"wb.host_address": fmt.Sprintf("%s%%", val)})
		}
	}

	if vals, ok := filter.Filters.Strings("load_ip_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Eq{"load_ip_address": val})
		}
	}

	if vals, ok := filter.Filters.Strings("load_host_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Eq{"wb.load_host_address": val})
		}
	}

	if vals, ok := filter.Filters.Strings("ends_load_host_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Like{"wb.load_host_address": fmt.Sprintf("%%%s", val)})
		}
	}

	if vals, ok := filter.Filters.Strings("starts_load_host_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Like{"wb.load_host_address": fmt.Sprintf("%s%%", val)})
		}
	}
	return p
}

func buildCertificateFilter(userContext am.UserContext, filter *am.WebCertificateFilter) (string, []interface{}, error) {
	p := sq.Select().Columns(strings.Split(certificateColumns, ",")...).
		From("am.web_certificates as wb").
		Where(sq.Eq{"organization_id": userContext.GetOrgID()}).
		Where(sq.Eq{"scan_group_id": filter.GroupID})

	if val, ok := filter.Filters.Bool("deleted"); ok {
		p = p.Where(sq.Eq{"deleted": val})
	}

	if val, ok := filter.Filters.Int64("after_response_time"); ok && val != 0 {
		p = p.Where(sq.Gt{"after_response_time": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64("before_response_time"); ok && val != 0 {
		p = p.Where(sq.Lt{"before_response_time": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64("after_valid_to"); ok && val != 0 {
		p = p.Where(sq.Gt{"valid_to": val})
	}

	if val, ok := filter.Filters.Int64("before_valid_to"); ok && val != 0 {
		p = p.Where(sq.Lt{"valid_to": val})
	}

	if val, ok := filter.Filters.Int64("after_valid_from"); ok && val != 0 {
		p = p.Where(sq.Gt{"valid_from": val})
	}

	if val, ok := filter.Filters.Int64("before_valid_from"); ok && val != 0 {
		p = p.Where(sq.Lt{"valid_from": val})
	}

	if val, ok := filter.Filters.String("host_address_equals"); ok && val != "" {
		p = p.Where(sq.Eq{"host_address": val})
	}

	p = p.Where(sq.Gt{"certificate_id": filter.Start}).OrderBy("certificate_id").Limit(uint64(filter.Limit))

	return p.PlaceholderFormat(sq.Dollar).ToSql()
}

func buildURLListFilterQuery(userContext am.UserContext, filter *am.WebResponseFilter) (string, []interface{}, error) {
	var latestOnly bool
	latestOnly, _ = filter.Filters.Bool("latest_only")

	p := sq.Select().Columns("wb.organization_id", "wb.scan_group_id", "wb.url_request_timestamp", "load_host_address", "load_ip_address").
		Column("array_agg(wb.url) as urls").
		Column("array_agg(wb.raw_body_link) as raw_body_links").
		Column("array_agg(wb.response_id) as response_ids").
		Column("array_agg((select mime_type from am.web_mime_type where mime_type_id=wb.mime_type_id)) as mime_types")

	if latestOnly {
		sub := sq.Select("url").Column(sq.Alias(sq.Expr("max(url_request_timestamp)"), "url_request_timestamp")).
			From("am.web_responses").
			GroupBy("url")
		p = p.FromSelect(sub, "latest").Join("am.web_responses as wb on wb.url=latest.url and wb.url_request_timestamp=latest.url_request_timestamp").
			Where(sq.Eq{"wb.organization_id": filter.OrgID}).Where(sq.Eq{"wb.scan_group_id": filter.GroupID})
	} else {
		p = p.From("am.web_responses as wb").Where(sq.Eq{"wb.organization_id": filter.OrgID}).Where(sq.Eq{"wb.scan_group_id": filter.GroupID})
	}

	if val, ok := filter.Filters.Int64("after_request_time"); ok && val != 0 {
		p = p.Where(sq.Or{sq.Eq{"wb.url_request_timestamp": "1970-01-01 00:00:00+00"}, sq.Eq{"wb.url_request_timestamp": val}})
	} else {
		log.Info().Msgf("%v %v\n", val, ok)
	}
	p = p.Where(sq.Gt{"wb.response_id": filter.Start})

	if latestOnly {
		p = p.GroupBy("wb.organization_id", "wb.scan_group_id", "load_host_address", "load_ip_address", "wb.url_request_timestamp").OrderBy("wb.url_request_timestamp")
	} else {
		p = p.GroupBy("wb.organization_id", "wb.scan_group_id", "load_host_address", "load_ip_address", "wb.url_request_timestamp").OrderBy("wb.url_request_timestamp")
	}

	p = p.Limit(uint64(filter.Limit))
	return p.PlaceholderFormat(sq.Dollar).ToSql()
}