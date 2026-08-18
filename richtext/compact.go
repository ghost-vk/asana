package richtext

import (
	"regexp"
	"strconv"
	"strings"
)

// cosmeticBreakRe matches whitespace containing a newline that sits wholly
// between two tags.
var cosmeticBreakRe = regexp.MustCompile(`>[ \t\r]*\n[ \t\r\n]*<`)

// preInnerRe captures the content of a <pre> block, tags excluded.
var preInnerRe = regexp.MustCompile(`(?s)<pre>(.*?)</pre>`)

// Compact removes the newlines a hand-formatted document carries between block
// tags. Asana renders those as blank lines, so indentation meant for the source
// file shows up in the task. Content inside <pre> is left alone: that is where
// alignment and ASCII diagrams live.
func Compact(html string) string {
	var preserved []string
	html = preInnerRe.ReplaceAllStringFunc(html, func(m string) string {
		preserved = append(preserved, preInnerRe.FindStringSubmatch(m)[1])
		return "<pre>\x00" + strconv.Itoa(len(preserved)-1) + "\x00</pre>"
	})

	html = cosmeticBreakRe.ReplaceAllString(html, "><")

	for i, content := range preserved {
		html = strings.Replace(html, "\x00"+strconv.Itoa(i)+"\x00", content, 1)
	}
	return html
}
