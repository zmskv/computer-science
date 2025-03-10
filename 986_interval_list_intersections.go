package main

func intervalIntersection(firstList [][]int, secondList [][]int) [][]int {
	ans := make([][]int, 0)

	for f, l := 0, 0; f < len(firstList) && l < len(secondList); {
		fs := firstList[f][0]
		fe := firstList[f][1]
		ss := secondList[l][0]
		se := secondList[l][1]

		start := min(fs, ss)
		end := max(fe, se)

		if start <= end {
			ans = append(ans, []int{start, end})
		}

		if fe <= se {
			f++
		} else {
			l++
		}
	}

	return ans
}

// ассимптотическая сложность O(n + m), где n - длина массива firstList, а m - длина массива secondList
// сложность по памяти O(n), где n - количество полученных интервалов
