package main

import "fmt"

func summaryRanges(nums []int) []string {
	ans := make([]string, 0)
	begin := 0
	end := 0

	for i := 0; i < len(nums); i++{
		begin = nums[i]
		for i + 1 < len(nums) && nums[i+1] - nums[i] == 1{
			i++
		}
		end = nums[i]
		if begin == end{
			ans = append(ans, fmt.Sprint(begin))
		}else{
			ans = append(ans, fmt.Sprintf("%d->%d", begin, end))
		}
	}
	return ans
}


// accимптотика O(n) - n количество элементов в nums
// сложность по памяти O(n) - n количество элементов в nums
