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

const fatalConfigErr = "fatal error config file: %w"

var Instance *Config

func Load() *Config {
	// Environment variables
	// env variables will look like USERMGMT_SERVER_PORT, USERMGMT_LOGGING_LEVEL
	viper.SetEnvPrefix("USERMGMT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Read config files
	viper.AddConfigPath("./config")

	// Load server config
	viper.SetConfigName("server")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("server config file not found")
		} else {
			panic(fmt.Errorf(fatalConfigErr, err))
		}
	}

	// Load cors config
	viper.SetConfigName("cors")
	viper.SetConfigType("yaml")
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("cors config file not found")
		} else {
			panic(fmt.Errorf(fatalConfigErr, err))
		}
	}

	// Merge logging config
	viper.SetConfigName("logging")
	viper.SetConfigType("yaml")
	if err := viper.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic("logging config file not found")
		} else {
			panic(fmt.Errorf(fatalConfigErr, err))
		}
	}

	// Merge database config
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

	// Postgres defaults and env binding
	_ = viper.BindEnv("postgres.host", "POSTGRES_HOST")
	_ = viper.BindEnv("postgres.port", "POSTGRES_PORT")
	_ = viper.BindEnv("postgres.database", "POSTGRES_DB")
	_ = viper.BindEnv("postgres.schema", "POSTGRES_SCHEMA")
	_ = viper.BindEnv("postgres.user", "POSTGRES_USER")
	_ = viper.BindEnv("postgres.password", "POSTGRES_PASSWORD")

	// Redis defaults and env binding
	_ = viper.BindEnv("redis.host", "REDIS_HOST")
	_ = viper.BindEnv("redis.port", "REDIS_PORT")
	_ = viper.BindEnv("redis.database", "REDIS_DB")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	Instance = &cfg
	return Instance
}
