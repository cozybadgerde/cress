package content

import (
	"fmt"
	"net/mail"
	"regexp"
	"slices"
	"strings"
	"time"
)

var (
	profileType = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*$`)
	profileDays = map[string]string{
		"mo": "Monday", "tu": "Tuesday", "we": "Wednesday", "th": "Thursday",
		"fr": "Friday", "sa": "Saturday", "su": "Sunday",
	}
)

// readProfile validates [profile] and returns schema.org JSON-LD data, or nil
// when absent. Origin-dependent URLs are added by the caller.
func (input fields) readProfile(siteTitle string) (map[string]any, error) {
	if _, exists := input.values["profile"]; !exists {
		return nil, nil
	}
	profile, err := input.readTable("profile")
	if err != nil {
		return nil, err
	}
	profile.warnUnknown("type", "name", "telephone", "email", "address", "hours")
	kind, err := profile.readNonblank("type", "Organization")
	if err != nil {
		return nil, err
	}
	if !profileType.MatchString(kind) {
		return nil, profile.invalidField("type", "must be a schema.org type name, such as LocalBusiness")
	}
	name, err := profile.readNonblank("name", siteTitle)
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, profile.invalidField("name", "is required when site.title is not set")
	}
	data := map[string]any{"@context": "https://schema.org", "@type": kind, "name": name}
	if err := profile.copyStrings(data, map[string]string{"telephone": "telephone", "email": "email"}); err != nil {
		return nil, err
	}
	if email, ok := data["email"].(string); ok {
		if parsed, err := mail.ParseAddress(email); err != nil || parsed.Address != email {
			return nil, profile.invalidField("email", "must be a plain email address, such as hello@example.org")
		}
	}
	address, err := profile.readTable("address")
	if err != nil {
		return nil, err
	}
	address.warnUnknown("street", "postal_code", "city", "region", "country")
	postal := map[string]any{}
	if err := address.copyStrings(postal, map[string]string{
		"street": "streetAddress", "postal_code": "postalCode", "city": "addressLocality",
		"region": "addressRegion", "country": "addressCountry",
	}); err != nil {
		return nil, err
	}
	if len(postal) > 0 {
		postal["@type"] = "PostalAddress"
		data["address"] = postal
	}
	hours, err := profile.readHours()
	if err != nil {
		return nil, err
	}
	if len(hours) > 0 {
		data["openingHoursSpecification"] = hours
	}
	return data, nil
}

// copyStrings copies nonblank string fields to their JSON-LD property names.
func (input fields) copyStrings(target map[string]any, properties map[string]string) error {
	for key, property := range properties {
		value, err := input.readString(key, "")
		if err != nil {
			return err
		}
		if strings.TrimSpace(value) != "" {
			target[property] = value
		}
	}
	return nil
}

func (input fields) readHours() ([]any, error) {
	value, exists := input.values["hours"]
	if !exists {
		return nil, nil
	}
	items, ok := value.([]any)
	if !ok {
		return nil, input.invalidField("hours", "must be an array of tables")
	}
	var hours []any
	for index, item := range items {
		values, ok := item.(map[string]any)
		if !ok {
			return nil, input.invalidField(fmt.Sprintf("hours[%d]", index), "must be a table")
		}
		entry := fields{
			values: values, source: input.source, kind: input.kind,
			prefix: fmt.Sprintf("%shours[%d].", input.prefix, index), warnings: input.warnings,
		}
		entry.warnUnknown("days", "opens", "closes")
		days, ok := values["days"].([]any)
		if !ok || len(days) == 0 {
			return nil, entry.invalidField("days", "must be a nonempty array of mo, tu, we, th, fr, sa, or su")
		}
		var names []string
		for _, day := range days {
			key, _ := day.(string)
			name, known := profileDays[key]
			if !known || slices.Contains(names, name) {
				return nil, entry.invalidField("days", "must list distinct days from mo, tu, we, th, fr, sa, and su")
			}
			names = append(names, name)
		}
		specification := map[string]any{"@type": "OpeningHoursSpecification", "dayOfWeek": names}
		for _, key := range []string{"opens", "closes"} {
			text, err := entry.readString(key, "")
			if err != nil {
				return nil, err
			}
			if parsed, err := time.Parse("15:04", text); err != nil || parsed.Format("15:04") != text {
				return nil, entry.invalidField(key, "must be a 24-hour HH:MM time")
			}
			specification[key] = text
		}
		hours = append(hours, specification)
	}
	return hours, nil
}
