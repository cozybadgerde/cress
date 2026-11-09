package content

import (
	"cmp"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/cozybadgerde/cress/internal/page"
)

const (
	feedName = "feed.xml"
	feedSize = 20
)

// feedEntry is a list member with the rendered body that its page publishes.
type feedEntry struct {
	source string
	entry  page.Entry
	body   template.HTML
}

type atomFeed struct {
	XMLName  xml.Name    `xml:"http://www.w3.org/2005/Atom feed"`
	Language string      `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
	ID       string      `xml:"id"`
	Title    string      `xml:"title"`
	Subtitle string      `xml:"subtitle,omitempty"`
	Updated  string      `xml:"updated"`
	Author   string      `xml:"author>name"`
	Links    []atomLink  `xml:"link"`
	Entries  []atomEntry `xml:"entry"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

type atomEntry struct {
	ID        string      `xml:"id"`
	Title     string      `xml:"title"`
	Link      atomLink    `xml:"link"`
	Published string      `xml:"published"`
	Updated   string      `xml:"updated"`
	Summary   string      `xml:"summary,omitempty"`
	Content   atomContent `xml:"content"`
}

// atomContent sets xml:base so readers resolve root-relative body URLs.
type atomContent struct {
	Type string `xml:"type,attr"`
	Base string `xml:"http://www.w3.org/XML/1998/namespace base,attr"`
	Body string `xml:",chardata"`
}

// selectFeed keeps the newest dated members and warns about undated ones,
// which a feed cannot order.
func selectFeed(list *Page, members []feedEntry, config Config) []string {
	if config.Host == "" {
		return nil
	}
	var warnings []string
	var dated []feedEntry
	for _, member := range members {
		if member.entry.PublishedAt == "" {
			warnings = append(warnings, member.source+": missing published_at; omitted from the feed of "+list.Source)
			continue
		}
		dated = append(dated, member)
	}
	slices.SortFunc(dated, func(first, second feedEntry) int {
		return cmp.Or(strings.Compare(second.entry.PublishedAt, first.entry.PublishedAt),
			strings.Compare(first.entry.URL, second.entry.URL))
	})
	if len(dated) > 0 {
		list.feed = dated[:min(len(dated), feedSize)]
		list.Data.Feed = list.Data.URL + feedName
	}
	return warnings
}

func encodeFeed(list Page, config Config) (Asset, error) {
	author := config.Site.Title
	if name, ok := config.Profile["name"].(string); ok {
		author = name
	}
	if author == "" {
		host, err := url.Parse(config.Host)
		if err != nil {
			return Asset{}, fmt.Errorf("%s: feed author: %w", list.Source, err)
		}
		author = host.Host
	}
	document := atomFeed{
		Language: config.Site.Language, ID: list.Data.Canonical, Title: list.Data.Title,
		Subtitle: list.Data.Description, Author: author,
		Links: []atomLink{
			{Rel: "self", Type: "application/atom+xml", Href: config.Host + list.Data.Feed},
			{Rel: "alternate", Type: "text/html", Href: list.Data.Canonical},
		},
	}
	for _, member := range list.feed {
		entry, link := member.entry, config.Host+member.entry.URL
		updated := feedTime(cmp.Or(entry.UpdatedAt, entry.PublishedAt))
		document.Updated = max(document.Updated, updated)
		document.Entries = append(document.Entries, atomEntry{
			ID: link, Title: entry.Title, Link: atomLink{Rel: "alternate", Type: "text/html", Href: link},
			Published: feedTime(entry.PublishedAt), Updated: updated, Summary: entry.Excerpt,
			Content: atomContent{Type: "html", Base: link, Body: string(member.body)},
		})
	}
	data, err := xml.MarshalIndent(document, "", "  ")
	if err != nil {
		return Asset{}, fmt.Errorf("%s: encode feed: %w", list.Source, err)
	}
	return Asset{
		Source: "generated feed for " + list.Source, Destination: path.Join(path.Dir(list.Destination), feedName),
		Data: []byte(xml.Header + string(data) + "\n"),
	}, nil
}

func feedTime(date string) string {
	return date + "T00:00:00Z"
}
