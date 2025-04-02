package leetcode

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func rangeSumBST(root *TreeNode, low int, high int) int {
	sum := 0
	for val := low; val <= high; val++ {
		if finder(root, val) {
			sum += val
		}
	}

	return sum

}

func finder(root *TreeNode, target int) bool {
	if root == nil {
		return false
	}

	if target > root.Val {
		return finder(root.Right, target)
	} else if target < root.Val {
		return finder(root.Left, target)
	} else if target == root.Val {
		return true
	}
	return false
}


// ассимптотическая сложность O(m * n), где m - количество чисел в диапазоне low - high, n - высота дерева
// сложность по памяти O(n), где n - высота дерева, тут оценивается глубина рекурсии
