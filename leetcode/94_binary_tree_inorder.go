package main

func inorderTraversal(root *TreeNode) []int {
	result := make([]int, 0)
	Traversal(root, &result)
	return result
}

func Traversal(node *TreeNode, result *[]int) {
	if node == nil {
		return
	}

	Traversal(node.Left, result)
	*result = append(*result, node.Val)
	Traversal(node.Right, result)
}


// сложность O(n), где n - количество вершин
// сложность по памяти O(h), где h - высота дерева или в худшем случае h == n
