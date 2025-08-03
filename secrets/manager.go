package secrets

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type DBSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SecretsManager struct {
	client *secretsmanager.SecretsManager
	region string
}

func NewSecretsManager(region string) *SecretsManager {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	return &SecretsManager{
		client: secretsmanager.New(sess),
		region: region,
	}
}

func (sm *SecretsManager) GetDatabaseCredentials(secretName string) (string, string, error) {
	log.Printf("Fetching database credentials from AWS Secrets Manager: %s", secretName)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := sm.client.GetSecretValue(input)
	if err != nil {
		return "", "", fmt.Errorf("failed to get secret from AWS Secrets Manager: %w", err)
	}

	var dbSecret DBSecret
	if err := json.Unmarshal([]byte(*result.SecretString), &dbSecret); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal secret JSON: %w", err)
	}

	log.Printf("Successfully retrieved database credentials for user: %s", dbSecret.Username)
	return dbSecret.Username, dbSecret.Password, nil
}
