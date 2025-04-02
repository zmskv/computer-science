package leetcode

func preorderTraversal(root *TreeNode) []int {
	result := make([]int, 0)
	traversal(root, &result)
	return result
}

func traversal(node *TreeNode, result *[]int) {
	if node == nil {
		return
	}
	*result = append(*result, node.Val)
	traversal(node.Left, result)
	traversal(node.Right, result)
}


// сложность O(n), где n - количество вершин в дереве
// сложность по памяти O(h), где h - высота дерева