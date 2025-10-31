package config

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/viper"
)

type Config struct {
	App       *AppConfig
	Tracer    *TracerConfig
	HTTP      *HTTPConfig
	MySQL     *DatabaseConfig
	Token     *TokenConfig
	Pubsub    *PubsubConfig
	Drive     *DriveConfig
	StaleTask *StaleTaskConfig
	Redis     *RedisConfig
	Gmail     *GmailConfig
}

type AppConfig struct {
	Name        string
	Version     string
	Environment string
	Debug       bool
	UsePubsub   bool
	FrontendURL string
}

type TracerConfig struct {
	ServerURL      string
	SecretToken    string
	ServiceName    string
	ServiceVersion string
	Environment    string
	NodeName       string
}

type DatabaseConfig struct {
	DSN                string
	MigrateDSN         string
	DBName             string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetime    int
	SlowQueryThreshold int
	Debug              bool
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type HTTPConfig struct {
	Host               string
	Port               int
	BasePath           string
	DomainName         string
	EnableMigrationAPI bool
}

type TokenConfig struct {
	AccessSecretKey      string
	AccessTokenDuration  int // in minutes
	RefreshSecretKey     string
	RefreshTokenDuration int // in minutes
}

type PubsubConfig struct {
	ProjectID string
	TopicID   string
	CredFile  string
}

type DriveConfig struct {
	IconFolderID string
}

type GmailConfig struct {
	CredFile string
	Sender   string
}

type StaleTaskConfig struct {
	MaxStaleTime  int
	CheckInterval int
}

func LoadConfig(envPath string) (*Config, error) {
	if envPath == "" {
		envPath = ".env"
	}

	viper.SetConfigFile(envPath)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		var cfgErr viper.ConfigFileNotFoundError
		if !errors.As(err, &cfgErr) {
			return nil, err
		}
	}

	config := &Config{
		App: &AppConfig{
			Name:        viper.GetString("APP_NAME"),
			Version:     viper.GetString("APP_VERSION"),
			Environment: viper.GetString("APP_ENVIRONMENT"),
			Debug:       viper.GetBool("APP_DEBUG"),
			UsePubsub:   viper.GetBool("APP_USE_PUBSUB"),
			FrontendURL: viper.GetString("FRONTEND_URL"),
		},
		Tracer: &TracerConfig{
			ServerURL:      viper.GetString("ELASTIC_APM_SERVER_URL"),
			SecretToken:    viper.GetString("ELASTIC_APM_SECRET_TOKEN"),
			ServiceName:    viper.GetString("ELASTIC_APM_SERVICE_NAME"),
			ServiceVersion: viper.GetString("ELASTIC_APM_SERVICE_VERSION"),
			Environment:    viper.GetString("ELASTIC_APM_ENVIRONMENT"),
			NodeName:       viper.GetString("ELASTIC_APM_NODE_NAME"),
		},
		HTTP: &HTTPConfig{
			Host:               viper.GetString("HTTP_HOST"),
			Port:               viper.GetInt("HTTP_PORT"),
			BasePath:           viper.GetString("HTTP_BASE_PATH"),
			DomainName:         viper.GetString("HTTP_DOMAIN_NAME"),
			EnableMigrationAPI: viper.GetBool("HTTP_ENABLE_MIGRATION_API"),
		},
		MySQL: &DatabaseConfig{
			DSN:                viper.GetString("MYSQL_DSN"),
			MigrateDSN:         viper.GetString("MYSQL_MIGRATE_DSN"),
			DBName:             viper.GetString("MYSQL_DB_NAME"),
			MaxOpenConns:       viper.GetInt("MYSQL_MAX_OPEN_CONNS"),
			MaxIdleConns:       viper.GetInt("MYSQL_MAX_IDLE_CONNS"),
			ConnMaxLifetime:    viper.GetInt("MYSQL_CONN_MAX_LIFETIME"),
			SlowQueryThreshold: viper.GetInt("MYSQL_SLOW_QUERY_THRESHOLD"),
			Debug:              viper.GetBool("MYSQL_DEBUG"),
		},
		Redis: &RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetString("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		Token: &TokenConfig{
			AccessSecretKey:      viper.GetString("ACCESS_TOKEN_SECRET_KEY"),
			AccessTokenDuration:  viper.GetInt("ACCESS_TOKEN_DURATION"),
			RefreshSecretKey:     viper.GetString("REFRESH_TOKEN_SECRET_KEY"),
			RefreshTokenDuration: viper.GetInt("REFRESH_TOKEN_DURATION"),
		},
		Pubsub: &PubsubConfig{
			ProjectID: viper.GetString("PUBSUB_PROJECT_ID"),
			TopicID:   viper.GetString("PUBSUB_TOPIC_ID"),
			CredFile:  viper.GetString("PUBSUB_CRED_FILE"),
		},
		Drive: &DriveConfig{
			IconFolderID: viper.GetString("DRIVE_ICON_FOLDER_ID"),
		},
		StaleTask: &StaleTaskConfig{
			MaxStaleTime:  viper.GetInt("STALE_TASK_MAX_STALE_TIME"),
			CheckInterval: viper.GetInt("STALE_TASK_CHECK_INTERVAL"),
		},
		Gmail: &GmailConfig{
			CredFile: viper.GetString("GMAIL_CRED_FILE"),
			Sender:   viper.GetString("GMAIL_SENDER"),
		},
	}

	return config, nil
}
