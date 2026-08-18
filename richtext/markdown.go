package richtext

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	headingRe   = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	trailingHRe = regexp.MustCompile(`\s+#+\s*$`)
	hrRe        = regexp.MustCompile(`^(-{3,}|\*{3,}|_{3,})$`)
	listRe      = regexp.MustCompile(`^(\s*)(?:([-*+])|(\d+)[.)])\s+(.*)$`)
	quoteRe     = regexp.MustCompile(`^>\s?(.*)$`)
	fenceRe     = regexp.MustCompile("^(```|~~~)")
	tableSepRe  = regexp.MustCompile(`^[\s|:-]*-[\s|:-]*$`)

	codeSpanRe  = regexp.MustCompile("`([^`]*)`")
	imageRe     = regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)\)`)
	linkRe      = regexp.MustCompile(`\[([^\]]*)\]\(([^)\s]+)\)`)
	strongRe    = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	strongAltRe = regexp.MustCompile(`__([^_]+)__`)
	strikeRe    = regexp.MustCompile(`~~([^~]+)~~`)
	emRe        = regexp.MustCompile(`\*([^*]+)\*`)
	emAltRe     = regexp.MustCompile(`_([^_]+)_`)
)

// FromMarkdown converts markdown into the tag subset Asana accepts in
// html_notes. Constructs Asana cannot render are degraded rather than dropped:
// h3 and deeper collapse to h2, images become links, tables become text rows.
// The result always passes Validate.
func FromMarkdown(md string) string {
	lines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(md, "\r\n", "\n"), "\r", "\n"), "\n")
	var o writer
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		switch {
		case trimmed == "":
			continue

		case fenceRe.MatchString(trimmed):
			var body []string
			marker := fenceRe.FindString(trimmed)
			i++
			for ; i < len(lines); i++ {
				if strings.HasPrefix(strings.TrimSpace(lines[i]), marker) {
					break
				}
				body = append(body, lines[i])
			}
			o.block("<pre>" + escape(strings.Join(trimTrailingBlank(body), "\n")) + "</pre>")

		case headingRe.MatchString(trimmed):
			m := headingRe.FindStringSubmatch(trimmed)
			level := len(m[1])
			if level > 2 {
				level = 2 // Asana only renders h1 and h2
			}
			text := trailingHRe.ReplaceAllString(m[2], "")
			o.block(fmt.Sprintf("<h%d>%s</h%d>", level, inline(text), level))

		case hrRe.MatchString(trimmed):
			o.block("<hr/>")

		case quoteRe.MatchString(trimmed):
			var body []string
			for ; i < len(lines); i++ {
				m := quoteRe.FindStringSubmatch(strings.TrimSpace(lines[i]))
				if m == nil {
					break
				}
				body = append(body, inline(m[1]))
			}
			i--
			o.block("<blockquote>" + strings.Join(body, "\n") + "</blockquote>")

		case listRe.MatchString(line):
			var items []listItem
			for ; i < len(lines); i++ {
				m := listRe.FindStringSubmatch(lines[i])
				if m == nil {
					break
				}
				items = append(items, listItem{
					indent:  len(strings.ReplaceAll(m[1], "\t", "    ")),
					ordered: m[2] == "",
					text:    m[4],
				})
			}
			i--
			html, _ := renderList(items, 0)
			o.block(html)

		case strings.HasPrefix(trimmed, "|"):
			var rows []string
			for ; i < len(lines); i++ {
				cur := strings.TrimSpace(lines[i])
				if !strings.HasPrefix(cur, "|") {
					break
				}
				if tableSepRe.MatchString(cur) {
					continue
				}
				rows = append(rows, inline(flattenRow(cur)))
			}
			i--
			o.para(strings.Join(rows, "\n"))

		default:
			var body []string
			for ; i < len(lines); i++ {
				cur := lines[i]
				if strings.TrimSpace(cur) == "" || isBlockStart(cur) {
					break
				}
				body = append(body, inline(cur))
			}
			i--
			o.para(strings.Join(body, "\n"))
		}
	}
	return "<body>" + o.b.String() + "</body>"
}

// writer keeps track of whether the previous block was loose text, so that
// consecutive paragraphs get the blank line Asana renders from a bare "\n\n"
// while block tags stay glued together (a stray newline there shows up as an
// empty line in the task).
type writer struct {
	b        strings.Builder
	lastPara bool
}

func (o *writer) para(s string) {
	if s == "" {
		return
	}
	if o.lastPara {
		o.b.WriteString("\n\n")
	}
	o.b.WriteString(s)
	o.lastPara = true
}

func (o *writer) block(s string) {
	o.b.WriteString(s)
	o.lastPara = false
}

type listItem struct {
	indent  int
	ordered bool
	text    string
}

// renderList emits one list starting at items[pos] and returns the index of the
// first item that does not belong to it. Deeper-indented items recurse into a
// nested list inside the preceding <li>.
func renderList(items []listItem, pos int) (string, int) {
	if pos >= len(items) {
		return "", pos
	}
	base, ordered := items[pos].indent, items[pos].ordered
	tag := "ul"
	if ordered {
		tag = "ol"
	}
	var b strings.Builder
	b.WriteString("<" + tag + ">")
	for pos < len(items) {
		it := items[pos]
		if it.indent < base || (it.indent == base && it.ordered != ordered) {
			break
		}
		b.WriteString("<li>" + inline(it.text))
		pos++
		for pos < len(items) && items[pos].indent > base {
			nested, next := renderList(items, pos)
			b.WriteString(nested)
			pos = next
		}
		b.WriteString("</li>")
	}
	b.WriteString("</" + tag + ">")
	return b.String(), pos
}

func isBlockStart(line string) bool {
	trimmed := strings.TrimSpace(line)
	return fenceRe.MatchString(trimmed) ||
		headingRe.MatchString(trimmed) ||
		hrRe.MatchString(trimmed) ||
		quoteRe.MatchString(trimmed) ||
		listRe.MatchString(line) ||
		strings.HasPrefix(trimmed, "|")
}

func flattenRow(row string) string {
	cells := strings.Split(strings.Trim(row, "|"), "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	return strings.Join(cells, " | ")
}

func trimTrailingBlank(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// inline renders span-level markdown. Code spans are pulled out first so their
// emphasis markers stay literal, then the rest is escaped once and rewritten.
func inline(text string) string {
	var codes []string
	text = codeSpanRe.ReplaceAllStringFunc(text, func(m string) string {
		codes = append(codes, codeSpanRe.FindStringSubmatch(m)[1])
		return "\x00" + strconv.Itoa(len(codes)-1) + "\x00"
	})

	text = spans(escape(text))

	for i, code := range codes {
		text = strings.ReplaceAll(text, "\x00"+strconv.Itoa(i)+"\x00", "<code>"+escape(code)+"</code>")
	}
	return text
}

// spanPattern is one span-level construct. Patterns are tried in order, and at
// an equal starting offset the earlier one wins, so ** beats * and __ beats _.
type spanPattern struct {
	re   *regexp.Regexp
	wrap func(groups []string, inner string) string
}

var spanPatterns = []spanPattern{
	{imageRe, func(g []string, inner string) string { return anchor(g[2], inner) }},
	{linkRe, func(g []string, inner string) string { return anchor(g[2], inner) }},
	{strongRe, func(_ []string, inner string) string { return "<strong>" + inner + "</strong>" }},
	{strongAltRe, func(_ []string, inner string) string { return "<strong>" + inner + "</strong>" }},
	{strikeRe, func(_ []string, inner string) string { return "<s>" + inner + "</s>" }},
	{emRe, func(_ []string, inner string) string { return "<em>" + inner + "</em>" }},
	{emAltRe, func(_ []string, inner string) string { return "<em>" + inner + "</em>" }},
}

// spans rewrites the leftmost construct, then recurses separately into its
// content and into the remainder. Applying each pattern over the whole string
// instead would let two constructs overlap - "**a ~~b** c~~" - and emit crossed
// tags, which Asana rejects as invalid XML.
func spans(text string) string {
	best := -1
	var bestLoc []int
	for i, p := range spanPatterns {
		loc := p.re.FindStringSubmatchIndex(text)
		if loc == nil {
			continue
		}
		if best == -1 || loc[0] < bestLoc[0] {
			best, bestLoc = i, loc
		}
	}
	if best == -1 {
		return text
	}
	p := spanPatterns[best]
	groups := make([]string, len(bestLoc)/2)
	for i := range groups {
		if bestLoc[2*i] >= 0 {
			groups[i] = text[bestLoc[2*i]:bestLoc[2*i+1]]
		}
	}
	return text[:bestLoc[0]] + p.wrap(groups, spans(groups[1])) + spans(text[bestLoc[1]:])
}

func anchor(href, label string) string {
	return `<a href="` + strings.ReplaceAll(href, `"`, "&quot;") + `">` + label + `</a>`
}

func escape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
