package richtext

import "testing"

func TestCompact(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "newlines between block tags are cosmetic",
			in:   "<body>\n  <h1>a</h1>\n  <h2>b</h2>\n</body>",
			want: "<body><h1>a</h1><h2>b</h2></body>",
		},
		{
			name: "a single space between inline tags is meaningful",
			in:   "<body><strong>a</strong> <em>b</em></body>",
			want: "<body><strong>a</strong> <em>b</em></body>",
		},
		{
			name: "newlines inside text are the paragraph breaks and stay",
			in:   "<body>first\n\nsecond</body>",
			want: "<body>first\n\nsecond</body>",
		},
		{
			name: "text touching a tag is untouched",
			in:   "<body>text <a href=\"x\">l</a> tail</body>",
			want: "<body>text <a href=\"x\">l</a> tail</body>",
		},
		{
			name: "pre keeps its newlines and indentation",
			in:   "<body>\n<pre>a\n  b\n</pre>\n<h1>x</h1>\n</body>",
			want: "<body><pre>a\n  b\n</pre><h1>x</h1></body>",
		},
		{
			name: "blank line before a tag inside pre survives",
			in:   "<body><pre>a\n\n<em>not a tag</em></pre></body>",
			want: "<body><pre>a\n\n<em>not a tag</em></pre></body>",
		},
		{
			name: "already compact input is unchanged",
			in:   "<body><h1>a</h1><ul><li>b</li></ul></body>",
			want: "<body><h1>a</h1><ul><li>b</li></ul></body>",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Compact(tc.in); got != tc.want {
				t.Fatalf("Compact(%q)\n got: %q\nwant: %q", tc.in, got, tc.want)
			}
		})
	}
}
