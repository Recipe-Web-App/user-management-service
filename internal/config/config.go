package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig
	Logging LoggingConfig
	Cors    CorsConfig
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

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	Instance = &cfg
	return Instance
}
