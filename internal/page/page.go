package page

import "html/template"

// Type identifies an explicitly resolved page behavior.
type Type string

const (
	// Ordinary is an ordinary content page.
	Ordinary Type = "ordinary"
	// Landing is a landing page with distinct template selection.
	Landing Type = "landing"
	// List is a page with a prepared collection.
	List Type = "list"
)

// Image holds a resolved image URL and its plain-text alternative.
type Image struct {
	URL     string
	Alt     string
	Caption []CaptionPart
}

// CaptionPart is escaped caption text, optionally linked to a prepared URL.
type CaptionPart struct {
	Text string
	URL  string
}

// Link holds a resolved navigation destination and plain-text label.
type Link struct {
	Label string
	URL   string
}

// Navigation holds ordered, prepared main and footer links.
type Navigation struct {
	Main   []Link
	Footer []Link
}

// Site contains prepared site presentation data, not raw configuration.
type Site struct {
	Title      string
	Copyright  string
	Language   string
	URL        string
	Navigation Navigation
	Custom     map[string]any
}

// Page holds prepared page data, without source or generator access.
type Page struct {
	Title       string
	Description string
	Type        Type
	URL         string
	Image       *Image
	Body        template.HTML
	Site        Site
	Canonical   string
	PublishedAt string
	UpdatedAt   string
	Excerpt     string
	Entries     []Entry
}

// Entry holds prepared list metadata without an article body or nested entries.
type Entry struct {
	Title       string
	URL         string
	Description string
	Image       *Image
	PublishedAt string
	UpdatedAt   string
	Excerpt     string
}
