package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/yuriadams/lear/internal/model"
	"golang.org/x/net/html"
)

func ScrapeMetadata(bookID int) (*model.Metadata, error) {
	url := fmt.Sprintf("https://www.gutenberg.org/ebooks/%d", bookID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	metadata := &model.Metadata{}

	var parse func(*html.Node)
	parse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch {
			case n.Data == "h1":
				metadata.Title = strings.TrimSpace(getNodeText(n))
			case n.Data == "a" && hasAttr(n, "href", "/author"):
				metadata.Author = strings.TrimSpace(getNodeText(n))
			case n.Data == "td" && hasAttr(n.Parent, "contains", "Language"):
				metadata.Language = strings.TrimSpace(getNodeText(n))
			case n.Data == "td" && hasAttr(n.Parent, "contains", "Subject"):
				metadata.Subject = strings.TrimSpace(getNodeText(n))
			case n.Data == "td" && hasAttr(n.Parent, "contains", "Produced by"):
				metadata.Credits = strings.TrimSpace(getNodeText(n))
			case n.Data == "meta" && hasAttr(n, "name", "description"):
				metadata.Summary = strings.TrimSpace(getAttrValue(n, "content"))
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parse(c)
		}
	}

	parse(doc)

	return metadata, nil
}

func getNodeText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += getNodeText(c)
	}
	return text
}

func hasAttr(n *html.Node, key, value string) bool {
	for _, attr := range n.Attr {
		if attr.Key == key && strings.Contains(attr.Val, value) {
			return true
		}
	}
	return false
}

func getAttrValue(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
