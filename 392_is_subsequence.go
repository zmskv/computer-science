package main

func isSubsequence(s, t string) bool {
	ind := 0

	for i := 0; i < len(t); i++ {
		if ind < len(s) && s[ind] == t[i] {
			ind++
		}
	}

	return ind == len(s)
}

// accимптотическая сложность O(n), где n - длина строки t
// сложность по памяти O(1)
