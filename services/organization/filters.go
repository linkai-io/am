package organization

import (
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/linkai-io/am/am"
)

func buildListFilterQuery(userContext am.UserContext, filter *am.OrgFilter) (string, []interface{}, error) {
	p := sq.Select().Column("organization_id").Columns(strings.Split(defaultColumns, ",")...).From("am.organizations")

	if filter.Filters != nil {
		if val, ok := filter.Filters.String(am.FilterBillingSubscriptionID); ok {
			p = p.Where(sq.Eq{"billing_subscription_id": val})
		}
	}

	p = p.Where(sq.Gt{"organization_id": filter.Start}).OrderBy("organization_id")
	p = p.Limit(uint64(filter.Limit))
	return p.PlaceholderFormat(sq.Dollar).ToSql()
}
