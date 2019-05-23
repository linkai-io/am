package webdata

import (
	"fmt"
	"math"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/linkai-io/am/am"
	"github.com/rs/zerolog/log"
)

var (
	snapshotColumnsList = strings.Replace(snapshotColumns, "\n\t\t", "", -1)
	techColumnsList     = strings.Replace(techColumns, "\n\t\t", "", -1)
	responseColumnsList = strings.Replace(responseColumns, "\n\t\t", "", -1)
)

func buildSnapshotQuery(userContext am.UserContext, filter *am.WebSnapshotFilter) (string, []interface{}, error) {

	if filter.Start == 0 {
		filter.Start = time.Now().Add(time.Hour).UnixNano()
	}

	if val, ok := filter.Filters.String(am.FilterWebTechType); ok && val != "" {
		return buildSnapshotQueryWithTechType(userContext, filter, strings.ToLower(val))
	}

	if val, ok := filter.Filters.String(am.FilterWebDependentHostAddress); ok && val != "" {
		return buildSnapshotQueryWithDomainDep(userContext, filter, strings.ToLower(val))
	}

	p := sq.Select().Columns(strings.Split(snapshotColumnsList, ",")...).
		Columns(strings.Split(techColumnsList, ",")...).
		From("am.web_snapshots as ws").
		LeftJoin("am.web_technologies as wt on ws.snapshot_id=wt.snapshot_id").
		LeftJoin("am.web_techtypes as wtt on wt.techtype_id=wtt.techtype_id").
		Where(sq.Eq{"ws.organization_id": filter.OrgID}).
		Where(sq.Eq{"ws.scan_group_id": filter.GroupID}).
		Where(sq.Lt{"url_request_timestamp": time.Unix(0, filter.Start)})

	if val, ok := filter.Filters.Int64(am.FilterWebAfterResponseTime); ok && val != 0 {
		p = p.Where(sq.GtOrEq{"response_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.String(am.FilterWebEqualsHostAddress); ok && val != "" {
		p = p.Where(sq.Eq{"host_address": val})
	}

	p = p.GroupBy("ws.snapshot_id, ws.organization_id, ws.scan_group_id").
		OrderBy("ws.url_request_timestamp desc").
		Limit(uint64(filter.Limit)).PlaceholderFormat(sq.Dollar)

	return p.ToSql()
}

func buildSnapshotQueryWithTechType(userContext am.UserContext, filter *am.WebSnapshotFilter, techName string) (string, []interface{}, error) {
	// get snapshots that match techname first
	with := sq.Select().Columns("wt.snapshot_id").
		From("am.web_technologies as wt").
		Join("am.web_techtypes as wtt on wt.techtype_id=wtt.techtype_id").
		Where(sq.Eq{"organization_id": filter.OrgID}).
		Where(sq.Eq{"scan_group_id": filter.GroupID}).
		Where(sq.Eq{"lower(wtt.techname)": techName})

	agg := sq.Select().Columns(strings.Split(snapshotColumnsList, ",")...).
		Columns(strings.Split(techColumnsList, ",")...).
		From("snapshots_tech").
		LeftJoin("am.web_snapshots as ws on snapshots_tech.snapshot_id=ws.snapshot_id").
		LeftJoin("am.web_technologies as wt on ws.snapshot_id=wt.snapshot_id").
		LeftJoin("am.web_techtypes as wtt on wt.techtype_id=wtt.techtype_id").
		Where(sq.Eq{"ws.organization_id": filter.OrgID}).
		Where(sq.Eq{"ws.scan_group_id": filter.GroupID}).
		Where(sq.Lt{"url_request_timestamp": time.Unix(0, filter.Start)})

	if val, ok := filter.Filters.Int64(am.FilterWebAfterResponseTime); ok && val != 0 {
		agg = agg.Where(sq.GtOrEq{"response_timestamp": time.Unix(0, val)})
	}

	agg = agg.GroupBy("ws.snapshot_id, ws.organization_id, ws.scan_group_id").
		OrderBy("ws.url_request_timestamp desc").
		Limit(uint64(filter.Limit))

	withSql, withArgs, err := with.ToSql()
	if err != nil {
		return "", nil, err
	}

	aggSql, aggArgs, err := agg.ToSql()
	if err != nil {
		return "", nil, err
	}

	// Hack because sq doesn't support WITH queries, .Prefix will prefix subqueries :<
	pSql, err := sq.Dollar.ReplacePlaceholders(fmt.Sprintf("WITH snapshots_tech AS (%s) %s", withSql, aggSql))
	pArgs := append(withArgs, aggArgs...)
	return pSql, pArgs, err
}

func buildSnapshotQueryWithDomainDep(userContext am.UserContext, filter *am.WebSnapshotFilter, domain string) (string, []interface{}, error) {
	// get snapshots that match techname first

	with := sq.Select().Columns("wr.url_request_timestamp as uts").
		From("am.web_responses as wr").
		Where(sq.Eq{"organization_id": filter.OrgID}).
		Where(sq.Eq{"scan_group_id": filter.GroupID}).
		Where(sq.Eq{"host_address": domain})

	if val, ok := filter.Filters.Int64(am.FilterWebAfterURLRequestTime); ok && val != 0 {
		with = with.Where(sq.GtOrEq{"uts": time.Unix(0, val)})
	}
	with = with.GroupBy("uts").OrderBy("uts desc")

	agg := sq.Select().Columns(strings.Split(snapshotColumnsList, ",")...).
		Columns(strings.Split(techColumnsList, ",")...).
		From("domain_dep").
		LeftJoin("am.web_snapshots as ws on domain_dep.uts=ws.url_request_timestamp").
		LeftJoin("am.web_technologies as wt on ws.snapshot_id=wt.snapshot_id").
		LeftJoin("am.web_techtypes as wtt on wt.techtype_id=wtt.techtype_id").
		Where(sq.Eq{"ws.organization_id": filter.OrgID}).
		Where(sq.Eq{"ws.scan_group_id": filter.GroupID}).
		Where(sq.Lt{"ws.url_request_timestamp": time.Unix(0, filter.Start)})

	agg = agg.GroupBy("ws.snapshot_id, ws.organization_id, ws.scan_group_id").
		OrderBy("ws.url_request_timestamp desc").
		Limit(uint64(filter.Limit))

	withSql, withArgs, err := with.ToSql()
	if err != nil {
		return "", nil, err
	}

	aggSql, aggArgs, err := agg.ToSql()
	if err != nil {
		return "", nil, err
	}

	// Hack because sq doesn't support WITH queries, .Prefix will prefix subqueries :<
	pSql, err := sq.Dollar.ReplacePlaceholders(fmt.Sprintf("WITH domain_dep AS (%s) %s", withSql, aggSql))
	pArgs := append(withArgs, aggArgs...)
	return pSql, pArgs, err
}

func buildWebFilterQuery(userContext am.UserContext, filter *am.WebResponseFilter) (string, []interface{}, error) {
	p := sq.Select().Columns(strings.Split(responseColumnsList, ",")...)

	if filter.Start == 0 {
		filter.Start = math.MaxInt64
	}

	if latestOnly, _ := filter.Filters.Bool(am.FilterWebLatestOnly); latestOnly {
		return latestWebResponseFilter(p, userContext, filter)
	}

	p = p.From("am.web_responses as wb").
		Join("am.web_status_text as wst on wb.status_text_id = wst.status_text_id").
		Join("am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id")

	p = p.Where(sq.Eq{"organization_id": userContext.GetOrgID()}).
		Where(sq.Eq{"scan_group_id": filter.GroupID}).
		Where(sq.Lt{"response_id": filter.Start})

	p = webResponseFilterClauses(p, userContext, filter)

	p = p.OrderBy("response_id desc").
		Limit(uint64(filter.Limit)).PlaceholderFormat(sq.Dollar)
	return p.ToSql()
}

func latestWebResponseFilter(p sq.SelectBuilder, userContext am.UserContext, filter *am.WebResponseFilter) (string, []interface{}, error) {
	sub := sq.Select("web_responses.url").
		Column(sq.Alias(sq.Expr("max(web_responses.url_request_timestamp)"), "url_request_timestamp")).
		From("am.web_responses").
		Where(sq.Eq{"organization_id": userContext.GetOrgID()}).
		Where(sq.Eq{"scan_group_id": filter.GroupID}).
		Where(sq.Lt{"response_id": filter.Start}).GroupBy("url")

	p = webResponseFilterClauses(p, userContext, filter)
	p = p.FromSelect(sub, "latest").
		Join("am.web_responses as wb on wb.url_request_timestamp=latest.url_request_timestamp and wb.url=latest.url").
		Join("am.web_status_text as wst on wb.status_text_id = wst.status_text_id").
		Join("am.web_mime_type as wmt on wb.mime_type_id = wmt.mime_type_id").
		OrderBy("response_id desc").
		Limit(uint64(filter.Limit)).
		PlaceholderFormat(sq.Dollar)
	return p.ToSql()
}

func webResponseFilterClauses(p sq.SelectBuilder, userContext am.UserContext, filter *am.WebResponseFilter) sq.SelectBuilder {
	if vals, ok := filter.Filters.Strings(am.FilterWebMimeType); ok && len(vals) > 0 {
		args := make([]interface{}, len(vals))
		for i, v := range vals {
			args[i] = v
		}
		p = p.Where("mime_type IN ("+sq.Placeholders(len(vals))+")", args...)
	}

	if vals, ok := filter.Filters.Strings(am.FilterWebHeaderNames); ok && len(vals) > 0 {
		for _, v := range vals {
			p = p.Where("headers ?? "+sq.Placeholders(1), v)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterWebNotHeaderNames); ok && len(vals) > 0 {
		for _, v := range vals {
			p = p.Where("not(headers ?? "+sq.Placeholders(1)+")", v)
		}
	}

	if nameValues, ok := filter.Filters.Strings(am.FilterWebHeaderPairNames); ok && len(nameValues) > 0 {
		if headerValues, ok := filter.Filters.Strings(am.FilterWebHeaderPairValues); ok && len(headerValues) == len(nameValues) {
			for i := 0; i < len(nameValues); i++ {
				log.Info().Msg("ADDING HEADER")
				p = p.Where("headers->>"+sq.Placeholders(1)+"="+sq.Placeholders(1), nameValues[i], headerValues[i])
			}
		}
	}

	if val, ok := filter.Filters.Int64(am.FilterWebEqualsURLRequestTime); ok && val != 0 {
		p = p.Where(sq.Eq{"wb.url_request_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64(am.FilterWebEqualsResponseTime); ok && val != 0 {
		p = p.Where(sq.Eq{"wb.response_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64(am.FilterWebAfterURLRequestTime); ok && val != 0 {
		p = p.Where(sq.Gt{"wb.url_request_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64(am.FilterWebBeforeURLRequestTime); ok && val != 0 {
		p = p.Where(sq.Lt{"wb.url_request_timestamp": time.Unix(0, val)})
	}

	if vals, ok := filter.Filters.Strings(am.FilterWebEqualsIPAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Eq{"ip_address": vals[0]})
		} else {
			var equals sq.Or
			for _, val := range vals {
				equals = append(equals, sq.Eq{"ip_address": val})
			}
			p = p.Where(equals)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterWebEqualsHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Eq{"wb.host_address": vals[0]})
		} else {
			var equals sq.Or
			for _, val := range vals {
				equals = append(equals, sq.Eq{"wb.host_address": val})
			}
			p = p.Where(equals)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterWebEqualsLoadIPAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Eq{"load_ip_address": vals[0]})
		} else {
			var equals sq.Or
			for _, val := range vals {
				equals = append(equals, sq.Eq{"load_ip_address": val})
			}
			p = p.Where(equals)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterWebEqualsLoadHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Eq{"wb.load_host_address": vals[0]})
		} else {
			var equals sq.Or
			for _, val := range vals {
				equals = append(equals, sq.Eq{"wb.load_host_address": val})
			}
			p = p.Where(equals)
		}

	}

	if vals, ok := filter.Filters.Strings(am.FilterWebStartsHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Like{"wb.host_address": fmt.Sprintf("%s%%", vals[0])})
		} else {
			var like sq.Or
			for _, val := range vals {
				like = append(like, sq.Like{"wb.host_address": fmt.Sprintf("%s%%", val)})
			}
			p = p.Where(like)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterWebEndsHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Like{"wb.host_address": fmt.Sprintf("%%%s", vals[0])})
		} else {
			var like sq.Or
			for _, val := range vals {
				like = append(like, sq.Like{"wb.host_address": fmt.Sprintf("%%%s", val)})
			}
			p = p.Where(like)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterWebStartsLoadHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Like{"wb.load_host_address": fmt.Sprintf("%s%%", vals[0])})
		} else {
			var like sq.Or
			for _, val := range vals {
				like = append(like, sq.Like{"wb.load_host_address": fmt.Sprintf("%s%%", val)})
			}
			p = p.Where(like)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterWebEndsLoadHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Like{"wb.load_host_address": fmt.Sprintf("%%%s", vals[0])})
		} else {
			var like sq.Or
			for _, val := range vals {
				like = append(like, sq.Like{"wb.load_host_address": fmt.Sprintf("%%%s", val)})
			}
			p = p.Where(like)
		}
	}

	if val, ok := filter.Filters.String(am.FilterWebEqualsServerType); ok && val != "" {
		p = p.Where(sq.Eq{"headers->>'server'": val})
	}

	if val, ok := filter.Filters.String(am.FilterWebEqualsURL); ok && val != "" {
		p = p.Where(sq.Eq{"url": val})
	}

	return p
}

func buildCertificateFilter(userContext am.UserContext, filter *am.WebCertificateFilter) (string, []interface{}, error) {
	p := sq.Select().Columns(strings.Split(certificateColumns, ",")...).
		From("am.web_certificates as wb").
		Where(sq.Eq{"organization_id": userContext.GetOrgID()}).
		Where(sq.Eq{"scan_group_id": filter.GroupID})

	if val, ok := filter.Filters.Bool(am.FilterDeleted); ok {
		p = p.Where(sq.Eq{"deleted": val})
	}

	if val, ok := filter.Filters.Int64(am.FilterWebAfterResponseTime); ok && val != 0 {
		p = p.Where(sq.GtOrEq{"after_response_time": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64(am.FilterWebBeforeResponseTime); ok && val != 0 {
		p = p.Where(sq.LtOrEq{"before_response_time": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64(am.FilterWebAfterValidTo); ok && val != 0 {
		p = p.Where(sq.GtOrEq{"valid_to": val})
	}

	if val, ok := filter.Filters.Int64(am.FilterWebBeforeValidTo); ok && val != 0 {
		p = p.Where(sq.LtOrEq{"valid_to": val})
	}

	if val, ok := filter.Filters.Int64(am.FilterWebAfterValidFrom); ok && val != 0 {
		p = p.Where(sq.GtOrEq{"valid_from": val})
	}

	if val, ok := filter.Filters.Int64(am.FilterWebBeforeValidFrom); ok && val != 0 {
		p = p.Where(sq.LtOrEq{"valid_from": val})
	}

	if val, ok := filter.Filters.String(am.FilterWebEqualsHostAddress); ok && val != "" {
		p = p.Where(sq.Eq{"host_address": val})
	}

	p = p.Where(sq.Gt{"certificate_id": filter.Start}).OrderBy("certificate_id").Limit(uint64(filter.Limit))

	return p.PlaceholderFormat(sq.Dollar).ToSql()
}

func buildURLListFilterQuery(userContext am.UserContext, filter *am.WebResponseFilter) (string, []interface{}, error) {
	var latestOnly bool
	latestOnly, _ = filter.Filters.Bool(am.FilterWebLatestOnly)
	// start high since we are using timestamp as index for start/limit
	if filter.Start == 0 {
		filter.Start = time.Now().Add(time.Hour).UnixNano()
	}
	p := sq.Select().
		Columns("wb.organization_id", "wb.scan_group_id", "wb.url_request_timestamp", "load_host_address", "load_ip_address").
		Column("wb.url").
		Column("wb.raw_body_link").
		Column("wb.response_id").
		Column("(select mime_type from am.web_mime_type where mime_type_id=wb.mime_type_id) as mime_type")

	if latestOnly {
		sub := sq.Select("url").Column(sq.Alias(sq.Expr("max(url_request_timestamp)"), "url_request_timestamp")).
			From("am.web_responses").
			GroupBy("url")
		p = p.FromSelect(sub, "latest").Join("am.web_responses as wb on wb.url=latest.url and wb.url_request_timestamp=latest.url_request_timestamp").
			Where(sq.Eq{"wb.organization_id": filter.OrgID}).Where(sq.Eq{"wb.scan_group_id": filter.GroupID})
	} else {
		if val, ok := filter.Filters.Int64(am.FilterWebEqualsURLRequestTime); ok && val != 0 {
			p = p.Where(sq.Eq{"wb.url_request_timestamp": time.Unix(0, val)})
		}
		p = p.From("am.web_responses as wb").Where(sq.Eq{"wb.organization_id": filter.OrgID}).Where(sq.Eq{"wb.scan_group_id": filter.GroupID})
	}

	if val, ok := filter.Filters.Int64(am.FilterWebAfterURLRequestTime); ok && val != 0 {
		p = p.Where(sq.Or{sq.Eq{"wb.url_request_timestamp": "1970-01-01 00:00:00+00"}, sq.Gt{"wb.url_request_timestamp": time.Unix(0, val)}})
	} else {
		log.Info().Msgf("%v %v\n", val, ok)
	}

	p = p.Where(sq.Lt{"wb.url_request_timestamp": time.Unix(0, filter.Start)})

	agg := sq.Select().
		Columns("resp.organization_id", "resp.scan_group_id", "resp.url_request_timestamp", "resp.load_host_address", "resp.load_ip_address").
		Column("array_agg(resp.url) as urls").
		Column("array_agg(resp.raw_body_link) as raw_body_links").
		Column("array_agg(resp.response_id) as response_ids").
		Column("array_agg(resp.mime_type) as mime_types").
		From("resp").
		GroupBy("resp.organization_id", "resp.scan_group_id", "resp.load_host_address", "resp.load_ip_address", "resp.url_request_timestamp").
		OrderBy("resp.url_request_timestamp desc").
		Limit(uint64(filter.Limit))

	pSql, pArgs, err := p.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return "", nil, err
	}
	aggSql, _, err := agg.ToSql()
	if err != nil {
		return "", nil, err
	}
	// Hack because sq doesn't support WITH queries, .Prefix will prefix subqueries :<
	// TODO: this won't allow parameters in the aggSql (since the placeholder count would be different)
	withSql := fmt.Sprintf("WITH resp AS (%s) %s", pSql, aggSql)
	//withArgs := append(pArgs, aggArgs)
	return withSql, pArgs, nil
	//return p.PlaceholderFormat(sq.Dollar).ToSql()
}

func buildDomainDependencies(userContext am.UserContext, filter *am.WebResponseFilter) (string, []interface{}, error) {
	if filter.Start == 0 {
		filter.Start = time.Now().Add(time.Hour).UnixNano()
	}
	p := sq.Select().
		Columns("wb.organization_id", "wb.scan_group_id", "wb.load_host_address", "wb.host_address", "max(wb.url_request_timestamp)")

	p = p.From("am.web_responses as wb").Where(sq.Eq{"wb.organization_id": filter.OrgID}).Where(sq.Eq{"wb.scan_group_id": filter.GroupID})

	if val, ok := filter.Filters.Int64(am.FilterWebAfterURLRequestTime); ok && val != 0 {
		p = p.Where(sq.Gt{"wb.url_request_timestamp": time.Unix(0, val)})
	}

	p = p.Where(sq.Lt{"wb.url_request_timestamp": time.Unix(0, filter.Start)}).
		GroupBy("wb.organization_id", "wb.scan_group_id", "wb.load_host_address", "wb.host_address").
		OrderBy("max(wb.url_request_timestamp) desc").
		Limit(uint64(filter.Limit))

	return p.PlaceholderFormat(sq.Dollar).ToSql()
}
