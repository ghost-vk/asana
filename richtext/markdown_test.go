package richtext

import "testing"

func TestFromMarkdown(t *testing.T) {
	cases := []struct {
		name string
		md   string
		want string
	}{
		{
			name: "empty input",
			md:   "",
			want: "<body></body>",
		},
		{
			name: "plain paragraph",
			md:   "hello world",
			want: "<body>hello world</body>",
		},
		{
			name: "two paragraphs keep a blank line",
			md:   "first\n\nsecond",
			want: "<body>first\n\nsecond</body>",
		},
		{
			name: "soft line break inside a paragraph",
			md:   "first\nsecond",
			want: "<body>first\nsecond</body>",
		},
		{
			name: "headings, h3 and deeper collapse to h2",
			md:   "# One\n## Two\n### Three\n#### Four",
			want: "<body><h1>One</h1><h2>Two</h2><h2>Three</h2><h2>Four</h2></body>",
		},
		{
			name: "no cosmetic newlines between block tags",
			md:   "# One\n\n## Two\n\n- a\n",
			want: "<body><h1>One</h1><h2>Two</h2><ul><li>a</li></ul></body>",
		},
		{
			name: "unordered list",
			md:   "- a\n- b\n* c\n+ d",
			want: "<body><ul><li>a</li><li>b</li><li>c</li><li>d</li></ul></body>",
		},
		{
			name: "ordered list",
			md:   "1. a\n2. b\n10) c",
			want: "<body><ol><li>a</li><li>b</li><li>c</li></ol></body>",
		},
		{
			name: "nested list by indentation",
			md:   "- a\n  - b\n    - c\n- d",
			want: "<body><ul><li>a<ul><li>b<ul><li>c</li></ul></li></ul></li><li>d</li></ul></body>",
		},
		{
			name: "adjacent lists of different type do not merge",
			md:   "- a\n\n1. b",
			want: "<body><ul><li>a</li></ul><ol><li>b</li></ol></body>",
		},
		{
			name: "thematic break is self-closing",
			md:   "a\n\n---\n\nb",
			want: "<body>a<hr/>b</body>",
		},
		{
			name: "thematic break variants",
			md:   "***\n___\n---",
			want: "<body><hr/><hr/><hr/></body>",
		},
		{
			name: "fenced code preserves newlines and indentation",
			md:   "```sh\nnpx prisma migrate deploy && node dist/index.js\n  indented\n```",
			want: "<body><pre>npx prisma migrate deploy &amp;&amp; node dist/index.js\n  indented</pre></body>",
		},
		{
			name: "unterminated fence still closes",
			md:   "```\nabc",
			want: "<body><pre>abc</pre></body>",
		},
		{
			name: "markdown inside a fence stays literal",
			md:   "```\n# not a heading\n- not a list\n```",
			want: "<body><pre># not a heading\n- not a list</pre></body>",
		},
		{
			name: "blockquote merges consecutive lines",
			md:   "> a\n> b",
			want: "<body><blockquote>a\nb</blockquote></body>",
		},
		{
			name: "inline emphasis",
			md:   "**bold** __also bold__ *em* _also em_ ~~gone~~ `code`",
			want: "<body><strong>bold</strong> <strong>also bold</strong> <em>em</em> <em>also em</em> <s>gone</s> <code>code</code></body>",
		},
		{
			name: "inline markers inside code span stay literal",
			md:   "`a**b**c`",
			want: "<body><code>a**b**c</code></body>",
		},
		{
			name: "link",
			md:   "see [docs](https://asana.com/a?b=1&c=2)",
			want: "<body>see <a href=\"https://asana.com/a?b=1&amp;c=2\">docs</a></body>",
		},
		{
			name: "image degrades to a link",
			md:   "![alt text](https://x/y.png)",
			want: "<body><a href=\"https://x/y.png\">alt text</a></body>",
		},
		{
			name: "special characters are escaped",
			md:   "a && b < c > d",
			want: "<body>a &amp;&amp; b &lt; c &gt; d</body>",
		},
		{
			name: "table degrades to text rows",
			md:   "| a | b |\n|---|---|\n| 1 | 2 |",
			want: "<body>a | b\n1 | 2</body>",
		},
		{
			name: "html tags in source are escaped, not passed through",
			md:   "<script>alert(1)</script>",
			want: "<body>&lt;script&gt;alert(1)&lt;/script&gt;</body>",
		},
		{
			name: "emphasis inside a list item",
			md:   "- **a** and [b](https://x)",
			want: "<body><ul><li><strong>a</strong> and <a href=\"https://x\">b</a></li></ul></body>",
		},
		{
			name: "heading strips trailing hashes",
			md:   "## Two ##",
			want: "<body><h2>Two</h2></body>",
		},
		{
			name: "crlf input",
			md:   "# One\r\n\r\ntext\r\n",
			want: "<body><h1>One</h1>text</body>",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FromMarkdown(tc.md)
			if got != tc.want {
				t.Fatalf("FromMarkdown(%q)\n got: %q\nwant: %q", tc.md, got, tc.want)
			}
		})
	}
}

func TestFromMarkdownAlwaysValidates(t *testing.T) {
	inputs := []string{
		"",
		"# H\n\n- a\n  - b\n\n> q\n\n```\nx & y\n```\n\n---\n\n| a | b |\n|--|--|\n| 1 | 2 |",
		"unclosed **bold and `code",
		"<div>raw & html</div>",
		"[link](https://x?a=1&b=2)",
	}
	for _, in := range inputs {
		if err := Validate(FromMarkdown(in)); err != nil {
			t.Fatalf("FromMarkdown(%q) produced invalid HTML: %v", in, err)
		}
	}
}
