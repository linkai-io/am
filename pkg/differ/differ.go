package differ

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/pkg/parsers"
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

func (d *Diff) DiffRemove(text1, text2 string) string {
	dom1, scripts1 := d.domParse(text1)
	dom2, scripts2 := d.domParse(text2)
	fmt.Printf("%s\nDOM2\n%s\n", dom1, dom2)
	fmt.Printf("%s\n%s\n", convert.HashData([]byte(dom1)), convert.HashData([]byte(dom2)))
	fmt.Printf("%#v %#v\n\n\n\n", scripts1, scripts2)
	fmt.Printf("%v\n", d.compareJS(scripts1, scripts2))
	return ""
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
			if prevNode == "script" || prevNode == "style" || prevNode == "noscript" {
				prevNode = ""
				return nil
			}
			data.WriteString(n.Data)
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
