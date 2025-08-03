package database

import (
	"database/sql"
	"fmt"
	"log"
)

const DefaultBatchSize = 1000

func LoadLandAddressesBatch(db *sql.DB, batchSize int, processor func([]string) error) error {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	totalProcessed := 0
	offset := 0

	for {
		// DB에서 배치 단위로 데이터 가져오기
		query := "SELECT address FROM land WHERE address IS NOT NULL AND address != '' ORDER BY full_code LIMIT $1 OFFSET $2"

		rows, err := db.Query(query, batchSize, offset)
		if err != nil {
			return fmt.Errorf("failed to query land addresses at offset %d: %w", offset, err)
		}

		batch := make([]string, 0, batchSize)

		// 현재 배치의 데이터 읽기
		for rows.Next() {
			var address string
			if err := rows.Scan(&address); err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan address: %w", err)
			}
			batch = append(batch, address)
		}

		rows.Close()

		// 더 이상 데이터가 없으면 종료
		if len(batch) == 0 {
			break
		}

		// 배치 처리
		if err := processor(batch); err != nil {
			return fmt.Errorf("failed to process batch at offset %d: %w", offset, err)
		}

		totalProcessed += len(batch)
		log.Printf("Processed batch: %d addresses (offset: %d, total: %d)", len(batch), offset, totalProcessed)

		// 다음 배치로 이동
		offset += batchSize

		// 배치 크기보다 적게 가져왔다면 마지막 배치
		if len(batch) < batchSize {
			break
		}
	}

	log.Printf("Total addresses processed: %d", totalProcessed)
	return nil
}
