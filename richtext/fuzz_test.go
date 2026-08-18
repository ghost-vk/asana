package richtext

import (
	"math/rand"
	"strings"
	"testing"
)

// TestFromMarkdownNeverEmitsInvalidHTML guards the invariant the whole feature
// leans on: whatever the user feeds --md, Asana must not answer with
// xml_parsing_error.
func TestFromMarkdownNeverEmitsInvalidHTML(t *testing.T) {
	tokens := []string{
		"#", "##", "###", "-", "*", "+", "1.", "2)", ">", "```", "~~~", "---", "___", "***",
		"|", "|a|b|", "|---|", "**", "__", "~~", "`", "[", "]", "(", ")", "!", "&", "<", ">",
		"\"", "'", "\\", "текст", "word", "  ", "\t", "", "a_b", "<p>", "</body>", "<hr>",
	}
	rnd := rand.New(rand.NewSource(1))
	for i := 0; i < 3000; i++ {
		var lines []string
		for l := rnd.Intn(6); l >= 0; l-- {
			var line []string
			for w := rnd.Intn(6); w >= 0; w-- {
				line = append(line, tokens[rnd.Intn(len(tokens))])
			}
			lines = append(lines, strings.Join(line, " "))
		}
		in := strings.Join(lines, "\n")
		out := FromMarkdown(in)
		if err := Validate(out); err != nil {
			t.Fatalf("FromMarkdown(%q)\n = %q\n -> %v", in, out, err)
		}
	}
}
