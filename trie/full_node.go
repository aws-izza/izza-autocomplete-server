package trie

type FullNode struct {
	Value    rune
	Parent   *FullNode
	Children []*FullNode
	IsEnd    bool
}

func (node *FullNode) Insert(word string) {
	node.insertInternal([]rune(word), 0)
}

func (node *FullNode) insertInternal(word []rune, depth int) {
	if depth == len(word) {
		node.IsEnd = true
		return
	}

	if node.Children == nil {
		node.Children = make([]*FullNode, 0)
	}

	var nextChild *FullNode

	for _, child := range node.Children {
		if word[depth] == child.Value {
			nextChild = child
			break
		}
	}

	if nextChild == nil {
		nextChild = &FullNode{Value: word[depth], Parent: node}
		node.Children = append(node.Children, nextChild)
	}

	nextChild.insertInternal(word, depth+1)
}

func (node *FullNode) Search(results *[]string, word string) {
	runeWord := []rune(word)
	node.searchInternal(results, runeWord, 0, "")
}

func (node *FullNode) searchInternal(results *[]string, word []rune, depth int, result string) {
	if depth != 0 {
		result += string(node.Value)
	}

	if depth < len(word)-1 {
		for _, child := range node.Children {
			if child.Value == word[depth] {
				child.searchInternal(results, word, depth+1, result)
				break
			}
		}
	} else {
		if node.IsEnd {
			*results = append(*results, result)
		}

		if node.Children == nil {
			return
		}

		for _, child := range node.Children {
			if len(*results) >= 5 {
				return
			}
			child.searchInternal(results, word, depth+1, result)
		}
	}
}

func (node *FullNode) searchNode(word string) *FullNode {
	nWord := []rune(word)
	return node.searchNodeInternal(nWord, 0)
}

func (node *FullNode) searchNodeInternal(word []rune, depth int) *FullNode {
	if depth == len(word) {
		return node
	}

	for _, child := range node.Children {
		if child.Value == word[depth] {
			return child.searchNodeInternal(word, depth+1)
		}
	}

	return nil
}

func (node *FullNode) searchInMiddle(results *[]string, word []rune) {
	if node.Value != word[0] {
		return
	}
	result := node.combineParentValues()

	node.searchInternal(results, word, 1, result)
}

func (node *FullNode) combineParentValues() string {
	return node.Parent.combineParentsInternal("")
}

func (node *FullNode) combineParentsInternal(result string) string {
	if node.Parent != nil {
		result = node.Parent.combineParentsInternal(result)
	}
	if node.Parent == nil {
		return result
	}
	return result + string(node.Value)
}
