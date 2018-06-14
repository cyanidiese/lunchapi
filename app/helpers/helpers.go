package helpers

func UniqueInt(m map[int]int) map[int]int {
	n := make(map[int]int, len(m))
	ref := make(map[int]bool, len(m))
	for k, v := range m {
		if _, ok := ref[v]; !ok {
			ref[v] = true
			n[k] = v
		}
	}
	return n
}