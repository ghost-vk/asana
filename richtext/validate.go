// Package richtext converts and validates the rich-text subset Asana accepts
// in the html_notes / html_text fields.
//
// Asana parses those fields as strict XML, not as HTML, and answers every
// violation with the same opaque `xml_parsing_error`. Validating locally turns
// that into a message naming the actual offending tag.
package richtext

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
)

// AllowedTags is the tag set Asana accepts inside html_notes / html_text.
// Notably <p> is not among them: paragraphs are plain newlines in the body.
var AllowedTags = map[string]bool{
	"body":       true,
	"h1":         true,
	"h2":         true,
	"ul":         true,
	"ol":         true,
	"li":         true,
	"strong":     true,
	"em":         true,
	"u":          true,
	"s":          true,
	"code":       true,
	"pre":        true,
	"blockquote": true,
	"a":          true,
	"hr":         true,
}

// voidTagRe catches tags that must be self-closed for the XML parser: <hr>, <br>, <img ...>.
var voidTagRe = regexp.MustCompile(`(?i)<(hr|br|img)((\s[^<>]*[^/<>])?)>`)

// bareAmpRe catches an & that does not start a character reference.
var bareAmpRe = regexp.MustCompile(`&(?:#[0-9]+;|#x[0-9a-fA-F]+;|[a-zA-Z][a-zA-Z0-9]*;)?`)

// Validate reports whether html is accepted by Asana's html_notes parser.
// The error names the concrete problem, which the API itself never does.
func Validate(html string) error {
	trimmed := strings.TrimSpace(html)
	if trimmed == "" {
		return fmt.Errorf("html body is empty")
	}
	if !strings.HasPrefix(trimmed, "<body>") || !strings.HasSuffix(trimmed, "</body>") {
		return fmt.Errorf("html must be wrapped in a single root <body> element")
	}
	if m := voidTagRe.FindStringSubmatch(trimmed); m != nil {
		return fmt.Errorf("<%s> must be self-closing: write <%s/> instead of <%s>", m[1], m[1], m[1])
	}
	if err := checkAmpersands(trimmed); err != nil {
		return err
	}

	unsupported := map[string]bool{}
	dec := xml.NewDecoder(strings.NewReader(trimmed))
	depth := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("html is invalid XML: %v", err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := strings.ToLower(t.Name.Local)
			if depth == 0 && name != "body" {
				return fmt.Errorf("root element must be <body>, got <%s>", name)
			}
			if !AllowedTags[name] {
				unsupported[name] = true
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(t)) != "" {
				return fmt.Errorf("text outside the root <body> element")
			}
		}
	}
	if len(unsupported) > 0 {
		names := make([]string, 0, len(unsupported))
		for n := range unsupported {
			names = append(names, "<"+n+">")
		}
		sort.Strings(names)
		return fmt.Errorf("unsupported tag(s): %s (Asana allows only %s)",
			strings.Join(names, ", "), allowedList())
	}
	return nil
}

func checkAmpersands(html string) error {
	for _, m := range bareAmpRe.FindAllString(html, -1) {
		if m == "&" {
			return fmt.Errorf("bare '&' found; escape it as &amp; (also escape < as &lt; and > as &gt;)")
		}
	}
	return nil
}

func allowedList() string {
	names := make([]string, 0, len(AllowedTags))
	for n := range AllowedTags {
		names = append(names, n)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
