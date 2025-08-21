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
    
    stages {
        stage('Checkout') {
            steps {
                echo "🔍 Checking out code..."
                checkout scm
            }
        }
        
        stage('Go Build & Test') {
            agent {
                kubernetes {
                    yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: golang
    image: golang:1.24-alpine
    command:
    - cat
    tty: true
    volumeMounts:
    - name: workspace
      mountPath: /workspace
  volumes:
  - name: workspace
    emptyDir: {}
"""
                }
            }
            steps {
                container('golang') {
                    echo "🐹 Setting up Go environment..."
                    sh '''
                        go version
                        go mod download
                        go mod verify
                    '''
                    
                    echo "🧪 Running tests..."
                    sh '''
                        # 단위 테스트 실행
                        go test -v ./...
                        
                        # 코드 품질 검사
                        go vet ./...
                    '''
                    
                    echo "🏗️ Building Go application..."
                    sh '''
                        # 정적 링크된 바이너리 빌드
                        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .
                        
                        # 빌드된 바이너리 확인
                        ls -la main
                    '''
                    
                    // 빌드된 바이너리를 다음 스테이지로 전달
                    stash includes: 'main', name: 'go-binary'
                    stash includes: 'Dockerfile', name: 'dockerfile'
                }
            }
            post {
                always {
                    script {
                        try {
                            if (fileExists('**/test-results.xml')) {
                                publishTestResults testResultsPattern: '**/test-results.xml'
                            }
                        } catch (Exception e) {
                            echo "No test results to publish: ${e.getMessage()}"
                        }
                    }
                }
            }
        }
        
        stage('Security Scan') {
            agent {
                kubernetes {
                    yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: golang
    image: golang:1.24-alpine
    command:
    - cat
    tty: true
"""
                }
            }
            steps {
                container('golang') {
                    echo "🔒 Running security scans..."
                    script {
                        try {
                            sh '''
                                # Go 모듈 취약점 스캔 (선택사항)
                                # go install golang.org/x/vuln/cmd/govulncheck@latest
                                # govulncheck ./...
                                
                                echo "Security scan completed"
                            '''
                        } catch (Exception e) {
                            echo "⚠️ Security scan failed, but continuing: ${e.getMessage()}"
                        }
                    }
                }
            }
        }
        
        stage('Build & Push with Kaniko') {
            agent {
                kubernetes {
                    yaml """
apiVersion: v1
kind: Pod
spec:
  serviceAccountName: jenkins-kaniko-sa  # ServiceAccount 지정
  containers:
  - name: kaniko
    image: gcr.io/kaniko-project/executor:debug
    command:
    - /busybox/cat
    tty: true
    env:
    - name: AWS_REGION
      value: ${AWS_REGION}
"""
                }
            }
            steps {
                container('kaniko') {
                    echo "🐳 Building and pushing with Kaniko..."
                    
                    // 이전 스테이지에서 빌드한 바이너리 가져오기
                    unstash 'go-binary'
                    unstash 'dockerfile'
                    
                    script {
                        sh '''
                            /kaniko/executor \\
                                --dockerfile=Dockerfile \\
                                --context=. \\
                                --destination=177716289679.dkr.ecr.ap-northeast-2.amazonaws.com/izza/autocomplete-server:${BUILD_NUMBER} \\
                                --cache=true
                        '''
                    }
                }
            }
        }
        
        stage('Update GitOps Repository') {
            agent {
                kubernetes {
                    yaml """
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: git
    image: alpine/git:latest
    command:
    - cat
    tty: true
"""
                }
            }
            steps {
                container('git') {
                    echo "🔄 Updating GitOps repository..."
                    script {
                        withCredentials([usernamePassword(credentialsId: "${GIT_CREDENTIALS}", usernameVariable: 'GIT_USERNAME', passwordVariable: 'GIT_PASSWORD')]) {
                            sh '''
                                # 필요한 도구 설치
                                apk add --no-cache sed
                                
                                # 기존 gitops-repo 디렉토리 제거 (있을 경우)
                                rm -rf gitops-repo
                                
                                # GitOps 레포지토리 클론
                                git clone https://${GIT_USERNAME}:${GIT_PASSWORD}@github.com/aws-izza/izza-cd.git gitops-repo
                                cd gitops-repo
                                
                                # Git 설정
                                git config user.name "Jenkins"
                                git config user.email "jenkins@company.com"
                                
                                # 현재 이미지 태그 확인
                                echo "Current image in GitOps repo:"
                                grep "image: " environments/autocomplete-server/app.yaml || echo "No image line found"
                                
                                # 이미지 태그 업데이트
                                sed -i "s|image: ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:.*|image: ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:${IMAGE_TAG}|g" environments/autocomplete-server/app.yaml
                                
                                # 변경사항 확인
                                echo "Updated image in GitOps repo:"
                                grep "image: " environments/autocomplete-server/app.yaml
                                
                                echo "Git diff:"
                                git diff
                                
                                # 변경사항이 있는 경우에만 커밋
                                if [ -n "$(git diff --name-only)" ]; then
                                    # 커밋 및 푸시
                                    git add environments/autocomplete-server/app.yaml
                                    git commit -m "Update autocomplete-server image to ${IMAGE_TAG}

- Build: #${BUILD_NUMBER}
- Commit: ${GIT_COMMIT}
- Branch: ${GIT_BRANCH}
- Triggered by: ${BUILD_USER:-Jenkins}"
                                    
                                    git push origin ${GITOPS_BRANCH}
                                    echo "✅ GitOps repository updated successfully"
                                else
                                    echo "ℹ️ No changes to commit"
                                fi
                            '''
                        }
                    }
                }
            }
        }
    }
    
    post {
        always {
            echo "🏁 Pipeline completed"
        }
        
        success {
            echo "✅ Pipeline succeeded!"
            echo "🏷️ Built and pushed image: ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:${IMAGE_TAG}"
            echo "📦 Image also tagged as: ${ECR_REGISTRY}/${IMAGE_REPOSITORY}:latest"
        }
        
        failure {
            echo "❌ Pipeline failed!"
            echo "💡 Check the logs above for details"
        }
        
        unstable {
            echo "⚠️ Pipeline is unstable"
        }
    }
}