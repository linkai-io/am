package webtech

import (
	"regexp"

	"github.com/linkai-io/am/am"
)

// Match type encapsulates the App information from a match on a document
type Match struct {
	AppName       string     `json:"app_name"`
	Matches       [][]string `json:"matches"`
	Version       string     `json:"version"`
	MatchLocation string     `json:"match_location"`
}

type AppRegexp struct {
	AppName string
	Name    string
	Regexp  *regexp.Regexp
	Version string
}

type Detector interface {
	Init(config []byte) error
	// injects js into the current browser context to detect versions via js objects
	JS(jsObjects []*JSObject) map[string][]*Match
	// runs detector against headers/cookies
	Headers(headers map[string]string) map[string][]*Match
	// runs detector against HTML, meta, script tags etc
	DOM(dom string) map[string][]*Match
	// returns a script of js code to inject for detecting tech
	JSToInject() string
	// JSResultsToObjects takes in an interface, and marshals it to a slice
	// of JSObjects
	JSResultsToObjects(in interface{}) []*JSObject
	// Merges all matches found and returns a map of WebTech results
	MergeMatches(results []map[string][]*Match) map[string]*am.WebTech
}
