package trie

import (
	"strings"
	"unicode"
)

type NodeManager struct {
	MainNode FullNode
	SubNodes []JumpNode
}

func CreateNodes() NodeManager {
	return NodeManager{FullNode{}, make([]JumpNode, 0)}
}

func (nodes *NodeManager) Insert(address string) {
	// 안전장치: 빈 문자열 체크
	if len(address) == 0 {
		return
	}
	
	split := strings.Split(address, " ")

	maxDepth := -1
	for i, splitAddress := range split {
		// 안전장치: 빈 부분 문자열 체크
		if len(splitAddress) == 0 {
			continue
		}
		
		runes := []rune(splitAddress)
		if len(runes) > 0 && !unicode.IsDigit(runes[0]) {
			maxDepth = i
		}
	}

	nodes.MainNode.Insert(address)

	for i := 0; i < maxDepth; i++ {
		if len(nodes.SubNodes) < i+1 {
			nodes.SubNodes = append(nodes.SubNodes, CreateJumpNode())
		}
		nodes.SubNodes[i].Insert(&nodes.MainNode, address, i+1)
	}
}

func (nodes *NodeManager) Search(query string) []string {
	results := make([]string, 0)
	nodes.MainNode.Search(&results, query)

	for _, subNodes := range nodes.SubNodes {
		if len(results) >= 5 {
			break
		}
		subNodes.Search(&results, query)
	}

	return results
}
