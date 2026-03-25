package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	S3       S3Config       `mapstructure:"s3"`
	Logger   Logger         `mapstructure:"logger"`
	Telegram TelegramConfig `mapstructure:"telegram"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

type ServerConfig struct {
	Address string `mapstructure:"address"`
	Env     string `mapstructure:"env"`
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type JWTConfig struct {
	Secret       string `mapstructure:"secret"`
	ExpiresHours int    `mapstructure:"expires_hours"`
}

type Logger struct {
	Level string `mapstructure:"level"`
}

type S3Config struct {
	Bucket    string `mapstructure:"bucket"`
	Endpoint  string `mapstructure:"endpoint"`
	CDNBase   string `mapstructure:"cdn_base"`
	CDNToken  string `mapstructure:"cdn_token"`
	Region    string `mapstructure:"region"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
}

type TelegramConfig struct {
	BotToken    string `mapstructure:"bot_token"`
	AdminChatID string `mapstructure:"admin_chat_id"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("../")

	// Allow ENV overrides: DATABASE_URL → database.url
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind explicit env mappings
	_ = viper.BindEnv("database.url", "DATABASE_URL")
	_ = viper.BindEnv("logger.level", "LOG_LEVEL")
	_ = viper.BindEnv("jwt.secret", "JWT_SECRET")
	_ = viper.BindEnv("s3.access_key", "S3_ACCESS_KEY")
	_ = viper.BindEnv("s3.secret_key", "S3_SECRET_KEY")
	_ = viper.BindEnv("s3.bucket", "S3_BUCKET")
	_ = viper.BindEnv("s3.endpoint", "S3_ENDPOINT")
	_ = viper.BindEnv("s3.cdn_base", "S3_CDN_BASE")
	_ = viper.BindEnv("s3.cdn_token", "S3_CDN_TOKEN")
	_ = viper.BindEnv("telegram.bot_token", "TG_BOT_TOKEN")
	_ = viper.BindEnv("telegram.admin_chat_id", "TG_ADMIN_CHAT_ID")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
