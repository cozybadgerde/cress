package content

import (
	"cmp"
	"math"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"

	"github.com/cozybadgerde/cress/internal/page"
)

type listOptions struct {
	order   string
	reverse bool
	perPage int
}

func (input fields) readDate(key string) (string, error) {
	value, exists := input.values[key]
	if !exists {
		return "", nil
	}
	var date string
	switch value := value.(type) {
	case string:
		date = value
	case toml.LocalDate:
		date = value.String()
	default:
		return "", input.invalidField(key, "must be a YYYY-MM-DD string or TOML local date")
	}
	parsed, err := time.Parse(time.DateOnly, date)
	if err != nil || parsed.Year() < 1 || parsed.Format(time.DateOnly) != date {
		return "", input.invalidField(key, "must be a valid YYYY-MM-DD calendar date")
	}
	return date, nil
}

func (input fields) readPerPage(fallback int) (int, error) {
	value, exists := input.values["per_page"]
	if !exists {
		return fallback, nil
	}
	size, ok := value.(int64)
	if !ok || size < 0 || size > math.MaxInt32 {
		return 0, input.invalidField("per_page", "must be a nonnegative integer; 0 keeps lists on one page")
	}
	return int(size), nil
}

func (input fields) readList(pageType page.Type, perPage int) (listOptions, error) {
	options := listOptions{order: "published_at", perPage: perPage}
	if _, exists := input.values["list"]; exists && pageType != page.List {
		return options, input.invalidField("list", "is only valid on list pages")
	}
	list, err := input.readTable("list")
	if err != nil {
		return options, err
	}
	list.warnUnknown("order", "reverse", "per_page")
	options.perPage, err = list.readPerPage(options.perPage)
	if err != nil {
		return options, err
	}
	options.order, err = list.readString("order", options.order)
	if err != nil {
		return options, err
	}
	if options.order != "published_at" && options.order != "title" {
		return options, list.invalidField("order", "must be published_at or title")
	}
	if value, exists := list.values["reverse"]; exists {
		reverse, ok := value.(bool)
		if !ok {
			return options, list.invalidField("reverse", "must be a boolean")
		}
		options.reverse = reverse
	}
	return options, nil
}

func prepareLists(pages []Page, config Config) ([]Page, []string) {
	var warnings []string
	var numbered []Page
	for index := range pages {
		list := &pages[index]
		if (list.Draft && !config.IncludeDrafts) || list.Data.Type != page.List {
			continue
		}
		section := path.Dir(path.Dir(list.Destination))
		var entries []page.Entry
		for _, candidate := range pages {
			if (candidate.Draft && !config.IncludeDrafts) || candidate.Data.Type != page.Ordinary ||
				candidate.Destination == list.Destination ||
				path.Dir(candidate.Destination) == section ||
				path.Dir(path.Dir(candidate.Destination)) != section {
				continue
			}
			if list.list.order == "published_at" && candidate.Data.PublishedAt == "" {
				warnings = append(warnings, candidate.Source+": missing published_at; omitted from chronological list "+list.Source)
				continue
			}
			data := candidate.Data
			entries = append(entries, page.Entry{
				Title: data.Title, URL: data.URL, Description: data.Description, Image: data.Image,
				PublishedAt: data.PublishedAt, UpdatedAt: data.UpdatedAt, Excerpt: data.Excerpt,
			})
		}
		slices.SortFunc(entries, func(first, second page.Entry) int {
			var order int
			if list.list.order == "title" {
				order = strings.Compare(titleKey(first.Title), titleKey(second.Title))
			} else {
				order = strings.Compare(second.PublishedAt, first.PublishedAt)
			}
			return cmp.Or(order, strings.Compare(first.URL, second.URL))
		})
		if list.list.reverse {
			slices.Reverse(entries)
		}
		list.Data.Entries = entries
		numbered = append(numbered, paginate(list, config)...)
	}
	return append(pages, numbered...), warnings
}

// paginate keeps the first page of entries on the list and returns the
// numbered pages that follow it.
func paginate(list *Page, config Config) []Page {
	size, entries := list.list.perPage, list.Data.Entries
	if size == 0 || len(entries) <= size {
		return nil
	}
	links := make([]page.PageLink, (len(entries)+size-1)/size)
	for index := range links {
		links[index] = page.PageLink{Number: index + 1, URL: list.Data.URL}
		if index > 0 {
			links[index].URL += strconv.Itoa(index+1) + "/"
		}
	}
	var numbered []Page
	for index, link := range links {
		item := *list
		if index > 0 {
			item.Destination = path.Join(path.Dir(list.Destination), strconv.Itoa(link.Number), "index.html")
			item.Data.URL = link.URL
			if config.Host != "" {
				item.Data.Canonical = config.Host + link.URL
			}
		}
		item.Data.Entries = entries[index*size : min((index+1)*size, len(entries))]
		pagination := &page.Pagination{Number: link.Number, Total: len(links), Pages: slices.Clone(links)}
		pagination.Pages[index].Current = true
		if index > 0 {
			pagination.Previous = links[index-1].URL
		}
		if index < len(links)-1 {
			pagination.Next = links[index+1].URL
		}
		item.Data.Pagination = pagination
		if index == 0 {
			list.Data = item.Data
		} else {
			numbered = append(numbered, item)
		}
	}
	return numbered
}

func titleKey(title string) string {
	return norm.NFC.String(cases.Fold().String(norm.NFC.String(title)))
}
