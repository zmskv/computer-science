package leetcode

func levelOrder(root *TreeNode) [][]int {
	return preorder(root, 0, [][]int{})
}

func preorder(node *TreeNode, level int, result [][]int) [][]int {
	if node == nil {
		return result
	}
	if level == len(result) {
		result = append(result, []int{})
	}
	result[level] = append(result[level], node.Val)
	result = preorder(node.Left, level+1, result)
	result = preorder(node.Right, level+1, result)
	return result
}


// сложность O(n), где n - количество вершин
// сложность по памяти O(n), где n - количество вершин