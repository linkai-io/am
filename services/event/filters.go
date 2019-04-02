package event

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/linkai-io/am/am"
)

func buildGetFilterQuery(userContext am.UserContext, filter *am.EventFilter) (string, []interface{}, error) {
	p := sq.Select().Columns("en.organization_id",
		"en.scan_group_id",
		"en.notification_id",
		"en.type_id",
		"en.event_timestamp",
		"en.event_data").From("am.event_notifications as en").
		Join("lateral (select user_id, type_id, event_timestamp from am.user_notification_subscriptions as uns where uns.subscribed=true and en.type_id=uns.type_id and en.event_timestamp >= uns.subscribed_since) as uns on true").
		Where(sq.Eq{"uns.user_id": userContext.GetUserID()})
		//Where(sq.GtOrEq{"en.event_timestamp": "uns.subscribed_since"})

	includeRead := false
	if val, ok := filter.Filters.Bool("include_read"); ok {
		includeRead = val
	}

	if !includeRead {
		p = p.LeftJoin("am.user_notifications_read as unr on en.notification_id=unr.notification_id").
			Where(sq.Eq{"unr.notification_id": nil})
	}

	p = p.Where(sq.Eq{"en.organization_id": userContext.GetOrgID()}).
		Where(sq.Gt{"en.notification_id": filter.Start})

	if val, ok := filter.Filters.Int32("group_id"); ok {
		p = p.Where(sq.Eq{"en.scan_group_id": val})
	}

	p = p.Limit(uint64(filter.Limit)).PlaceholderFormat(sq.Dollar)
	return p.ToSql()
}
