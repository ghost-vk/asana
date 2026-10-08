package richtext

import "testing"

func TestValidateAccepts(t *testing.T) {
	cases := []struct {
		name string
		html string
	}{
		{"minimal", "<body>hello</body>"},
		{"all supported tags", "<body><h1>a</h1><h2>b</h2><ul><li>c</li></ul><ol><li>d</li></ol>" +
			"<strong>e</strong><em>f</em><u>g</u><s>h</s><code>i</code><pre>j</pre>" +
			"<blockquote>k</blockquote><a href=\"https://x\">l</a><hr/></body>"},
		{"escaped entities", "<body>a &amp; b &lt; c &gt; d</body>"},
		{"nested lists", "<body><ul><li>a<ul><li>b</li></ul></li></ul></body>"},
		{"leading whitespace", "\n  <body>hi</body>\n"},
		{"task mention", "<body><ol><li><a data-asana-gid=\"1219065173477806\"/></li></ol></body>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := Validate(tc.html); err != nil {
				t.Fatalf("Validate(%q) = %v, want nil", tc.html, err)
			}
		})
	}
}

func TestValidateRejects(t *testing.T) {
	cases := []struct {
		name string
		html string
		want string
	}{
		{"empty", "", "empty"},
		{"no body root", "<h1>hi</h1>", "<body>"},
		{"wrong root", "<div><h1>hi</h1></div>", "<body>"},
		{"unsupported p", "<body><p>hi</p></body>", "<p>"},
		{"unsupported h3", "<body><h3>hi</h3></body>", "<h3>"},
		{"unsupported table", "<body><table><tr><td>x</td></tr></table></body>", "<table>"},
		{"unclosed hr", "<body><hr></body>", "<hr/>"},
		{"unclosed br", "<body>a<br>b</body>", "<br/>"},
		{"unescaped ampersand", "<body>a && b</body>", "escape"},
		{"unbalanced tag", "<body><strong>hi</body>", "invalid"},
		{"trailing junk", "<body>hi</body>tail", "<body>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.html)
			if err == nil {
				t.Fatalf("Validate(%q) = nil, want error mentioning %q", tc.html, tc.want)
			}
			if !contains(err.Error(), tc.want) {
				t.Fatalf("Validate(%q) error = %q, want it to mention %q", tc.html, err, tc.want)
			}
		})
	}
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
