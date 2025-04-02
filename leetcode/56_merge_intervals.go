package leetcode

import "sort"

func MergeIntervals(intervals [][]int) [][]int {
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})

	solve := make([][]int, 0)
	n := len(intervals)

	for i := 0; i < n; i++ {
		if len(solve) == 0 {
			solve = append(solve, intervals[i])
			continue
		}

		last := solve[len(solve)-1]

		if last[1] >= intervals[i][0] {
			if last[1] < intervals[i][1] {
				last[1] = intervals[i][1]
			}
		} else {
			solve = append(solve, intervals[i])
		}
	}

	return solve
}

// ассимптотическая сложность O(n log n), где n - количество интервалов
// сложность по памяти O(n), где n - количество интервалов после слияния
