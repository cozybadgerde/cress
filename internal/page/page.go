package page

import "html/template"

// Page holds prepared ordinary-page data, without source or generator access.
type Page struct {
	Title string
	URL   string
	Body  template.HTML
}
