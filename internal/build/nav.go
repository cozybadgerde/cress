package build

import (
	"fmt"
	"path/filepath"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/theme"
)

// resolveNav resolves every navigation group against the collected pages.
func resolveNav(nav config.Nav, pages []*content.Page, basePath string) (theme.NavView, []string) {
	bySource := make(map[string]*content.Page, len(pages))
	for _, p := range pages {
		bySource[p.SourcePath] = p
	}
	main, mainWarnings := resolveNavGroup("nav.main", nav.Main, bySource, basePath)
	footer, footerWarnings := resolveNavGroup("nav.footer", nav.Footer, bySource, basePath)
	return theme.NavView{Main: main, Footer: footer}, append(mainWarnings, footerWarnings...)
}

// resolveNavGroup maps one group's entries to their page URLs. Entries pointing
// at an unknown file are dropped and reported as warnings naming the group, so
// the rest of the menu still renders.
func resolveNavGroup(group string, items []config.NavItem, bySource map[string]*content.Page, basePath string) ([]theme.NavLink, []string) {
	var (
		links    []theme.NavLink
		warnings []string
	)
	for _, item := range items {
		page, ok := bySource[filepath.ToSlash(item.Path)]
		if !ok {
			warnings = append(warnings, fmt.Sprintf("%s entry %q points at missing content %q", group, item.Title, item.Path))
			continue
		}
		title := item.Title
		if title == "" {
			title = page.Title
		}
		links = append(links, theme.NavLink{Title: title, URL: prefixURL(basePath, page.URL)})
	}
	return links, warnings
}

// activeNav returns a copy of nav with the entries matching currentURL flagged,
// in every group: a footer link to the current page is current too.
func activeNav(nav theme.NavView, currentURL string) theme.NavView {
	return theme.NavView{
		Main:   activeLinks(nav.Main, currentURL),
		Footer: activeLinks(nav.Footer, currentURL),
	}
}

// activeLinks returns a copy of links with the entry matching currentURL flagged.
func activeLinks(links []theme.NavLink, currentURL string) []theme.NavLink {
	out := make([]theme.NavLink, len(links))
	copy(out, links)
	for i := range out {
		out[i].Active = out[i].URL == currentURL
	}
	return out
}
