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

type Config struct {
	Postgres         PostgresConfig          `mapstructure:"postgres"`
	AdminCredentials []AdminCredentialConfig `mapstructure:"admin_credentials"`
	HTTP             HTTPConfig              `mapstructure:"http"`
	JWTSecret        string                  `mapstructure:"jwt_secret"`
}

type PostgresConfig struct {
	Address string `mapstructure:"address"`
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

func LoadConfig() (*Config, error) {
	v := viper.New()

	setDefaults(v)

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	v.AutomaticEnv()

	v.SetEnvPrefix("SESC")
	replacer := strings.NewReplacer(".", "_")
	v.SetEnvKeyReplacer(replacer)

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
		id, err := uuid.FromString(credential.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID for admin credential: %w", err)
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
