package address

import (
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/linkai-io/am/am"
)

func buildGetFilterQuery(userContext am.UserContext, filter *am.ScanGroupAddressFilter) (string, []interface{}, error) {
	columns := strings.Replace(sharedColumns, "\n\t\t", "", -1)
	p := sq.Select().Columns(strings.Split(columns, ",")...).From("am.scan_group_addresses as sga")

	p = p.Where(sq.Eq{"sga.organization_id": userContext.GetOrgID()}).
		Where(sq.Eq{"sga.scan_group_id": filter.GroupID}).
		Where(sq.Eq{"sga.deleted": false})

	if val, ok := filter.Filters.Bool("ignored"); ok {
		p = p.Where(sq.Eq{"sga.ignored": val})
	}

	if val, ok := filter.Filters.Bool("wildcard"); ok {
		p = p.Where(sq.Eq{"sga.is_wildcard_zone": val})
	}

	if val, ok := filter.Filters.Bool("hosted"); ok {
		p = p.Where(sq.Eq{"sga.is_hosted_service": val})
	}

	if val, ok := filter.Filters.Int64("after_scanned_time"); ok && val != 0 {
		p = p.Where(sq.Gt{"sga.last_scanned_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64("before_scanned_time"); ok && val != 0 {
		p = p.Where(sq.Lt{"sga.last_scanned_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64("after_seen_time"); ok && val != 0 {
		p = p.Where(sq.Gt{"sga.last_seen_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64("before_seen_time"); ok && val != 0 {
		p = p.Where(sq.Lt{"sga.last_seen_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64("after_discovered_time"); ok && val != 0 {
		p = p.Where(sq.Gt{"sga.discovered_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64("before_discovered_time"); ok && val != 0 {
		p = p.Where(sq.Lt{"sga.discovered_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Float32("above_confidence"); ok && val != 0 {
		p = p.Where(sq.Gt{"sga.confidence_score": val})
	}

	if val, ok := filter.Filters.Float32("below_confidence"); ok && val != 0 {
		p = p.Where(sq.Lt{"sga.confidence_score": val})
	}

	if val, ok := filter.Filters.Float32("above_user_confidence"); ok && val != 0 {
		p = p.Where(sq.Gt{"sga.user_confidence_score": val})
	}

	if val, ok := filter.Filters.Float32("below_user_confidence"); ok && val != 0 {
		p = p.Where(sq.Lt{"sga.user_confidence_score": val})
	}

	if vals, ok := filter.Filters.Int32s("ns_record"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Eq{"sga.ns_record": val})
		}
	}

	if vals, ok := filter.Filters.Strings("ip_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Eq{"sga.ip_address": val})
		}
	}

	if vals, ok := filter.Filters.Strings("host_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Eq{"sga.host_address": val})
		}
	}

	if vals, ok := filter.Filters.Strings("ends_host_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Like{"sga.host_address": fmt.Sprintf("%%%s", val)})
		}
	}

	if vals, ok := filter.Filters.Strings("starts_host_address"); ok && len(vals) > 0 {
		for _, val := range vals {
			p = p.Where(sq.Like{"sga.host_address": fmt.Sprintf("%s%%", val)})
		}
	}

	p = p.Where(sq.Gt{"sga.address_id": filter.Start}).OrderBy("sga.address_id")
	p = p.Limit(uint64(filter.Limit))
	return p.PlaceholderFormat(sq.Dollar).ToSql()
}
