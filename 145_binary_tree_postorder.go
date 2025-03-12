package main

func postorderTraversal(root *TreeNode) []int {
	result := make([]int, 0)
	travers(root, &result)
	return result
}

func travers(node *TreeNode, result *[]int) {
	if node == nil {
		return
	}
	travers(node.Left, result)
	travers(node.Right, result)
	*result = append(*result, node.Val)
}

// сложность O(n), где n - количество вершин
// сложность по памяти O(h), где h - высота дерева или в худшем случае h == n
