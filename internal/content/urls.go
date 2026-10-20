package content

import (
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

func parseURL(value string) (*url.URL, error) {
	if strings.Contains(value, `\`) || strings.IndexFunc(value, func(character rune) bool {
		return unicode.IsSpace(character) || unicode.IsControl(character)
	}) >= 0 {
		return nil, fmt.Errorf("must be a URL without whitespace, control characters, or backslashes")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.User != nil {
		return nil, fmt.Errorf("URL credentials are not allowed")
	}
	if parsed.Scheme != "" {
		if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
			return nil, fmt.Errorf("URL scheme must be HTTP or HTTPS")
		}
		if parsed.Opaque != "" || parsed.Hostname() == "" {
			return nil, fmt.Errorf("absolute URL must have a hostname")
		}
		if strings.ContainsAny(parsed.Host, "[]") {
			address, err := netip.ParseAddr(parsed.Hostname())
			if err != nil {
				return nil, fmt.Errorf("invalid IP address in URL: %w", err)
			}
			if !address.Is6() {
				return nil, fmt.Errorf("bracketed URL address must be IPv6")
			}
		} else if strings.Contains(parsed.Hostname(), ":") {
			return nil, fmt.Errorf("IPv6 addresses must use brackets")
		}
		if port := parsed.Port(); port != "" {
			number, err := strconv.Atoi(port)
			if err != nil || number < 1 || number > 65535 {
				return nil, fmt.Errorf("URL port must be between 1 and 65535")
			}
		} else if strings.HasSuffix(parsed.Host, ":") {
			return nil, fmt.Errorf("URL port must not be empty")
		}
	}
	return parsed, nil
}

// ResolveURL prefixes site-root URLs and leaves absolute HTTP/HTTPS URLs unchanged.
func ResolveURL(value, prefix string) (string, error) {
	parsed, err := parseURL(value)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() {
		return value, nil
	}
	if !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") ||
		parsed.Host != "" || strings.HasPrefix(parsed.Path, "//") {
		return "", fmt.Errorf("must be a site-root path or absolute HTTP/HTTPS URL")
	}
	for _, segment := range strings.Split(parsed.Path, "/") {
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("URL path must not contain dot segments")
		}
	}
	if strings.ContainsAny(parsed.Path, "\\\x00") {
		return "", fmt.Errorf("URL path must not contain backslashes or null bytes")
	}
	return strings.TrimSuffix(prefix, "/") + value, nil
}

func validatePath(value string) (string, error) {
	if !strings.HasPrefix(value, "/") || !strings.HasSuffix(value, "/") ||
		strings.Contains(value, "//") || strings.ContainsAny(value, "%\\?#") {
		return "", fmt.Errorf("must start and end with /, without repeated slashes, escapes, queries, fragments, or backslashes")
	}
	if _, err := parseURL(value); err != nil {
		return "", err
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("must not contain dot segments")
		}
	}
	return (&url.URL{Path: value}).EscapedPath(), nil
}

func validateOrigin(value string) (string, error) {
	parsed, err := parseURL(value)
	if err != nil {
		return "", err
	}
	if !parsed.IsAbs() || parsed.Host == "" || (parsed.Path != "" && parsed.Path != "/") ||
		parsed.RawPath != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" ||
		strings.Contains(value, "#") {
		return "", fmt.Errorf("must be an HTTP/HTTPS origin without a path, query, or fragment")
	}
	return strings.TrimSuffix(value, "/"), nil
}
