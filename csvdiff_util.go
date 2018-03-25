package csvdiff

// IndexOf returns the index of an element in a slice or -1 otherwise.
func IndexOf(slice []string, s string) int {
	for i, v := range slice {
		if v == s {
			return i
		}
	}
	return -1
}

// OrderedSet returns a set from a slice
func OrderedSet(slices ...[]string) []string {
	var set []string
	var m map[string]bool
	for _, slice := range slices {
		for _, s := range slice {
			if _, exists := m[s]; !exists {
				set = append(set, s)
				m[s] = true
			}
		}
	}
	return set
}
