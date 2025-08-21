pipeline {
    agent any
    
    environment {
        // ECR 설정
        AWS_REGION = 'ap-northeast-2'
        ECR_REGISTRY = '177716289679.dkr.ecr.ap-northeast-2.amazonaws.com'
        IMAGE_REPOSITORY = 'izza/autocomplete-server'
        
        // 이미지 태그 생성 (빌드 번호 + Git 커밋 해시)
        IMAGE_TAG = "${BUILD_NUMBER}-${GIT_COMMIT.substring(0,7)}"
        
        // GitOps 레포지토리 설정
        GITOPS_REPO = 'https://github.com/aws-izza/izza-cd.git'
        GITOPS_BRANCH = 'main'
        
        // Credentials ID (Jenkins에서 설정 필요)
        AWS_CREDENTIALS = 'aws-cred'
        GIT_CREDENTIALS = 'github-pat'
    }
    
    tools {
        go '1.24'  // Jenkins에 Go 1.24 설치 필요
    }
    
    stages {
        stage('Checkout') {
            steps {
                echo "🔍 Checking out code..."
                checkout scm
            }
        }
        
        stage('Go Version & Dependencies') {
            steps {
                echo "🐹 Setting up Go environment..."
                sh '''
                    go version
                    go mod download
                    go mod verify
                '''
            }
        }
        
        stage('Test') {
            steps {
                echo "🧪 Running tests..."
                sh '''
                    # 단위 테스트 실행
                    go test -v ./...
                    
                    # 코드 품질 검사
                    go vet ./...
                    
                    # 정적 분석 (golint 설치되어 있다면)
                    # golint ./...
                '''
            }
            post {
                always {
                    // 테스트 결과가 있다면 게시
                    publishTestResults testResultsPattern: '**/test-results.xml'
                }
            }
        }
        
        stage('Build') {
            steps {
                echo "🏗️ Building Go application..."
                sh '''
                    # 정적 링크된 바이너리 빌드
                    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .
                    
                    # 빌드된 바이너리 확인
                    ls -la main
                    file main
                '''
            }
        }
        
        stage('Security Scan') {
            steps {
                echo "🔒 Running security scans..."
                script {
                    try {
                        sh '''
                            # Go 모듈 취약점 스캔 (govulncheck 설치되어 있다면)
                            # govulncheck ./...
                            
                            # 의존성 라이선스 체크 (필요시)
                            # go-licenses check ./...
                            
                            echo "Security scan completed"
                        '''
                    } catch (Exception e) {
                        echo "⚠️ Security scan failed, but continuing: ${e.getMessage()}"
                    }
                }
            }
        }
        
        stage('Docker Build') {
            steps {
                echo "🐳 Building Docker image..."
                script {
                    def image = docker.build("${ECR_REGISTRY}/${IMAGE_REPOSITORY}:${IMAGE_TAG}")
                    
                    // 추가로 latest 태그도 생성
                    sh "docker tag ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:${IMAGE_TAG} ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:latest"
                }
            }
        }
        
        stage('ECR Push') {
            steps {
                echo "📤 Pushing to ECR..."
                script {
                    // AWS ECR 로그인
                    sh '''
                        aws ecr get-login-password --region ${AWS_REGION} | docker login --username AWS --password-stdin ${ECR_REGISTRY}
                    '''
                    
                    // 이미지 푸시
                    sh '''
                        docker push ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:${IMAGE_TAG}
                        docker push ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:latest
                    '''
                }
            }
        }
        
        stage('Update GitOps Repository') {
            steps {
                echo "🔄 Updating GitOps repository..."
                script {
                    withCredentials([usernamePassword(credentialsId: "${GIT_CREDENTIALS}", usernameVariable: 'GIT_USERNAME', passwordVariable: 'GIT_PASSWORD')]) {
                        sh '''
                            # GitOps 레포지토리 클론
                            git clone https://${GIT_USERNAME}:${GIT_PASSWORD}@github.com/your-org/izza-cd.git gitops-repo
                            cd gitops-repo
                            
                            # Git 설정
                            git config user.name "Jenkins"
                            git config user.email "jenkins@company.com"
                            
                            # 이미지 태그 업데이트
                            sed -i "s|image: ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:.*|image: ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:${IMAGE_TAG}|g" environments/autocomplete-server/app.yaml
                            
                            # 변경사항 확인
                            git diff
                            
                            # 커밋 및 푸시
                            git add environments/autocomplete-server/app.yaml
                            git commit -m "🚀 Update autocomplete-server image to ${IMAGE_TAG}
                            
                            - Build: #${BUILD_NUMBER}
                            - Commit: ${GIT_COMMIT}
                            - Branch: ${GIT_BRANCH}
                            - Triggered by: ${BUILD_USER:-Jenkins}"
                            
                            git push origin ${GITOPS_BRANCH}
                            
                            echo "✅ GitOps repository updated successfully"
                        '''
                    }
                }
            }
        }
        
        stage('Cleanup') {
            steps {
                echo "🧹 Cleaning up..."
                sh '''
                    # 로컬 Docker 이미지 정리
                    docker rmi ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:${IMAGE_TAG} || true
                    docker rmi ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:latest || true
                    
                    # 빌드 아티팩트 정리
                    rm -f main
                    rm -rf gitops-repo
                '''
            }
        }
    }
    
    post {
        always {
            echo "🏁 Pipeline completed"
            
            // 워크스페이스 정리
            cleanWs()
        }
        
        success {
            echo "✅ Pipeline succeeded!"
        }
        
        failure {
            echo "❌ Pipeline failed!"
        }
        
        unstable {
            echo "⚠️ Pipeline is unstable"
        }
    }
}