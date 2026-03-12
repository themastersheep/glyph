package scratch

import "testing"

func TestIsBalanced(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// happy path
		{"empty string", "", true},
		{"single pair parens", "()", true},
		{"single pair square", "[]", true},
		{"single pair curly", "{}", true},
		{"nested same type", "((()))", true},
		{"nested mixed", "({[]})", true},
		{"sequential pairs", "()[]{}",  true},
		{"complex nested", "{[()()]}", true},
		{"non-bracket chars ignored", "a(b[c]d)e", true},
		{"brackets in code-like string", "func foo() { return bar[0] }", true},

		// failure cases
		{"unclosed open", "(", false},
		{"extra close", ")", false},
		{"wrong order", "([)]", false},
		{"reversed nesting", "}{", false},
		{"opens without close", "(()", false},
		{"close before open", "][]", false},
		{"interleaved wrong", "[(])", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBalanced(tt.input)
			if got != tt.want {
				t.Errorf("IsBalanced(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
