package commands

import "testing"

func TestPrepareBody(t *testing.T) {
	cases := []struct {
		name     string
		text     string
		format   bodyFormat
		wantHTML bool
		want     string
	}{
		{"plain passes through", "# not converted", formatPlain, false, "# not converted"},
		{"markdown converts", "# Title", formatMarkdown, true, "<body><h1>Title</h1></body>"},
		{"valid html passes through", "<body><h1>a</h1></body>", formatHTML, true, "<body><h1>a</h1></body>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, html := prepareBody(tc.text, tc.format)
			if got != tc.want || html != tc.wantHTML {
				t.Fatalf("prepareBody(%q, %v) = (%q, %v), want (%q, %v)",
					tc.text, tc.format, got, html, tc.want, tc.wantHTML)
			}
		})
	}
}

func TestTrimKeepsMarkdownHeadings(t *testing.T) {
	editorBuffer := "# Heading\n\ntext\n\n# =================================== \n# Task name\n#\n# old notes\n"

	if got, want := trim(editorBuffer, formatMarkdown), "# Heading\n\ntext"; got != want {
		t.Fatalf("trim(markdown) = %q, want %q", got, want)
	}
	// Plain mode keeps dropping every '#' line, as it always has.
	if got, want := trim(editorBuffer, formatPlain), "text"; got != want {
		t.Fatalf("trim(plain) = %q, want %q", got, want)
	}
}

func TestTrimWithoutTemplateBlock(t *testing.T) {
	if got, want := trim("\n\n# Heading\n", formatMarkdown), "# Heading"; got != want {
		t.Fatalf("trim = %q, want %q", got, want)
	}
}

func TestTrimNoLongerEscapesNewlines(t *testing.T) {
	// CommentTo now encodes the payload as JSON, so pre-escaping here would
	// reach Asana as a literal backslash-n.
	if got, want := trim("a\nb\n", formatPlain), "a\nb"; got != want {
		t.Fatalf("trim = %q, want %q", got, want)
	}
}

func TestDescribeFormat(t *testing.T) {
	if got := describeFormat(true); got != "html_notes" {
		t.Fatalf("describeFormat(true) = %q", got)
	}
	if got := describeFormat(false); got != "notes" {
		t.Fatalf("describeFormat(false) = %q", got)
	}
}
