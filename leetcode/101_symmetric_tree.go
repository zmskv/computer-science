package leetcode

func isSymmetricTree(root *TreeNode) bool {
	if root == nil {
		return true
	}

	return check(root.Left, root.Right)
}

func check(left *TreeNode, right *TreeNode) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	if left.Val != right.Val {
		return false
	}

	return check(left.Left, right.Right) && check(left.Right, right.Left)
}


// сложность O(n), где n - высота дерева
// сложность по памяти O(h), где h - высота дерева, если брать полную оценку то O(2h)
