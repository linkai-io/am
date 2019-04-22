package differ

import (
	"context"
	"fmt"
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
	result1 := d.diffPatch(dom1, dom2)
	result2 := d.diffPatch(dom2, dom1)
	hash1 := convert.HashData([]byte(result1))
	hash2 := convert.HashData([]byte(result2))
	same := hash1 == hash2
	log.Info().Str("hash1", hash1).Str("hash2", hash2).Msg("comparing hash v2")
	return hash1, same
}

func (d *Diff) diffPatch(dom1, dom2 string) string {
	diff := d.dmp.DiffMain(dom1, dom2, true)
	result := ""
	for _, d := range diff {
		if d.Type == diffmatchpatch.DiffEqual && len(d.Text) > 20 {
			result += d.Text
		}
	}
	return result
}

/*
log.Info().Str("hash1", hash1).Str("hash2", hash2).Msg("comparing hash")
	same := hash1 == hash2
	if !same {
		log.Warn().Msg("hashes did not match, doing diffpatch")
		d.compareJS(scripts1, scripts2)
		log.Info().Msg("diffpatch 1")
		result1 := d.diffPatch(dom1, dom2)
		//fmt.Printf("RESULT1: %s\n", result1)
		log.Info().Msg("diffpatch 2")
		result2 := d.diffPatch(dom2, dom1)
		//fmt.Printf("RESULT2: %s\n", result2)
		hash1 = convert.HashData([]byte(result1))
		hash2 = convert.HashData([]byte(result2))
		same = hash1 == hash2
		log.Info().Str("hash1", hash1).Str("hash2", hash2).Msg("comparing hash v2")
	}
*/
func parseScript(scripts []string) string {
	toSort := make([]string, len(scripts))
	for i := range scripts {
		url, err := url.Parse(scripts[i])
		if err != nil {
			fmt.Printf("%s not a url\n", scripts[i])
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
			fmt.Printf("%s not a url\n", scripts1[i])
		}

		u2, err2 := url.Parse(scripts2[i])
		if err2 != nil {
			fmt.Printf("%s not a url\n", scripts2[i])
		}
		fmt.Printf("host: %s %s\n", u.Host, u2.Host)
		if u.Host != u2.Host {
			fmt.Printf("NOT EQUAL HOST\n")
		}
		_, file1 := filepath.Split(u.Path)
		_, file2 := filepath.Split(u2.Path)
		fmt.Printf("uri: [%s] [%s]\n", file1, file2)
		if file1 != file2 {
			fmt.Printf("NOT EQUAL PATH\n")
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
