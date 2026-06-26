package api

import "testing"

func TestJsonVal(t *testing.T) {
	cases := map[string]string{
		"null":             "null",
		"Feature":          `"Feature"`,
		"1198862357412458": `"1198862357412458"`, // enum gid stays a quoted string, never a raw number
		"7":                `"7"`,                 // Asana coerces a quoted number for number fields
	}
	for in, want := range cases {
		if got := jsonVal(in); got != want {
			t.Errorf("jsonVal(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolveEnum(t *testing.T) {
	f := CustomFieldDef{Name: "Type", EnumOptions: []Base{{Gid: "111", Name: "Feature"}, {Gid: "222", Name: "Bug"}}}
	if got := resolveEnum(f, "bug"); got != "222" { // by name, case-insensitive
		t.Errorf("by name: got %q, want 222", got)
	}
	if got := resolveEnum(f, "111"); got != "111" { // already a gid
		t.Errorf("by gid: got %q, want 111", got)
	}
}
