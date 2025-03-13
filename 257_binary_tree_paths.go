package main 


import (
    "strings"
    "strconv"
)

func IsLeaf(node *TreeNode) bool{
    return node.Left == nil && node.Right == nil
}
func binaryTreePaths(root *TreeNode) []string {
    return preorderPath(root, []string{}, []string{})
}

func preorderPath(node *TreeNode, cur []string, result []string) []string{
    if node == nil{
        return result
    }
    if IsLeaf(node) {
        cur = append(cur, strconv.Itoa(node.Val))
        result = append(result, strings.Join(cur, "->"))
        cur = []string{}
        return result
    }

    cur = append(cur, strconv.Itoa(node.Val))
    result = preorderPath(node.Left, cur, result)
    result = preorderPath(node.Right, cur, result)
    return result
}


// сложность O(n), где n - количество вершин в дереве 
// сложность по памяти  O(h), где h - высота дерева
