package trie

type JumpNode struct {
	Ref []*FullNode
}

func CreateJumpNode() JumpNode {
	return JumpNode{Ref: make([]*FullNode, 0)}
}

func (node *JumpNode) Insert(root *FullNode, word string, depth int) {
	searchKey := node.buildSearchKeyByDepth(word, depth)

	if searchKey == "" {
		return
	}

	targetNode := root.searchNode(searchKey)

	if targetNode != nil {
		if !node.containsRef(targetNode) {
			node.Ref = append(node.Ref, targetNode)
		}
	}
}

func (node *JumpNode) buildSearchKeyByDepth(word string, depth int) string {
	runes := []rune(word)
	spaceCount := 0

	for i, char := range runes {
		if char == ' ' {
			spaceCount++
			if spaceCount == depth {
				return string(runes[:i+2])
			}
		}
	}

	return ""
}

func (node *JumpNode) containsRef(target *FullNode) bool {
	for _, ref := range node.Ref {
		if ref == target {
			return true
		}
	}
	return false
}

func (node *JumpNode) Search(results *[]string, word string) {
	for _, refNode := range node.Ref {
		runeWord := []rune(word)

		refNode.searchInMiddle(results, runeWord)
		if len(*results) >= 5 {
			break
		}
	}
}
