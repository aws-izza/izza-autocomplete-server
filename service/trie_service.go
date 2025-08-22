package service

import (
	"fmt"
	"gin-project/database"
	"gin-project/trie"
	"log"
	"sync"
)

type TrieService struct {
	nodeManager *trie.NodeManager
}

var (
	instance *TrieService
	once     sync.Once
)

// GetTrieService returns singleton instance of TrieService
func GetTrieService() *TrieService {
	once.Do(func() {
		instance = &TrieService{
			nodeManager: createNodes(),
		}
	})
	return instance
}

func createNodes() *trie.NodeManager {
	nodes := trie.CreateNodes()
	return &nodes
}

func (ts *TrieService) InitializeFromS3(batchSize int) error {
	// Reset nodes
	nodes := trie.CreateNodes()
	ts.nodeManager = &nodes

	// 배치 처리 함수 정의
	processor := func(addresses []string) error {
		for _, address := range addresses {
			ts.nodeManager.Insert(address)
		}
		return nil
	}

	// S3에서 배치로 주소 로드 및 처리
	if err := database.LoadLandAddressesFromS3Batch(batchSize, processor); err != nil {
		return fmt.Errorf("failed to load addresses from S3 in batches: %w", err)
	}

	log.Println("Successfully completed loading all addresses from S3 into trie")

	// Trie 상태 출력
	ts.printTrieStatus()

	return nil
}

// InitializeFromDatabase - 기존 DB 방식 (호환성을 위해 유지)
func (ts *TrieService) InitializeFromDatabase(dbConfig database.Config, batchSize int) error {
	db, err := database.Connect(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Reset nodes
	nodes := trie.CreateNodes()
	ts.nodeManager = &nodes

	// 배치 처리 함수 정의
	processor := func(addresses []string) error {
		for _, address := range addresses {
			ts.nodeManager.Insert(address)
		}
		return nil
	}

	// 배치로 주소 로드 및 처리
	if err := database.LoadLandAddressesBatch(db, batchSize, processor); err != nil {
		return fmt.Errorf("failed to load addresses in batches: %w", err)
	}

	log.Println("Successfully completed loading all addresses into trie")

	// Trie 상태 출력
	ts.printTrieStatus()

	return nil
}

// Search performs search on the trie
func (ts *TrieService) Search(query string) []string {
	return ts.nodeManager.Search(query)
}

// printTrieStatus prints the current status of the trie
func (ts *TrieService) printTrieStatus() {
	log.Println("=== Trie Status ===")

	// MainNode 상태
	mainNodeChildrenCount := len(ts.nodeManager.MainNode.Children)
	log.Printf("MainNode - Children count: %d", mainNodeChildrenCount)

	// MainNode의 첫 번째 레벨 자식들 일부 출력
	if mainNodeChildrenCount > 0 {
		log.Printf("MainNode - First level children (first 10):")
		for i, child := range ts.nodeManager.MainNode.Children {
			if i >= 10 {
				log.Printf("  ... and %d more children", mainNodeChildrenCount-10)
				break
			}
			log.Printf("  [%d] '%c' (children: %d, isEnd: %t)", i, child.Value, len(child.Children), child.IsEnd)
		}
	}

	// SubNodes 상태
	subNodesCount := len(ts.nodeManager.SubNodes)
	log.Printf("SubNodes count: %d", subNodesCount)

	for i, subNode := range ts.nodeManager.SubNodes {
		refCount := len(subNode.Ref)
		log.Printf("SubNode[%d] - References count: %d", i, refCount)

		// 각 SubNode의 참조들 일부 출력
		if refCount > 0 {
			log.Printf("  SubNode[%d] references (first 5):", i)
			for j, ref := range subNode.Ref {
				if j >= 5 {
					log.Printf("    ... and %d more references", refCount-5)
					break
				}
				// 참조된 노드의 값과 부모 경로 출력
				path := ts.getNodePath(ref)
				log.Printf("    [%d] Node path: '%s' (isEnd: %t)", j, path, ref.IsEnd)
			}
		}
	}

	log.Println("=== End Trie Status ===")
}

// getNodePath returns the path from root to the given node
func (ts *TrieService) getNodePath(node *trie.FullNode) string {
	if node == nil {
		return ""
	}

	path := ""
	current := node
	for current != nil && current.Parent != nil {
		path = string(current.Value) + path
		current = current.Parent
	}

	return path
}
