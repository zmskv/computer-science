package main

func hasPathSum(root *TreeNode, targetSum int) bool {
	return preorderSum(root, 0, targetSum)
}

func isLeaf(node *TreeNode) bool {
	return node.Left == nil && node.Right == nil
}

func preorderSum(node *TreeNode, sum, target int) bool {
	if node == nil {
		return false
	}
	if isLeaf(node) && sum+node.Val == target {
		return true
	}

	return preorderSum(node.Left, sum+node.Val, target) || preorderSum(node.Right, sum+node.Val, target)
}

// сложность O(n), где n - количество вершин в дереве
// сложность по памяти O(h), где h - высота дерева
