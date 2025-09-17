package l2

import (
	"sort"
	"strings"
)

type group struct {
	first string
	set   map[string]struct{}
}

func signature(s string) string {
	r := []rune(s)
	sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })
	return string(r)
}

func groupWords(words []string) map[string]*group {
	groups := make(map[string]*group)
	for _, w := range words {
		word := strings.ToLower(w)
		key := signature(word)
		if _, ok := groups[key]; !ok {
			groups[key] = &group{
				first: word,
				set:   make(map[string]struct{}),
			}
		}
		groups[key].set[word] = struct{}{}
	}
	return groups
}

func FindAnagramSets(words []string) map[string][]string {
	groups := groupWords(words)

	result := make(map[string][]string)
	for _, gr := range groups {
		if len(gr.set) < 2 {
			continue
		}
		lst := make([]string, 0, len(gr.set))
		for w := range gr.set {
			lst = append(lst, w)
		}
		sort.Strings(lst)
		result[gr.first] = lst
	}

	return result
}
