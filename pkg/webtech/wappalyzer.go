package webtech

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// WappalyzerURL is the link to the latest apps.json file in the Wappalyzer repo
const WappalyzerURL = "https://raw.githubusercontent.com/AliasIO/Wappalyzer/master/src/apps.json"

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

// Match type encapsulates the App information from a match on a document
type Match struct {
	App     `json:"app"`
	AppName string     `json:"app_name"`
	Matches [][]string `json:"matches"`
	Version string     `json:"version"`
}

// App type encapsulates all the data about an App from apps.json
type App struct {
	Cats     StringArray       `json:"cats"`
	CatNames []string          `json:"category_names"`
	Cookies  map[string]string `json:"cookies"`
	Headers  map[string]string `json:"headers"`
	Meta     map[string]string `json:"meta"`
	HTML     StringArray       `json:"html"`
	Script   StringArray       `json:"script"`
	URL      StringArray       `json:"url"`
	Website  string            `json:"website"`
	Implies  StringArray       `json:"implies"`

	HTMLRegex   []AppRegexp `json:"-"`
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

type AppRegexp struct {
	Name    string
	Regexp  *regexp.Regexp
	Version string
}

// Wappalyzer for finding technology in http responses/data
type Wappalyzer struct {
	url         string
	definitions *AppsDefinition
}

// NewWappalyzer for finding technology in http responses/data
func NewWappalyzer(url string) *Wappalyzer {
	return &Wappalyzer{url: url}
}

// Init downloads and initializes the application definitions
func (w *Wappalyzer) Init() error {
	return w.load()
}

func (w *Wappalyzer) load() error {
	resp, err := http.Get(w.url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&w.definitions); err != nil {
		return err
	}

	for techName, value := range w.definitions.Apps {
		log.Info().Msgf("techName: %s", techName)
		app := w.definitions.Apps[techName]
		app.HTMLRegex = compileRegexes(value.HTML)
		app.ScriptRegex = compileRegexes(value.Script)
		app.URLRegex = compileRegexes(value.URL)
		app.HeaderRegex = compileNamedRegexes(app.Headers)
		app.MetaRegex = compileNamedRegexes(app.Meta)
		app.CookieRegex = compileNamedRegexes(app.Cookies)

		app.CatNames = make([]string, 0)
		if len(app.Cats) > 1 {
			log.Info().Msgf("multi cat: %#v", app.Cats)
		}

		for _, cid := range app.Cats {
			if category, ok := w.definitions.Cats[string(cid)]; ok && category.Name != "" {
				app.CatNames = append(app.CatNames, category.Name)
			}
		}
		log.Info().Msgf("%#v %#v", app.Cats, app.CatNames)
		w.definitions.Apps[techName] = app
	}

	return nil
}

func compileNamedRegexes(from map[string]string) []AppRegexp {
	var list []AppRegexp

	for key, value := range from {
		h := AppRegexp{
			Name: key,
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

		if len(splitted) > 1 && strings.HasPrefix(splitted[1], "version:") {
			h.Version = splitted[1][8:]
		}

		h.Regexp = r
		list = append(list, h)
	}
	return list
}

func compileRegexes(s StringArray) []AppRegexp {
	var list []AppRegexp

	for _, regexString := range s {
		// Split version detection
		splitted := strings.Split(regexString, "\\;")

		regex, err := regexp.Compile(splitted[0])
		if err != nil {
			log.Error().Err(err).Msgf("warning: compiling regexp for failed: %v", regexString)
		} else {
			rv := AppRegexp{
				Regexp: regex,
			}

			if len(splitted) > 1 && strings.HasPrefix(splitted[0], "version") {
				rv.Version = splitted[1][8:]
			}

			list = append(list, rv)
		}
	}
	return list
}
