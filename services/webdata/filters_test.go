package webdata

import (
	"testing"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/amtest"
)

func TestURLQuery2(t *testing.T) {
	userContext := amtest.CreateUserContext(1, 1)
	filter := &am.WebResponseFilter{
		OrgID:   userContext.GetOrgID(),
		GroupID: 2,
		Filters: &am.FilterType{},
		Start:   0,
		Limit:   1000,
	}
	filter.Filters.AddInt64("after_request_time", time.Now().Add(time.Hour-(24*7)).UnixNano())
	query, args, err := buildURLListFilterQuery(userContext, filter)
	if err != nil {
		t.Fatalf("error: %#v\n", err)
	}
	t.Logf("query: %s\n", query)
	t.Logf("%#v\n", args)
}
