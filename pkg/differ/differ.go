package differ

import (
	"context"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/parsers"
	"github.com/rs/zerolog/log"
	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/net/html"
)

type Diff struct {
	dmp *diffmatchpatch.DiffMatchPatch
}

func New() *Diff {
	return &Diff{
		dmp: diffmatchpatch.New(),
	}
}

// DiffHash parses text1 and text2 html and
func (d *Diff) DiffHash(ctx context.Context, text1, text2 string) (string, bool) {
	dom1, scripts1 := d.domParse(text1)
	dom2, scripts2 := d.domParse(text2)
	dom1 += parseScript(scripts1)
	dom2 += parseScript(scripts2)
	result1 := d.DiffPatch(dom1, dom2, 20)
	result2 := d.DiffPatch(dom2, dom1, 20)
	hash1 := convert.HashData([]byte(result1))
	hash2 := convert.HashData([]byte(result2))
	same := hash1 == hash2
	log.Info().Str("hash1", hash1).Str("hash2", hash2).Msg("comparing hash v2")
	return hash1, same
}

// DiffPatch takes the diff of two inputs, and only keeps text that is equal and the length of the text is > length
func (d *Diff) DiffPatch(in1, in2 string, length int) string {
	diff := d.dmp.DiffMain(in1, in2, true)
	result := ""
	for _, d := range diff {
		if d.Type == diffmatchpatch.DiffEqual && len(d.Text) > length {
			result += d.Text
		}
	}
	return result
}

// DiffPatchURL takes in two urls and patches (removes) any differences between them
func (d *Diff) DiffPatchURL(in1, in2 string) string {
	if in1 == in2 {
		return in1
	}

	u1, err1 := url.Parse(in1)
	if err1 != nil {
		log.Warn().Err(err1).Str("url", in1).Msg("failed to parse url")
		return in1
	}

	u2, err2 := url.Parse(in2)
	if err2 != nil {
		log.Warn().Err(err2).Str("url", in2).Msg("failed to parse url")
		return in1
	}

	if u1.Host != u2.Host {
		log.Warn().Str("host1", u1.Host).Str("host2", u2.Host).Msg("hosts of url did not match")
		return u2.String()
	}

	log.Info().Msgf("%s and %s", u1.Path, u2.Path)
	if u1.Path != u2.Path {
		log.Warn().Str("path1", u1.Path).Str("path2", u2.Path).Msg("paths of url did not match")
		return d.DiffPatch(in1, in2, 1)
	}

	// iterate over query value (keys) and compare in1 with in2 url query parameters.
	// if they don't match, set it to an empty string and re-encode the query back to the rawquery
	for k := range u1.Query() {
		q := u1.Query()
		// Note this does not handle if params have multiple values but that should be rare
		if u1.Query().Get(k) != u2.Query().Get(k) {
			q.Set(k, "")
			u1.RawQuery = q.Encode()
			log.Info().Str("param", q.Get(k)).Str("param", u2.Query().Get(k)).Str("raw_query", u1.RawQuery).Msg("param of url did not match")
		}
	}
	return u1.String()
}

func parseScript(scripts []string) string {
	toSort := make([]string, len(scripts))
	for i := range scripts {
		url, err := url.Parse(scripts[i])
		if err != nil {
			log.Info().Msgf("%s not a url\n", scripts[i])
			continue
		}
		_, fileName := filepath.Split(url.Path)
		toSort[i] = url.Host + fileName
	}

	sort.Strings(toSort)
	return strings.Join(toSort, "")
}

func (d *Diff) compareJS(scripts1, scripts2 []string) bool {
	if len(scripts1) != len(scripts2) {
		return false
	}
	for i := 0; i < len(scripts1); i++ {
		if scripts1[i] == scripts2[i] {
			continue
		}
		u, err := url.Parse(scripts1[i])
		if err != nil {
			log.Info().Msgf("%s not a url\n", scripts1[i])
		}

		u2, err2 := url.Parse(scripts2[i])
		if err2 != nil {
			log.Info().Msgf("%s not a url\n", scripts2[i])
		}
		log.Info().Msgf("host: %s %s\n", u.Host, u2.Host)
		if u.Host != u2.Host {
			log.Info().Msgf("NOT EQUAL HOST\n")
		}
		_, file1 := filepath.Split(u.Path)
		_, file2 := filepath.Split(u2.Path)
		log.Info().Msgf("uri: [%s] [%s]\n", file1, file2)
		if file1 != file2 {
			log.Info().Msgf("NOT EQUAL PATH\n")
		}
	}
	return true
}

func (d *Diff) domParse(text string) (string, []string) {
	prevNode := ""
	data := &strings.Builder{}
	scripts := make([]string, 0)

	fn := func(n *html.Node) error {
		if n.Type == html.TextNode {
			//prevNode == "script" ||
			if prevNode == "style" || prevNode == "noscript" {
				prevNode = ""
				return nil
			}
			data.WriteString(strings.Trim(n.Data, " \t"))
		}
		if n.Type != html.ElementNode {
			return nil
		}
		prevNode = n.Data
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
			scripts = append(scripts, src)
		}

		return nil
	}
	parsers.ParseHTML(text, fn)

	return data.String(), scripts
}
