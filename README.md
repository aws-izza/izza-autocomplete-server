# 개요

Kiro IDE와 함께한 인메모리 Trie 자료구조로 빠른 주소 검색을 제공하는 Go 기반 API 서버

## 설치 및 실행

1. 의존성 설치:

```bash
go mod tidy
```

2. 환경변수 설정:

```bash
cp .env.example .env
# .env 파일을 편집하여 설정 입력
```

3. 데이터베이스 설정 방법:

**방법 1: AWS Secrets Manager 사용 (권장)**

```bash
# .env 파일에서 설정
AWS_REGION=ap-northeast-2
DB_SECRET_NAME=your-secret-name

# DB 연결 정보 (host, port, dbname, sslmode)
DB_HOST=your-db-host
DB_PORT=5432
DB_NAME=your-database
DB_SSLMODE=require
```

**방법 2: 환경변수 직접 설정**

```bash
# DB_SECRET_NAME을 비워두거나 주석 처리하고 직접 설정
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=your_database
```

4. 애플리케이션 실행:

```bash
go run main.go
```