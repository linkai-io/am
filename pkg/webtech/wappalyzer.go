package webtech

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"time"

	"text/template"

	"github.com/gammazero/workerpool"
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/parsers"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
)

const jsInjectionTemplate = `function run() {
	var data = [{{range .}} {app: '{{.App}}', obj: '{{.Obj}}', regex: '{{.Regex}}', value: ''},{{end}}{value:''}];
	results = data.map(k => {
	  try {
		k.value = eval(k.obj).toString();
		if (k.value === undefined) { k.value = ''; } else if (k.value.length > 50) { k.value = k.value.substr(0, 50);}
	  } catch(e) {
	  } finally {
		return k;
	  }
	}).filter(k => k.value != '');
	return results;
  };
  run();`

var (
	// AppDefs provides access to the unmarshalled apps.json file

	timeout = 8 * time.Second
)

// Result type encapsulates the result information from a given host
type Result struct {
	Host     string        `json:"host"`
	Matches  []Match       `json:"matches"`
	Duration time.Duration `json:"duration"`
	Error    error         `json:"error"`
}

// App type encapsulates all the data about an App from apps.json
type App struct {
	Cats     StringArray       `json:"cats"`
	CatNames []string          `json:"category_names"`
	Cookies  map[string]string `json:"cookies"`
	Headers  map[string]string `json:"headers"`
	Meta     map[string]string `json:"meta"`
	JS       map[string]string `json:"js"`
	HTML     StringArray       `json:"html"`
	Script   StringArray       `json:"script"`
	URL      StringArray       `json:"url"`
	Website  string            `json:"website"`
	Implies  StringArray       `json:"implies"`
	Icon     string            `json:"icon"`

	HTMLRegex   []AppRegexp `json:"-"`
	JSRegex     []AppRegexp `json:"-"`
	ScriptRegex []AppRegexp `json:"-"`
	URLRegex    []AppRegexp `json:"-"`
	HeaderRegex []AppRegexp `json:"-"`
	MetaRegex   []AppRegexp `json:"-"`
	CookieRegex []AppRegexp `json:"-"`
}

// Category names defined by wappalyzer
type Category struct {
	Name string `json:"name"`
}

// AppsDefinition type encapsulates the json encoding of the whole apps.json file
type AppsDefinition struct {
	Apps map[string]App      `json:"apps"`
	Cats map[string]Category `json:"categories"`
}

type JSObject struct {
	App   string `json:"app"`
	Obj   string `json:"obj"`
	Regex string `json:"regex"`
	Value string `json:"value"`
}

// Wappalyzer for finding technology in http responses/data
type Wappalyzer struct {
	definitions     *AppsDefinition
	headerDetect    map[string][]AppRegexp // k: lowered header names
	htmlDetect      map[string][]AppRegexp
	cookieDetect    map[string][]AppRegexp // k: cookie name
	scriptTagDetect map[string][]AppRegexp
	metaTagDetect   map[string][]AppRegexp
	JSObjects       []JSObject
	jsInject        string
}

// NewWappalyzer for finding technology in http responses/data
func NewWappalyzer() *Wappalyzer {
	return &Wappalyzer{
		headerDetect:    make(map[string][]AppRegexp),
		htmlDetect:      make(map[string][]AppRegexp),
		cookieDetect:    make(map[string][]AppRegexp),
		metaTagDetect:   make(map[string][]AppRegexp),
		scriptTagDetect: make(map[string][]AppRegexp),
		JSObjects:       make([]JSObject, 0),
	}
}

// Init downloads and initializes the application definitions
func (w *Wappalyzer) Init(config []byte) error {
	if err := w.load(config); err != nil {
		return err
	}

	tpl := &bytes.Buffer{}
	templ := template.Must(template.New("jsinject").Parse(jsInjectionTemplate))
	templ.Execute(tpl, w.JSObjects)
	w.jsInject = tpl.String()
	return nil
}

func (w *Wappalyzer) JSToInject() string {
	return w.jsInject
}

// JSResultsToObjects takes in an interface, and marshals it to a slice
// of JSObjects
func (w *Wappalyzer) JSResultsToObjects(in interface{}) []*JSObject {
	results := make([]*JSObject, 0)
	d, err := json.Marshal(in)
	if err != nil {
		log.Warn().Err(err).Msg("unable to marshal js inject results")
		return nil
	}
	if err := json.Unmarshal(d, &results); err != nil {
		log.Warn().Err(err).Msg("unable to unmarshal js inject results")
		return nil
	}

	return results
}

func (w *Wappalyzer) load(data []byte) error {
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	if err := decoder.Decode(&w.definitions); err != nil {
		return err
	}

	for techName, value := range w.definitions.Apps {
		app := w.definitions.Apps[techName]
		app.HTMLRegex = compileRegexes(techName, value.HTML)
		app.ScriptRegex = compileRegexes(techName, value.Script)
		app.URLRegex = compileRegexes(techName, value.URL)
		app.JSRegex = compileNamedRegexes(techName, app.JS)
		app.HeaderRegex = compileNamedRegexes(techName, app.Headers)
		app.MetaRegex = compileNamedRegexes(techName, app.Meta)
		app.CookieRegex = compileNamedRegexes(techName, app.Cookies)

		app.CatNames = make([]string, 0)

		for _, jsInject := range app.JSRegex {
			w.JSObjects = append(w.JSObjects, JSObject{
				App:   jsInject.AppName,
				Obj:   jsInject.Name,
				Value: "",
				Regex: jsInject.Regexp.String(),
			})
		}

		for _, cid := range app.Cats {
			if category, ok := w.definitions.Cats[string(cid)]; ok && category.Name != "" {
				app.CatNames = append(app.CatNames, category.Name)
			}
		}

		for _, headerRegex := range app.HeaderRegex {
			headerName := strings.ToLower(headerRegex.Name)
			if _, ok := w.headerDetect[headerName]; !ok {
				w.headerDetect[headerName] = make([]AppRegexp, 0)
			}
			w.headerDetect[headerName] = append(w.headerDetect[headerName], headerRegex)
		}

		for _, cookieRegex := range app.CookieRegex {
			if _, ok := w.cookieDetect[cookieRegex.Name]; !ok {
				w.cookieDetect[cookieRegex.Name] = make([]AppRegexp, 0)
			}
			w.cookieDetect[cookieRegex.Name] = append(w.cookieDetect[cookieRegex.Name], cookieRegex)
		}

		for _, htmlRegex := range app.HTMLRegex {
			w.htmlDetect[htmlRegex.Name] = append(w.htmlDetect[htmlRegex.Name], htmlRegex)
		}

		for _, scriptRegex := range app.ScriptRegex {
			w.scriptTagDetect[scriptRegex.AppName] = append(w.scriptTagDetect[scriptRegex.AppName], scriptRegex)
		}

		for _, metaRegex := range app.MetaRegex {
			if _, ok := w.metaTagDetect[metaRegex.Name]; !ok {
				w.metaTagDetect[metaRegex.Name] = make([]AppRegexp, 0)
			}
			w.metaTagDetect[metaRegex.Name] = append(w.metaTagDetect[metaRegex.Name], metaRegex)
		}

		w.definitions.Apps[techName] = app
	}
	return nil
}

func (w *Wappalyzer) AppDefinitions() *AppsDefinition {
	return w.definitions
}

// Headers finds matches in headers/cookies
func (w *Wappalyzer) Headers(headers map[string]string) map[string][]*Match {
	results := make(map[string][]*Match, 0)

	for headerName, detect := range w.headerDetect {
		hn := strings.ToLower(headerName)
		if _, ok := w.headerDetect[hn]; !ok {
			continue
		}
		w.searchHeader(headerName, headers[headerName], detect, results)
	}

	cookies, ok := headers["set-cookie"]
	if !ok {
		return results
	}

	if cookies == "" {
		return results
	}

	for _, cookie := range parsers.ParseCookies(cookies) {
		w.searchCookie(cookie, results)
	}

	return results
}

func (w *Wappalyzer) searchCookie(cookie *http.Cookie, results map[string][]*Match) {
	if cookie == nil {
		return
	}

	// we just look to see if cookie name exists, if so we create a match
	for cookieDetectName, detect := range w.cookieDetect {
		if cookie.Name != cookieDetectName {
			continue
		}

		for _, search := range detect {
			if _, ok := results[search.AppName]; !ok {
				results[search.AppName] = make([]*Match, 0)
			}
			results[search.AppName] = append(results[search.AppName], &Match{MatchLocation: "cookie", AppName: search.AppName, Matches: [][]string{[]string{cookie.Name}}})
		}
	}
}

func (w *Wappalyzer) searchHeader(headerName, headerValue string, detect []AppRegexp, results map[string][]*Match) {
	if headerValue == "" {
		return
	}

	matchLocation := "header"
	if m := findMatches(headerValue, detect, matchLocation); len(m) > 0 {
		for _, v := range m {
			if _, ok := results[v.AppName]; !ok {
				results[v.AppName] = make([]*Match, 0)
			}
			results[v.AppName] = append(results[v.AppName], v)
		}
	}
}

// JS parses out the js objects to create a map of result matches
func (w *Wappalyzer) JS(jsObjects []*JSObject) map[string][]*Match {
	results := make(map[string][]*Match, 0)
	location := "javascript"
	if jsObjects == nil {
		return results
	}

	for _, obj := range jsObjects {
		if _, ok := results[obj.App]; !ok {
			results[obj.App] = make([]*Match, 0)
		}

		match := &Match{MatchLocation: location, AppName: obj.App}

		// skip trying to find versions if regex's are empty
		if obj.Regex == "" || obj.Regex == ".*" {
			results[obj.App] = append(results[obj.App], match)
			continue
		}

		if m := findMatches(obj.Value, w.definitions.Apps[obj.App].JSRegex, location); len(m) > 0 {
			for _, am := range m {
				log.Info().Msgf("%s %#v", obj.App, am)
			}
			results[obj.App] = append(results[obj.App], m...)
		} else {
			// no 'regex' matches, but we got a value from our js test so just add we found the app (but don't know version)
			results[obj.App] = append(results[obj.App], match)
		}
	}
	return results
}

// MergeMatches all matches found and returns a map of WebTech results
func (w *Wappalyzer) MergeMatches(results []map[string][]*Match) map[string]*am.WebTech {
	webTech := make(map[string]*am.WebTech, 0)
	for _, result := range results {
		for app, matches := range result {
			if _, ok := webTech[app]; !ok {
				webTech[app] = &am.WebTech{}
			}

			// this web tech app already has a match inserted into it
			if webTech[app].Version != "" {
				continue
			}

			// assume we have at least 1 match for this app
			for _, match := range matches {
				// just take any location in the event we don't have a version
				webTech[app].Location = match.MatchLocation
				b := strings.Builder{}
				if match.Matches != nil {
					for _, sl := range match.Matches {
						b.WriteString(strings.Join(sl, ","))
					}
				}
				webTech[app].Matched = b.String()
				// we *do* have a version so set this webtech app to match the version/location and exit matches
				if match.Version != "" {
					webTech[app].Version = match.Version
					webTech[app].Location = match.MatchLocation
					break
				}
			}
		}
	}
	return webTech
}

// DOM searches the pre-serialized dom and script/meta tags to extract matches
func (w *Wappalyzer) DOM(dom string) map[string][]*Match {
	results := make(map[string][]*Match, 0)
	matchLocation := "html"

	for _, detects := range w.htmlDetect {
		if m := findMatchesPool(dom, detects, matchLocation); len(m) > 0 {
			for _, v := range m {
				if _, ok := results[v.AppName]; !ok {
					results[v.AppName] = make([]*Match, 0)
				}
				results[v.AppName] = append(results[v.AppName], v)
			}
		}
	}

	fn := func(n *html.Node) error {
		if n.Type != html.ElementNode {
			return nil
		}
		switch n.Data {
		case "script":
			src := ""
			for _, attr := range n.Attr {
				if attr.Key == "src" {
					src = attr.Val
				}
			}
			if src == "" {
				return nil
			}

			for _, detects := range w.scriptTagDetect {
				if m := findMatches(src, detects, n.Data); len(m) > 0 {
					for _, v := range m {
						if _, ok := results[v.AppName]; !ok {
							results[v.AppName] = make([]*Match, 0)
						}
						results[v.AppName] = append(results[v.AppName], v)
					}
				}
			}
		case "meta":
			for name, detects := range w.metaTagDetect {
				content := ""
				foundName := false
				for _, attr := range n.Attr {
					if attr.Key == "name" && attr.Val != name {
						continue
					} else if attr.Key == "name" && attr.Val == name {
						foundName = true
					}
					if attr.Key == "content" {
						content = attr.Val
					}
				}

				if content == "" || !foundName {
					continue
				}

				if m := findMatches(content, detects, n.Data); len(m) > 0 {
					for _, v := range m {
						if _, ok := results[v.AppName]; !ok {
							results[v.AppName] = make([]*Match, 0)
						}
						results[v.AppName] = append(results[v.AppName], v)
					}
				}
			}
		}
		return nil
	}

	parsers.ParseHTML(dom, fn)
	return results
}

// runs a list of regexes on content
func findMatches(content string, regexes []AppRegexp, location string) []*Match {

	matches := make([]*Match, 0)
	for _, r := range regexes {

		match := &Match{}
		regexMatches := r.Regexp.FindAllStringSubmatch(content, -1)
		if regexMatches == nil {
			continue
		}
		match.AppName = r.AppName
		match.Matches = regexMatches
		match.MatchLocation = location

		if r.Version != "" {
			match.Version = findVersion(regexMatches, r.Version)
		}
		matches = append(matches, match)
	}
	return matches
}

// create a worker pool for concurrently attempting regexes against an HTML document, cut amazon.co.jp from 2.7s to 1.04s
// DO NOT USE this for 'small' work loads (headers/cookies) it costs less to just do synchronously.
func findMatchesPool(content string, regexes []AppRegexp, location string) []*Match {
	poolLen := runtime.NumCPU() // cpu bound so only need a pool of # cpus
	pool := workerpool.New(poolLen)
	out := make(chan *Match, 1000)
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	task := func(r AppRegexp) func() {
		return func() {
			match := &Match{}
			regexMatches := r.Regexp.FindAllStringSubmatch(content, -1)
			if regexMatches == nil {
				return
			}
			match.AppName = r.AppName
			match.Matches = regexMatches
			match.MatchLocation = location

			if r.Version != "" {
				match.Version = findVersion(regexMatches, r.Version)
			}
			select {
			case out <- match:
			case <-timeoutCtx.Done():
				log.Error().Msg("failed to send match")
			}
		}
	}

	for _, r := range regexes {
		regex := r
		pool.Submit(task(regex))
	}

	pool.StopWait()
	close(out)

	matches := make([]*Match, 0)
	for match := range out {
		matches = append(matches, match)
	}

	return matches
}

// parses a version against matches
func findVersion(matches [][]string, version string) string {
	var v string

	for _, matchPair := range matches {
		// replace backtraces (max: 3)
		for i := 1; i <= 3; i++ {
			bt := fmt.Sprintf("\\%v", i)
			if strings.Contains(version, bt) && len(matchPair) >= i {
				v = strings.Replace(version, bt, matchPair[i], 1)
			}
		}

		// return first found version
		if v != "" {
			return v
		}
	}
	return ""
}

func compileNamedRegexes(appName string, from map[string]string) []AppRegexp {
	var list []AppRegexp

	for key, value := range from {
		h := AppRegexp{
			AppName: appName,
			Name:    key,
		}

		if value == "" {
			value = ".*"
		}

		// Filter out webapplyzer attributes from regular expression
		splitted := strings.Split(value, "\\;")

		r, err := regexp.Compile(splitted[0])
		if err != nil {
			continue
		}

		if len(splitted) > 1 {
			for _, split := range splitted {
				if strings.HasPrefix(split, "version:") {
					h.Version = split[8:]
					break
				}
			}
		}

		h.Regexp = r
		list = append(list, h)
	}
	return list
}

func compileRegexes(appName string, s StringArray) []AppRegexp {
	var list []AppRegexp

	for _, regexString := range s {
		// Split version detection
		splitted := strings.Split(regexString, "\\;")

		regex, err := regexp.Compile(splitted[0])
		if err != nil {
			//log.Error().Err(err).Msgf("warning: compiling regexp for failed: %v", regexString)
		} else {
			rv := AppRegexp{
				AppName: appName,
				Regexp:  regex,
			}

			if len(splitted) > 1 {
				for _, split := range splitted {
					if strings.HasPrefix(split, "version:") {
						rv.Version = split[8:]
						break
					}
				}
			}

			list = append(list, rv)
		}
	}
	return list
}
