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

	if val, ok := filter.Filters.Bool(am.FilterIgnored); ok {
		p = p.Where(sq.Eq{"sga.ignored": val})
	}

	if val, ok := filter.Filters.Bool(am.FilterWildcard); ok {
		p = p.Where(sq.Eq{"sga.is_wildcard_zone": val})
	}

	if val, ok := filter.Filters.Bool(am.FilterHosted); ok {
		p = p.Where(sq.Eq{"sga.is_hosted_service": val})
	}

	if val, ok := filter.Filters.Int64(am.FilterAfterScannedTime); ok && val != 0 {
		p = p.Where(sq.Gt{"sga.last_scanned_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64(am.FilterBeforeScannedTime); ok && val != 0 {
		p = p.Where(sq.Lt{"sga.last_scanned_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64(am.FilterAfterSeenTime); ok && val != 0 {
		p = p.Where(sq.Gt{"sga.last_seen_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64(am.FilterBeforeSeenTime); ok && val != 0 {
		p = p.Where(sq.Lt{"sga.last_seen_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64(am.FilterAfterDiscoveredTime); ok && val != 0 {
		p = p.Where(sq.Gt{"sga.discovered_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Int64(am.FilterBeforeDiscoveredTime); ok && val != 0 {
		p = p.Where(sq.Lt{"sga.discovered_timestamp": time.Unix(0, val)})
	}

	if val, ok := filter.Filters.Float32(am.FilterAboveConfidence); ok && val != 0 {
		p = p.Where(sq.Gt{"sga.confidence_score": val})
	}

	if val, ok := filter.Filters.Float32(am.FilterBelowConfidence); ok && val != 0 {
		p = p.Where(sq.Lt{"sga.confidence_score": val})
	}

	if val, ok := filter.Filters.Float32(am.FilterEqualsConfidence); ok && val != 0 {
		p = p.Where(sq.Eq{"sga.confidence_score": val})
	}

	if val, ok := filter.Filters.Float32(am.FilterAboveUserConfidence); ok && val != 0 {
		p = p.Where(sq.Gt{"sga.user_confidence_score": val})
	}

	if val, ok := filter.Filters.Float32(am.FilterBelowUserConfidence); ok && val != 0 {
		p = p.Where(sq.Lt{"sga.user_confidence_score": val})
	}

	if val, ok := filter.Filters.Float32(am.FilterEqualsUserConfidence); ok && val != 0 {
		p = p.Where(sq.Eq{"sga.user_confidence_score": val})
	}

	if vals, ok := filter.Filters.Int32s(am.FilterEqualsNSRecord); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Eq{"sga.ns_record": vals[0]})
		} else {
			var equals sq.Or
			for _, val := range vals {
				equals = append(equals, sq.Eq{"sga.ns_record": val})
			}
			p = p.Where(equals)
		}
	}

	if vals, ok := filter.Filters.Int32s(am.FilterNotNSRecord); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.NotEq{"sga.ns_record": vals[0]})
		} else {
			var notEquals sq.Or
			for _, val := range vals {
				notEquals = append(notEquals, sq.NotEq{"sga.ns_record": val})
			}
			p = p.Where(notEquals)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterIPAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Eq{"sga.ip_address": vals[0]})
		} else {
			var equals sq.Or
			for _, val := range vals {
				equals = append(equals, sq.Eq{"sga.ip_address": val})
			}
			p = p.Where(equals)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterNotIPAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.NotEq{"sga.ip_address": vals[0]})
		} else {
			var notEquals sq.Or
			for _, val := range vals {
				notEquals = append(notEquals, sq.NotEq{"sga.ip_address": val})
			}
			p = p.Where(notEquals)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Eq{"sga.host_address": vals[0]})
		} else {
			var equals sq.Or
			for _, val := range vals {
				equals = append(equals, sq.Eq{"sga.host_address": val})
			}
			p = p.Where(equals)
		}

	}

	if vals, ok := filter.Filters.Strings(am.FilterNotHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.NotEq{"sga.host_address": vals[0]})
		} else {
			var notEquals sq.Or
			for _, val := range vals {
				notEquals = append(notEquals, sq.NotEq{"sga.host_address": val})
			}
			p = p.Where(notEquals)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterEndsHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Like{"sga.host_address": fmt.Sprintf("%%%s", vals[0])})
		} else {
			var like sq.Or
			for _, val := range vals {
				like = append(like, sq.Like{"sga.host_address": fmt.Sprintf("%%%s", val)})
			}
			p = p.Where(like)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterNotEndsHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.NotLike{"sga.host_address": fmt.Sprintf("%%%s", vals[0])})
		} else {
			var notLike sq.Or
			for _, val := range vals {
				notLike = append(notLike, sq.NotLike{"sga.host_address": fmt.Sprintf("%%%s", val)})
			}
			p = p.Where(notLike)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterStartsHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Like{"sga.host_address": fmt.Sprintf("%s%%", vals[0])})
		} else {
			var like sq.Or
			for _, val := range vals {
				like = append(like, sq.Like{"sga.host_address": fmt.Sprintf("%s%%", val)})
			}
			p = p.Where(like)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterNotStartsHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.NotLike{"sga.host_address": fmt.Sprintf("%s%%", vals[0])})
		} else {
			var notLike sq.Or
			for _, val := range vals {
				notLike = append(notLike, sq.NotLike{"sga.host_address": fmt.Sprintf("%s%%", val)})
			}
			p = p.Where(notLike)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterContainsHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.Like{"sga.host_address": fmt.Sprintf("%%%s%%", vals[0])})
		} else {
			var like sq.Or
			for _, val := range vals {
				like = append(like, sq.Like{"sga.host_address": fmt.Sprintf("%%%s%%", val)})
			}
			p = p.Where(like)
		}
	}

	if vals, ok := filter.Filters.Strings(am.FilterNotContainsHostAddress); ok && len(vals) > 0 {
		if len(vals) == 1 {
			p = p.Where(sq.NotLike{"sga.host_address": fmt.Sprintf("%%%s%%", vals[0])})
		} else {
			var notLike sq.Or
			for _, val := range vals {
				notLike = append(notLike, sq.NotLike{"sga.host_address": fmt.Sprintf("%%%s%%", val)})
			}
			p = p.Where(notLike)
		}
	}

	p = p.Where(sq.Gt{"sga.address_id": filter.Start}).OrderBy("sga.address_id")
	p = p.Limit(uint64(filter.Limit))
	return p.PlaceholderFormat(sq.Dollar).ToSql()
}
