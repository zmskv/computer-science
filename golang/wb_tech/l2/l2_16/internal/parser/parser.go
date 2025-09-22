package parser

import (
	"bytes"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/zmskv/computer-science/golang/wb_tech/l2/l2_16/internal/saver"
	"golang.org/x/net/html"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) IsHTML(contentType string) bool {
	return strings.Contains(contentType, "text/html")
}

func (p *Parser) ExtractLinks(body []byte, baseURL *url.URL) []*url.URL {
	var links []*url.URL
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return links
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "a", "link":
				for _, a := range n.Attr {
					if a.Key == "href" {
						link, err := baseURL.Parse(a.Val)
						if err == nil {
							links = append(links, link)
						}
					}
				}
			case "img", "script", "source":
				for _, a := range n.Attr {
					if a.Key == "src" {
						link, err := baseURL.Parse(a.Val)
						if err == nil {
							links = append(links, link)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return links
}

func (p *Parser) RewriteLinks(body []byte, baseURL *url.URL, s *saver.Saver) ([]byte, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	basePath := s.UrlToPath(baseURL)
	baseDir := filepath.Dir(basePath)

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			attrsToRewrite := []string{}
			switch n.Data {
			case "a", "link":
				attrsToRewrite = append(attrsToRewrite, "href")
			case "img", "script", "source":
				attrsToRewrite = append(attrsToRewrite, "src")
			}

			for _, attrName := range attrsToRewrite {
				for i, a := range n.Attr {
					if a.Key == attrName {
						linkURL, err := baseURL.Parse(a.Val)
						if err != nil {
							continue
						}

						if linkURL.Host != baseURL.Host {
							continue
						}

						targetPath := s.UrlToPath(linkURL)

						relPath, err := filepath.Rel(baseDir, targetPath)
						if err != nil {
							continue
						}

						n.Attr[i].Val = relPath
						break
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
