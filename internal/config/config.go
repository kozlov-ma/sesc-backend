package config

import (
	"errors"
	"fmt"
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

type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeSQLite   DatabaseType = "sqlite3"
)

type CompanyDataSource string

const (
	CompanyDataSourceLDAP CompanyDataSource = "ldap"
	CompanyDataSourceDemo CompanyDataSource = "demo"
)

type Config struct {
	Database          DatabaseConfig          `mapstructure:"database"`
	AdminCredentials  []AdminCredentialConfig `mapstructure:"admin_credentials"`
	HTTP              HTTPConfig              `mapstructure:"http"`
	JWTSecret         string                  `mapstructure:"jwt_secret"`
	MinIO             MinIOConfig             `mapstructure:"minio"`
	LDAP              LDAPConfig              `mapstructure:"ldap"`
	CompanyDataSource CompanyDataSource       `mapstructure:"company_data_source"`
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

type LDAPConfig struct {
	URL          string        `mapstructure:"url"`
	BindDN       string        `mapstructure:"bind_dn"`
	BindPassword string        `mapstructure:"bind_password"`
	BaseDN       string        `mapstructure:"base_dn"`
	SyncInterval time.Duration `mapstructure:"sync_interval"`
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

	_ = v.BindEnv("database.address")
	_ = v.BindEnv("jwt_secret")
	_ = v.BindEnv("minio.endpoint")
	_ = v.BindEnv("minio.access_key")
	_ = v.BindEnv("minio.secret_key")
	_ = v.BindEnv("minio.use_ssl")
	_ = v.BindEnv("minio.bucket_name")
	_ = v.BindEnv("minio.base_url")
	_ = v.BindEnv("http.server_address")
	_ = v.BindEnv("http.read_header_timeout")
	_ = v.BindEnv("http.read_timeout")
	_ = v.BindEnv("http.write_timeout")
	_ = v.BindEnv("ldap.url")
	_ = v.BindEnv("ldap.bind_dn")
	_ = v.BindEnv("ldap.bind_password")
	_ = v.BindEnv("ldap.base_dn")
	_ = v.BindEnv("ldap.sync_interval")
	_ = v.BindEnv("company_data_source")

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
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

	v.SetDefault("database.type", string(DatabaseTypePostgres))
	v.SetDefault("database.address", "postgres://postgres:postgres@localhost:5432/sesc?sslmode=disable")

	v.SetDefault("minio.endpoint", "localhost:9000")
	v.SetDefault("minio.access_key", "minioadmin")
	v.SetDefault("minio.secret_key", "minioadmin")
	v.SetDefault("minio.use_ssl", false)
	v.SetDefault("minio.bucket_name", "sesc-files")
	v.SetDefault("minio.base_url", "http://minio:9000/sesc-files")

	v.SetDefault("ldap.url", "ldap://localhost:389")
	v.SetDefault("ldap.bind_dn", "cn=admin,dc=sesc,dc=local")
	v.SetDefault("ldap.bind_password", "admin")
	v.SetDefault("ldap.base_dn", "dc=sesc,dc=local")
	v.SetDefault("ldap.sync_interval", 5*time.Minute)

	v.SetDefault("company_data_source", string(CompanyDataSourceLDAP))

	v.SetDefault("admin_credentials", []AdminCredentialConfig{
		{
			ID:       "f1157f63-65dc-4c3d-bcb2-4d6d55d2e3fd",
			Username: "admin",
			Password: "admin",
		},
	})
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
