package content

import (
	"cmp"
	"path"
	"slices"
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

func (input fields) readList(pageType page.Type) (listOptions, error) {
	options := listOptions{order: "published_at"}
	if _, exists := input.values["list"]; exists && pageType != page.List {
		return options, input.invalidField("list", "is only valid on list pages")
	}
	list, err := input.readTable("list")
	if err != nil {
		return options, err
	}
	list.warnUnknown("order", "reverse")
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

func prepareLists(pages []Page) []string {
	var warnings []string
	for index := range pages {
		list := &pages[index]
		if list.Draft || list.Data.Type != page.List {
			continue
		}
		section := path.Dir(path.Dir(list.Destination))
		var entries []page.Entry
		for _, candidate := range pages {
			if candidate.Draft || candidate.Data.Type != page.Ordinary ||
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
	}
	return warnings
}

func titleKey(title string) string {
	return norm.NFC.String(cases.Fold().String(norm.NFC.String(title)))
}
