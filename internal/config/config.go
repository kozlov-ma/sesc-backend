package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/kozlov-ma/sesc-backend/iam"
	"github.com/spf13/viper"
)

const (
	DefaultReadHeaderTimeout = 300 * time.Millisecond
	DefaultReadTimeout       = 3 * time.Second
	DefaultWriteTimeout      = 10 * time.Second
)

// DatabaseType represents the type of database to use
type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeSQLite   DatabaseType = "sqlite3"
)

type Config struct {
	Database         DatabaseConfig          `mapstructure:"database"`
	AdminCredentials []AdminCredentialConfig `mapstructure:"admin_credentials"`
	HTTP             HTTPConfig              `mapstructure:"http"`
	JWTSecret        string                  `mapstructure:"jwt_secret"`
	MinIO            MinIOConfig             `mapstructure:"minio"`
	DeletionDaemon   DeletionDaemonConfig    `mapstructure:"deletion_daemon"`
}

type DatabaseConfig struct {
	Type    DatabaseType `mapstructure:"type"`
	Address string       `mapstructure:"address"`
}

type AdminCredentialConfig struct {
	ID       string `mapstructure:"id"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type HTTPConfig struct {
	ServerAddress     string        `mapstructure:"server_address"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
}

type MinIOConfig struct {
	Endpoint   string `mapstructure:"endpoint"`
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	UseSSL     bool   `mapstructure:"use_ssl"`
	BucketName string `mapstructure:"bucket_name"`
	BaseURL    string `mapstructure:"base_url"`
}

type DeletionDaemonConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Interval time.Duration `mapstructure:"interval"`
	Delay    time.Duration `mapstructure:"delay"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	setDefaults(v)

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.SetEnvPrefix("SESC")
	replacer := strings.NewReplacer(".", "_")
	v.SetEnvKeyReplacer(replacer)
	v.AutomaticEnv()

	// Explicit bindings for all config fields
	// This ensures env variables take precedence over config.yml
	configKeys := []string{
		"database.address",
		"jwt_secret",
		"minio.endpoint",
		"minio.access_key",
		"minio.secret_key",
		"minio.use_ssl",
		"minio.bucket_name",
		"minio.base_url",
		"http.server_address",
		"http.read_header_timeout",
		"http.read_timeout",
		"http.write_timeout",
	}

	var bindErrors []error
	for _, key := range configKeys {
		if err := v.BindEnv(key); err != nil {
			bindErrors = append(bindErrors, fmt.Errorf("failed to bind %s: %w", key, err))
		}
	}

	if len(bindErrors) > 0 {
		return nil, fmt.Errorf("error binding environment variables: %w", errors.Join(bindErrors...))
	}

	// Admin credentials can be set via env variables
	// Format: SESC_ADMIN_CREDENTIALS_0_ID, SESC_ADMIN_CREDENTIALS_0_USERNAME, etc.
	// Only bind if the env vars are actually set to avoid overriding with empty values
	if os.Getenv("SESC_ADMIN_CREDENTIALS_0_ID") != "" {
		_ = v.BindEnv("admin_credentials.0.id")
	}
	if os.Getenv("SESC_ADMIN_CREDENTIALS_0_USERNAME") != "" {
		_ = v.BindEnv("admin_credentials.0.username")
	}
	if os.Getenv("SESC_ADMIN_CREDENTIALS_0_PASSWORD") != "" {
		_ = v.BindEnv("admin_credentials.0.password")
	}

	// Read config.yml if it exists (for local development only)
	// If file doesn't exist, we'll use env variables and defaults
	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK - we'll use env vars and defaults
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("http.server_address", ":8080")
	v.SetDefault("http.read_header_timeout", DefaultReadHeaderTimeout)
	v.SetDefault("http.read_timeout", DefaultReadTimeout)
	v.SetDefault("http.write_timeout", DefaultWriteTimeout)

	v.SetDefault("jwt_secret", "default_secret_change_me_in_production")

	// Default database configuration
	v.SetDefault("database.type", string(DatabaseTypePostgres))
	v.SetDefault("database.address", "postgres://postgres:postgres@localhost:5432/sesc?sslmode=disable")

	// Default MinIO configuration
	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.access_key", "minioadmin")
	v.SetDefault("minio.secret_key", "minioadmin")
	v.SetDefault("minio.use_ssl", false)
	v.SetDefault("minio.bucket_name", "sesc-files")
	v.SetDefault("minio.base_url", "http://localhost:9000/sesc-files")

	v.SetDefault("admin_credentials", []AdminCredentialConfig{
		{
			ID:       "f1157f63-65dc-4c3d-bcb2-4d6d55d2e3fd",
			Username: "admin",
			Password: "admin",
		},
	})

	v.SetDefault("deletion_daemon.enabled", true)
	v.SetDefault("deletion_daemon.interval", "24h")
	v.SetDefault("deletion_daemon.delay", "24h")
}

func (c *Config) ToIAMAdminCredentials() ([]iam.AdminCredentials, error) {
	result := make([]iam.AdminCredentials, len(c.AdminCredentials))

	for i, credential := range c.AdminCredentials {
		if credential.ID == "" {
			return nil, fmt.Errorf("empty UUID for admin credential at index %d", i)
		}

		id, err := uuid.FromString(credential.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID for admin credential at index %d: %w", i, err)
		}

		result[i] = iam.AdminCredentials{
			ID: id,
			Credentials: iam.Credentials{
				Username: credential.Username,
				Password: credential.Password,
			},
		}
	}

	return result, nil
}
