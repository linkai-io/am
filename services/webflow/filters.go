package webflow

import (
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/linkai-io/am/am"
)

func buildGetResultsQuery(userContext am.UserContext, filter *am.CustomWebFilter) (string, []interface{}, error) {

	if filter.Start == 0 {
		filter.Start = time.Now().Add(time.Hour).UnixNano()
	}

	p := sq.Select().Columns(strings.Split(resultsColumnsList, ",")...).
		From("am.custom_web_flow_results as r").
		Where(sq.Eq{"r.organization_id": filter.OrgID}).
		Where(sq.Eq{"r.scan_group_id": filter.GroupID}).
		Where(sq.Eq{"r.web_flow_id": filter.WebFlowID}).
		Where(sq.Lt{"r.response_timestamp": time.Unix(0, filter.Start)})

	if val, ok := filter.Filters.Int64("after_response_time"); ok && val != 0 {
		p = p.Where(sq.GtOrEq{"response_timestamp": time.Unix(0, val)})
	}

	p = p.OrderBy("r.response_timestamp desc").
		Limit(uint64(filter.Limit)).PlaceholderFormat(sq.Dollar)

	return p.ToSql()
}
