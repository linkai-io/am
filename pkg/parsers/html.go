package parsers

import (
	"strings"

	"golang.org/x/net/html"
)

type HTMLParserFn func(node *html.Node) error

func ParseHTML(dom string, fn HTMLParserFn) error {
	n, err := html.Parse(strings.NewReader(dom))
	if err != nil {
		return err
	}
	_, err = walkNodes(fn, n)
	return err
}

func walkNodes(fn HTMLParserFn, node *html.Node) (*html.Node, error) {
	err := fn(node)
	if err != nil {
		return nil, err
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		walkNodes(fn, c)
	}
	return nil, nil
}
