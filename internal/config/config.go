package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Logging     LoggingConfig
	Cors        CorsConfig
	Postgres    PostgresConfig
	Redis       RedisConfig
	OAuth2      OAuth2Config
}

type ServerConfig struct {
	Port         int
	Timeout      time.Duration
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type CorsConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type LoggingConfig struct {
	ConsoleEnabled bool
	ConsoleLevel   string
	FileEnabled    bool
	FileLevel      string
	Format         string
	File           string
	MaxSize        int // megabytes
	MaxBackups     int
	MaxAge         int // days
	Compress       bool
}

type PostgresConfig struct {
	Host                   string
	Port                   int
	Database               string
	Schema                 string
	User                   string
	Password               string
	DefaultMaxOpenConns    int
	DefaultMaxIdleConns    int
	DefaultConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host         string
	Port         int
	Database     int
	Password     string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
	MinIdleConns int
}

type OAuth2Config struct {
	Enabled              bool   `mapstructure:"enabled"`
	ServiceEnabled       bool   `mapstructure:"service_enabled"`
	IntrospectionEnabled bool   `mapstructure:"introspection_enabled"`
	ClientID             string `mapstructure:"client_id"`
	ClientSecret         string `mapstructure:"client_secret"`
	JWTSecret            string `mapstructure:"jwt_secret"`
	BaseAuthURL          string `mapstructure:"baseauthurl"`
	GetTokenPath         string `mapstructure:"gettokenpath"`
	RevokeTokenPath      string `mapstructure:"revoketokenpath"`
	IntrospectionPath    string `mapstructure:"introspectionpath"`
}

const (
	fatalConfigErr       = "fatal error config file: %w"
	defaultPostgresPort  = 5432
	defaultRedisPort     = 6379
	defaultRedisDatabase = 0
)

var Instance *Config

func Load() *Config {
	// Environment variables
	// env variables will look like USERMGMT_SERVER_PORT, USERMGMT_LOGGING_LEVEL
	viper.SetEnvPrefix("USERMGMT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read config files
	viper.AddConfigPath("./config")

	// Load config
	loadServerConfig()
	mergeDatabaseConfig()
	mergeOauth2Config()
	loadCorsConfig()
	loadLoggingConfig()
	loadEnvironmentConfig()
	loadPostgresConfig()
	loadRedisConfig()
	loadOauth2Config()

	var cfg Config

	err := viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}

	Instance = &cfg
	validateConfig(Instance)

	return Instance
}

func validateConfig(cfg *Config) {
	if cfg.OAuth2.Enabled {
		if cfg.OAuth2.ClientID == "" {
			panic("oauth2.client_id is required when oauth2 is enabled")
		}

		if cfg.OAuth2.ClientSecret == "" {
			panic("oauth2.client_secret is required when oauth2 is enabled")
		}

		if cfg.OAuth2.BaseAuthURL == "" {
			panic("oauth2.base_auth_url is required when oauth2 is enabled")
		}

		if cfg.OAuth2.GetTokenPath == "" {
			panic("oauth2.get_token_path is required when oauth2 is enabled")
		}
	}
}

func mergeDatabaseConfig() {
	viper.SetConfigName("database")
	viper.SetConfigType("yaml")

	err := viper.MergeInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			panic("database config file not found")
		} else {
			panic(fmt.Errorf(fatalConfigErr, err))
		}
	}
}

func mergeOauth2Config() {
	viper.SetConfigName("oauth2")
	viper.SetConfigType("yaml")

	err := viper.MergeInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Config file not found; ignore
			return
		} else {
			panic(fmt.Errorf(fatalConfigErr, err))
		}
	}
}

func loadCorsConfig() {
	viper.SetConfigName("cors")
	viper.SetConfigType("yaml")
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("cors config file not found")
		} else {
			panic(fmt.Errorf(fatalConfigErr, err))
		}
	}
}

func loadEnvironmentConfig() {
	viper.SetDefault("environment", "development")

	_ = viper.BindEnv("environment", "ENVIRONMENT")
}

func loadLoggingConfig() {
	viper.SetConfigName("logging")
	viper.SetConfigType("yaml")

	err := viper.MergeInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			panic("logging config file not found")
		} else {
			panic(fmt.Errorf(fatalConfigErr, err))
		}
	}
}

func loadOauth2Config() {
	viper.SetDefault("oauth2.enabled", false)
	viper.SetDefault("oauth2.service_enabled", false)
	viper.SetDefault("oauth2.introspection_enabled", false)

	_ = viper.BindEnv("oauth2.enabled", "OAUTH2_ENABLED")
	_ = viper.BindEnv("oauth2.service_enabled", "OAUTH2_SERVICE_ENABLED")
	_ = viper.BindEnv("oauth2.introspection_enabled", "OAUTH2_INTROSPECTION_ENABLED")
	_ = viper.BindEnv("oauth2.client_id", "OAUTH2_CLIENT_ID")
	_ = viper.BindEnv("oauth2.client_secret", "OAUTH2_CLIENT_SECRET")
	_ = viper.BindEnv("oauth2.jwt_secret", "OAUTH2_JWT_SECRET")
	_ = viper.BindEnv("oauth2.authBaseUrl", "OAUTH2_AUTH_BASE_URL")
	_ = viper.BindEnv("oauth2.getTokenPath", "OAUTH2_GET_TOKEN_PATH")
	_ = viper.BindEnv("oauth2.revokeTokenPath", "OAUTH2_REVOKE_TOKEN_PATH")
	_ = viper.BindEnv("oauth2.introspectionPath", "OAUTH2_INTROSPECTION_PATH")
}

func loadPostgresConfig() {
	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", defaultPostgresPort)
	viper.SetDefault("postgres.database", "postgres")
	viper.SetDefault("postgres.schema", "public")
	viper.SetDefault("postgres.user", "postgres")

	_ = viper.BindEnv("postgres.host", "POSTGRES_HOST")
	_ = viper.BindEnv("postgres.port", "POSTGRES_PORT")
	_ = viper.BindEnv("postgres.database", "POSTGRES_DB")
	_ = viper.BindEnv("postgres.schema", "POSTGRES_SCHEMA")
	_ = viper.BindEnv("postgres.user", "POSTGRES_USER")
	_ = viper.BindEnv("postgres.password", "POSTGRES_PASSWORD")
}

func loadRedisConfig() {
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", defaultRedisPort)
	viper.SetDefault("redis.database", defaultRedisDatabase)
	viper.SetDefault("redis.password", "")

	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.database", "REDIS_DB")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")
}

func loadServerConfig() {
	viper.SetConfigName("server")
	viper.SetConfigType("yaml")

	err := viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			panic("server config file not found")
		} else {
			panic(fmt.Errorf(fatalConfigErr, err))
		}
	}
}
