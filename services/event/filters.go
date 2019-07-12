package event

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/linkai-io/am/am"
)

func buildGetFilterQuery(userContext am.UserContext, filter *am.EventFilter) (string, []interface{}, error) {
	sub := sq.Select().Columns(
		"organization_id",
		"user_id",
		"type_id",
		"subscribed_since",
		"subscribed").From("am.user_notification_subscriptions").Where(sq.Eq{"organization_id": userContext.GetOrgID()}).
		Where(sq.Eq{"user_id": userContext.GetUserID()}).
		Where(sq.Eq{"subscribed": true})

	p := sq.Select().Columns("subs.organization_id",
		"sg.scan_group_id",
		"events.notification_id",
		"subs.type_id",
		"events.event_timestamp",
		"events.event_data").FromSelect(sub, "subs").
		Join("am.event_notifications as events on subs.type_id=events.type_id and events.organization_id=subs.organization_id and events.event_timestamp >= subs.subscribed_since").
		Join("am.scan_group as sg on events.scan_group_id=sg.scan_group_id and events.organization_id=sg.organization_id").
		Where(sq.Eq{"sg.deleted": false})
		/*
			p := sq.Select().Columns("en.organization_id",
				"en.scan_group_id",
				"en.notification_id",
				"en.type_id",
				"en.event_timestamp",
				"en.event_data").From("am.event_notifications as en").
				Join("lateral (select user_id, type_id, event_timestamp from am.user_notification_subscriptions as uns where uns.subscribed=true and en.type_id=uns.type_id and en.event_timestamp >= uns.subscribed_since) as uns on true").
				Where(sq.Eq{"uns.user_id": userContext.GetUserID()})
				//Where(sq.GtOrEq{"en.event_timestamp": "uns.subscribed_since"})
		*/

	includeRead := false
	if val, ok := filter.Filters.Bool("include_read"); ok {
		includeRead = val
	}

	if !includeRead {
		p = p.LeftJoin("am.user_notifications_read as user_read on events.notification_id=user_read.notification_id").
			Where(sq.Eq{"user_read.notification_id": nil})
	}

	p = p.Where(sq.Eq{"events.organization_id": userContext.GetOrgID()}).
		Where(sq.Gt{"events.notification_id": filter.Start})

	if val, ok := filter.Filters.Int32(am.FilterEventGroupID); ok {
		p = p.Where(sq.Eq{"events.scan_group_id": val})
	}

	p = p.OrderBy("events.event_timestamp desc").Limit(uint64(filter.Limit)).PlaceholderFormat(sq.Dollar)
	return p.ToSql()
}
