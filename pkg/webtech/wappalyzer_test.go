package webtech_test

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/browser"
	"github.com/linkai-io/am/pkg/webtech"
)

func testGetAppFile(t *testing.T) []byte {
	resp, err := http.Get("https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/apps.json")
	if err != nil {
		t.Fatalf("error getting latest apps.json file %v\n", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading latest apps.json file %v\n", err)
	}
	return data
}

func TestWappalyzer(t *testing.T) {
	appJSON := testGetAppFile(t)

	w := webtech.NewWappalyzer()
	if err := w.Init(appJSON); err != nil {
		t.Fatalf("error loading wappalyzer data: %v\n", err)
	}

	headers := make(map[string]string, 0)
	headers["server"] = "Apache/2.4.18 (Ubuntu)"
	headers["set-cookie"] = "JSESSIONID=114234, httpOnly"
	res := w.Headers(headers)
	for k, v := range res {
		t.Logf("app: %s", k)
		for _, match := range v {
			t.Logf("matches: %#v", match)
		}
	}

	start := time.Now()
	amazon, err := ioutil.ReadFile("testdata/amazon.co.jp.html")
	t.Logf("%s", time.Now().Sub(start))
	if err != nil {
		t.Fatalf("failed reading amazon test file: %v\n", err)
	}
	start = time.Now()
	res = w.DOM(string(amazon))
	t.Logf("%s", time.Now().Sub(start))
	/*for k, v := range res {
		t.Logf("app: %s", k)
		for _, match := range v {
			t.Logf("matches: %#v", match)
		}
	}*/

	react, err := ioutil.ReadFile("testdata/reactjs.org.html")
	if err != nil {
		t.Fatalf("failed reading amazon test file: %v\n", err)
	}
	start = time.Now()
	res = w.DOM(string(react))
	t.Logf("%s", time.Now().Sub(start))
	for k, v := range res {
		t.Logf("app: %s", k)
		for _, match := range v {
			t.Logf("matches: %#v", match)
		}
	}
}

func TestWappalyzerInject(t *testing.T) {
	appJSON := testGetAppFile(t)

	w := webtech.NewWappalyzer()
	if err := w.Init(appJSON); err != nil {
		t.Fatalf("error loading wappalyzer data: %v\n", err)
	}

	ctx := context.Background()
	b := browser.NewGCDBrowserPool(2, w)
	defer b.Close(ctx)

	if err := b.Init(); err != nil {
		t.Fatalf("error initializing browser: %v\n", err)
	}

	address := &am.ScanGroupAddress{
		HostAddress: "example.com",
		IPAddress:   "93.184.216.34",
	}
	brows := b.Acquire(ctx)
	defer b.Return(ctx, brows)
	ta, err := brows.GetFirstTab()
	if err != nil {
		t.Fatalf("error getting tab: %v\n", err)
	}
	defer brows.CloseTab(ta) // closes websocket go routines

	tab := browser.NewTab(ta, address)
	defer tab.Close()

	if err := tab.LoadPage(ctx, "https://angularjs.org/"); err != nil {
		t.Fatalf("error loading page:%v\n", err)
	}
	start := time.Now()
	result, err := tab.InjectJS(w.JSToInject())
	if err != nil {
		t.Fatalf("error injecting js: %v\n", err)
	}
	jsobjs := w.JSResultsToObjects(result)
	jsMatches := w.JS(jsobjs)

	t.Logf("%s", time.Now().Sub(start))

	dom := tab.SerializeDOM()
	domMatches := w.DOM(dom)

	results := w.MergeMatches([]map[string][]*webtech.Match{domMatches, jsMatches})
	for k, v := range results {
		t.Logf("%s %#v\n", k, v)
	}
}
