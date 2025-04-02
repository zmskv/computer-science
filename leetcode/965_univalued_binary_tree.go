package leetcode


func isUnivalTree(root *TreeNode) bool {
	return uniPreorder(root, map[int]int{})
}

func uniPreorder(node *TreeNode, m map[int]int) bool {
	if node == nil {
		return true
	}

	if _, exist := m[node.Val]; !exist {
		m[node.Val] = 1
	} else {
		m[node.Val]++
	}
	if len(m) != 1 {
		return false
	}

	return uniPreorder(node.Left, m) && uniPreorder(node.Right, m)

}

// сложность O(n), где n - количество вершин в дереве
// сложность по памяти O(n), в худшем случае количество вершин в дереве а так можно сказать что высота

