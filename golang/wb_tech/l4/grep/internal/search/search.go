package search

import (
	"regexp"
	"strconv"
	"strings"
)

type Options struct {
	Pattern    string `json:"pattern"`
	Fixed      bool   `json:"fixed"`
	IgnoreCase bool   `json:"ignoreCase"`
	Invert     bool   `json:"invert"`
	LineNumber bool   `json:"lineNumber"`
}

type Line struct {
	Source string `json:"source"`
	Number int    `json:"number"`
	Text   string `json:"text"`
}

type Matcher interface {
	Match(text string) bool
}

type matcherFunc func(string) bool

func (f matcherFunc) Match(text string) bool {
	return f(text)
}

func CompileMatcher(opts Options) (Matcher, error) {
	if opts.Fixed {
		pattern := opts.Pattern
		if opts.IgnoreCase {
			pattern = strings.ToLower(pattern)
			return matcherFunc(func(text string) bool {
				return strings.Contains(strings.ToLower(text), pattern)
			}), nil
		}

		return matcherFunc(func(text string) bool {
			return strings.Contains(text, pattern)
		}), nil
	}

	pattern := opts.Pattern
	if opts.IgnoreCase {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	return matcherFunc(func(text string) bool {
		return re.MatchString(text)
	}), nil
}

func Filter(lines []Line, opts Options) ([]Line, error) {
	matcher, err := CompileMatcher(opts)
	if err != nil {
		return nil, err
	}

	matches := make([]Line, 0, len(lines))
	for _, line := range lines {
		matched := matcher.Match(line.Text)
		if opts.Invert {
			matched = !matched
		}

		if matched {
			matches = append(matches, line)
		}
	}

	return matches, nil
}

func Format(lines []Line, includeSource bool, includeLineNumber bool) string {
	if len(lines) == 0 {
		return ""
	}

	var b strings.Builder
	for _, line := range lines {
		if includeSource && line.Source != "" {
			b.WriteString(line.Source)
			b.WriteByte(':')
		}

		if includeLineNumber {
			b.WriteString(strconv.Itoa(line.Number))
			b.WriteByte(':')
		}

		b.WriteString(line.Text)
		b.WriteByte('\n')
	}

	return b.String()
}
