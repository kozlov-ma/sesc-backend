//nolint:mnd // these aint magic numbers these are config values
package testutil

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/kozlov-ma/sesc-backend/internal/config"
)

func CreateTestConfig() *config.Config {
	return &config.Config{
		Database: config.DatabaseConfig{
			Type:    config.DatabaseTypeSQLite,
			Address: "file:ent?mode=memory&cache=shared&_fk=1",
		},
		HTTP: config.HTTPConfig{
			ServerAddress:     ":0", // Use port 0 to let the OS assign a free port
			ReadHeaderTimeout: 100 * time.Millisecond,
			ReadTimeout:       1 * time.Second,
			WriteTimeout:      1 * time.Second,
		},
		JWTSecret: "test_secret",
		AdminCredentials: []config.AdminCredentialConfig{
			{
				ID:       "f1157f63-65dc-4c3d-bcb2-4d6d55d2e3fd",
				Username: "admin",
				Password: "admin",
			},
		},
	}
}

func GetProjectRoot() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", errors.New("could not find project root (no go.mod found)")
		}
		currentDir = parentDir
	}
}
