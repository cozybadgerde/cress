package content

import (
	"encoding/xml"
	"fmt"
	"slices"
	"strings"
)

const (
	sitemapPath = "sitemap.xml"
	robotsPath  = "robots.txt"
)

type sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Location string `xml:"loc"`
	Modified string `xml:"lastmod,omitempty"`
}

func preparePublication(pages []Page, config Config) ([]Asset, []string, error) {
	robots := "User-agent: *\nAllow: /\n"
	files := []Asset{{Source: "generated robots.txt", Destination: robotsPath}}
	if config.Host == "" {
		files[0].Data = []byte(robots)
		return files, []string{"cress.toml: site.host is not configured; set an HTTP/HTTPS public origin to generate canonical URLs and sitemap.xml"}, nil
	}
	document := sitemap{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	for _, item := range pages {
		if !item.Draft {
			document.URLs = append(document.URLs, sitemapURL{
				Location: item.Data.Canonical, Modified: item.Data.UpdatedAt,
			})
		}
	}
	slices.SortFunc(document.URLs, func(first, second sitemapURL) int {
		return strings.Compare(first.Location, second.Location)
	})
	data, err := xml.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("encode sitemap: %w", err)
	}
	files[0].Data = []byte(robots + "\nSitemap: " + config.Host + config.Site.URL + sitemapPath + "\n")
	files = append(files, Asset{
		Source: "generated sitemap.xml", Destination: sitemapPath,
		Data: []byte(xml.Header + string(data) + "\n"),
	})
	return files, nil, nil
}
