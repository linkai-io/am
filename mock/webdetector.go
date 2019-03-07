package mock

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/webtech"
)

type Detector struct {
	InitFn      func(config []byte) error
	InitInvoked bool

	JSFn      func(jsObjects []*webtech.JSObject) map[string][]*webtech.Match
	JSInvoked bool

	HeadersFn      func(headers map[string]string) map[string][]*webtech.Match
	HeadersInvoked bool

	DOMFn      func(dom string) map[string][]*webtech.Match
	DOMInvoked bool

	JSToInjectFn      func() string
	JSToInjectInvoked bool

	JSResultsToObjectsFn      func(in interface{}) []*webtech.JSObject
	JSResultsToObjectsInvoked bool

	MergeMatchesFn      func(results []map[string][]*webtech.Match) map[string]*am.WebTech
	MergeMatchesInvoked bool
}

func (d *Detector) Init(config []byte) error {
	d.InitInvoked = true
	return d.InitFn(config)
}

func (d *Detector) JS(jsObjects []*webtech.JSObject) map[string][]*webtech.Match {
	d.JSInvoked = true
	return d.JSFn(jsObjects)
}

func (d *Detector) Headers(headers map[string]string) map[string][]*webtech.Match {
	d.HeadersInvoked = true
	return d.HeadersFn(headers)
}

func (d *Detector) DOM(dom string) map[string][]*webtech.Match {
	d.DOMInvoked = true
	return d.DOMFn(dom)
}

func (d *Detector) JSToInject() string {
	d.JSToInjectInvoked = true
	return d.JSToInjectFn()
}

func (d *Detector) JSResultsToObjects(in interface{}) []*webtech.JSObject {
	d.JSResultsToObjectsInvoked = true
	return d.JSResultsToObjectsFn(in)
}

func (d *Detector) MergeMatches(results []map[string][]*webtech.Match) map[string]*am.WebTech {
	d.MergeMatchesInvoked = true
	return d.MergeMatchesFn(results)
}
