package richtext

import "testing"

func TestEdgeCasesStayValid(t *testing.T) {
	inputs := []string{
		"- \n",
		"#\n",
		"```\n```",
		"> ",
		"|",
		"*a*b*c*",
		"[](())",
		"a_b_c_d",
		"____",
		"1. a\n- b\n1. c",
		"\t- tabbed",
		"- a\n    - deep with no middle",
		"~~~\nfence\n~~~",
		"![](x)",
		"**", "___", "`",
	}
	for _, in := range inputs {
		out := FromMarkdown(in)
		if err := Validate(out); err != nil {
			t.Errorf("FromMarkdown(%q) = %q -> invalid: %v", in, out, err)
		}
	}
}
