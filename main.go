package main

import (
	"gin-project/database"
	"gin-project/service"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// .env 파일 로드
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// 배치 사이즈 설정
	batchSize := getBatchSize("BATCH_SIZE", database.DefaultBatchSize)
	log.Printf("Using batch size: %d", batchSize)

	// 트라이 서비스 초기화 (S3에서 데이터 로드)
	trieService := service.GetTrieService()
	if err := trieService.InitializeFromS3(batchSize); err != nil {
		log.Fatalf("Failed to initialize trie service from S3: %v", err)
	}

	// Gin 라우터 생성
	r := gin.Default()

	// CORS 설정
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 모든 도메인 허용 (프로덕션에서는 특정 도메인으로 제한)
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 기본 라우트
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Land Address Search API",
		})
	})

	// 헬스체크 엔드포인트
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// 주소 검색 엔드포인트
	r.GET("/api/v1/ac/auto-complete", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Query parameter 'q' is required",
			})
			return
		}

		results := trieService.Search(query)
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"results": results,
			},
		})
	})

	log.Println("Starting server on :8080")
	r.Run(":8080")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBatchSize(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if size, err := strconv.Atoi(value); err == nil && size > 0 {
			return size
		}
	}
	return defaultValue
}

func getPort(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if port, err := strconv.Atoi(value); err == nil && port > 0 {
			return port
		}
	}
	return defaultValue
}
