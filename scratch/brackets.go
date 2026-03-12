package scratch

// IsBalanced reports whether all bracket pairs in s are properly nested and closed.
// supports (), [], {}
func IsBalanced(s string) bool {
	// maps each closing bracket to its expected opener
	pair := map[rune]rune{
		')': '(',
		']': '[',
		'}': '{',
	}

	openers := map[rune]bool{'(': true, '[': true, '{': true}

	stack := make([]rune, 0, len(s)/2)

	for _, ch := range s {
		switch {
		case openers[ch]:
			stack = append(stack, ch)
		case pair[ch] != 0:
			// closing bracket — stack must be non-empty and top must match
			if len(stack) == 0 || stack[len(stack)-1] != pair[ch] {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}

	return len(stack) == 0
}
