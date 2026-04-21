package utils

// SameStringSet returns true when both slices contain the same strings with the same frequencies,
// regardless of order.
func SameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]int, len(a))
	for _, v := range a {
		m[v]++
	}

	for _, v := range b {
		if m[v] == 0 {
			return false
		}
		m[v]--
	}

	return true
}
