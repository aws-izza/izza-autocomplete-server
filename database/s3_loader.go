package database

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	S3Bucket   = "izza-test-data"
	S3Key      = "land-address/extracted_addresses.zip"
	S3Region   = "ap-northeast-2"
	TempDir    = "/tmp/land-addresses"
	ZipFile    = "/tmp/land-addresses.zip"
)

func LoadLandAddressesFromS3Batch(batchSize int, processor func([]string) error) error {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	// S3에서 ZIP 파일 다운로드
	if err := downloadZipFromS3(); err != nil {
		return fmt.Errorf("failed to download ZIP from S3: %w", err)
	}

	// ZIP 파일 압축 해제
	if err := extractZip(); err != nil {
		return fmt.Errorf("failed to extract ZIP file: %w", err)
	}

	// TXT 파일들에서 주소 데이터 읽기
	if err := processTextFiles(batchSize, processor); err != nil {
		return fmt.Errorf("failed to process text files: %w", err)
	}

	// 임시 파일들 정리
	cleanupTempFiles()

	return nil
}

func downloadZipFromS3() error {
	log.Println("Downloading ZIP file from S3...")

	// AWS 세션 생성
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(S3Region),
	})
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// 임시 디렉토리 생성
	if err := os.MkdirAll(filepath.Dir(ZipFile), 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// 파일 생성
	file, err := os.Create(ZipFile)
	if err != nil {
		return fmt.Errorf("failed to create ZIP file: %w", err)
	}
	defer file.Close()

	// S3 다운로더 생성
	downloader := s3manager.NewDownloader(sess)

	// S3에서 파일 다운로드
	numBytes, err := downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(S3Bucket),
		Key:    aws.String(S3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to download file from S3: %w", err)
	}

	log.Printf("ZIP file downloaded successfully (%d bytes)", numBytes)
	return nil
}

func extractZip() error {
	log.Println("Extracting ZIP file...")

	// ZIP 파일 열기
	r, err := zip.OpenReader(ZipFile)
	if err != nil {
		return fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer r.Close()

	// 압축 해제 디렉토리 생성
	if err := os.MkdirAll(TempDir, 0755); err != nil {
		return fmt.Errorf("failed to create extraction directory: %w", err)
	}

	// 각 파일 압축 해제
	for _, f := range r.File {
		// 디렉토리 트래버셜 공격 방지
		if strings.Contains(f.Name, "..") {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s in ZIP: %w", f.Name, err)
		}

		path := filepath.Join(TempDir, f.Name)

		// 디렉토리인 경우
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.FileInfo().Mode())
			rc.Close()
			continue
		}

		// 파일인 경우
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			rc.Close()
			return fmt.Errorf("failed to create directory for %s: %w", path, err)
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create file %s: %w", path, err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	log.Println("ZIP file extracted successfully")
	return nil
}

func processTextFiles(batchSize int, processor func([]string) error) error {
	log.Println("Processing text files...")

	// TXT 파일들 찾기
	txtFiles, err := findTextFiles(TempDir)
	if err != nil {
		return fmt.Errorf("failed to find text files: %w", err)
	}

	if len(txtFiles) == 0 {
		return fmt.Errorf("no text files found in extracted directory")
	}

	log.Printf("Found %d text files to process", len(txtFiles))

	totalProcessed := 0
	batch := make([]string, 0, batchSize)

	// 각 TXT 파일 처리
	for _, txtFile := range txtFiles {
		if err := processTextFile(txtFile, batchSize, &batch, processor, &totalProcessed); err != nil {
			return fmt.Errorf("failed to process file %s: %w", txtFile, err)
		}
	}

	// 남은 배치 처리
	if len(batch) > 0 {
		if err := processor(batch); err != nil {
			return fmt.Errorf("failed to process final batch: %w", err)
		}
		totalProcessed += len(batch)
		log.Printf("Processed final batch: %d addresses", len(batch))
	}

	log.Printf("Total addresses processed: %d", totalProcessed)
	return nil
}

func findTextFiles(dir string) ([]string, error) {
	var txtFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// __MACOSX 폴더와 숨김 파일들 제외
		if strings.Contains(path, "__MACOSX") || strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".txt") {
			txtFiles = append(txtFiles, path)
		}

		return nil
	})

	return txtFiles, err
}

func processTextFile(filename string, batchSize int, batch *[]string, processor func([]string) error, totalProcessed *int) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	log.Printf("Processing file: %s", filename)

	scanner := bufio.NewScanner(file)
	fileProcessed := 0

	for scanner.Scan() {
		rawLine := scanner.Text()
		address := strings.TrimSpace(rawLine)
		
		// 디버그 로그: 문제가 될 수 있는 라인들 출력
		if len(address) < 2 {
			log.Printf("DEBUG: Skipping short/empty line in file %s - Raw: %q, Trimmed: %q, Length: %d", 
				filepath.Base(filename), rawLine, address, len(address))
			continue
		}
		
		// 특수문자나 제어문자 확인
		if len(address) != len(strings.TrimSpace(strings.ReplaceAll(address, " ", ""))) {
			log.Printf("DEBUG: Potential special characters in file %s - Address: %q, Length: %d", 
				filepath.Base(filename), address, len(address))
		}

		*batch = append(*batch, address)
		fileProcessed++

		// 배치가 가득 찼으면 처리
		if len(*batch) >= batchSize {
			if err := processor(*batch); err != nil {
				return fmt.Errorf("failed to process batch: %w", err)
			}

			*totalProcessed += len(*batch)
			log.Printf("Processed batch: %d addresses (file: %s, file total: %d, overall total: %d)", 
				len(*batch), filepath.Base(filename), fileProcessed, *totalProcessed)

			// 배치 초기화
			*batch = (*batch)[:0]
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	log.Printf("Completed file: %s (processed %d addresses)", filename, fileProcessed)
	return nil
}

func cleanupTempFiles() {
	log.Println("Cleaning up temporary files...")

	// ZIP 파일 삭제
	if err := os.Remove(ZipFile); err != nil {
		log.Printf("Warning: failed to remove ZIP file: %v", err)
	}

	// 압축 해제된 디렉토리 삭제
	if err := os.RemoveAll(TempDir); err != nil {
		log.Printf("Warning: failed to remove temp directory: %v", err)
	}

	log.Println("Cleanup completed")
}