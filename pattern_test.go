package gack

import (
	"testing"
)

func TestPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		subject  string
		match    bool
		captures map[string]string
	}{
		{":everything", "everything", true, map[string]string{"everything": "everything"}},
		{"build/:foo/bar", "build/foo/boo", false, map[string]string{}},
		{"build", "build", true, map[string]string{}},
		{"build", "pkg", false, map[string]string{}},
		{"build/foo-:foo", "build/foo-arch.ext", true, map[string]string{"foo": "arch.ext"}},
		{"build/foo-:foo-bar", "build/foo-arch.ext-bar", true, map[string]string{"foo": "arch.ext"}},
	}

	for i, test := range tests {
		matcher := NewPattern(test.pattern)
		match := matcher.Match(test.subject)
		if test.match {
			if match.Matches() {
				for name, value := range test.captures {
					if match.Param(name) != value {
						t.Errorf("Test %d: Expected parameter %s to be %s but got %s", i, name, value, match.Param(name))
					}
				}

				str := match.Interpolate(test.pattern)
				if str != test.subject {
					t.Errorf("Test %d: Interpolation should have been %s but got %s", i, test.subject, str)
				}
			} else {
				t.Errorf("Test %d: Expected a match", i)
			}
		} else {
			if len(match.captures) > 0 {
				t.Errorf("Test %d: Expected no captures, got %v", i, match.captures)
			}

			if match.Matches() {
				t.Errorf("Test %d: Expected no match", i)
			}
		}
	}
}
