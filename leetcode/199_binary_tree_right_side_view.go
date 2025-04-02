package leetcode


func rightSideView(root *TreeNode) []int {
    return levelorder(root, 0, []int{})
}

func levelorder(node *TreeNode, level int, result []int) []int{
	if node == nil{
		return result
	}

	if level == len(result){
		result = append(result, 0)
	}

	result[level] = node.Val
	result = levelorder(node.Left, level + 1, result)
	result = levelorder(node.Right, level + 1, result)

	return result
}

// сложность O(n), где n - количество вершин в дереве
// сложность по памяти O(n), где n - количество вершин