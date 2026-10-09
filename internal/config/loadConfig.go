package config

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type DBConfig struct {
	Dsn string `mapstructure:"dsn" validate:"required"`
}
type AppConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port" validate:"required,min=1,max=65535"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers" validate:"required"`
	Topic   string   `mapstructure:"topic" validate:"required"`
}

type LoggerConfig struct {
	Dev bool `mapstructure:"dev"`
}

type GrpcServersConfig struct {
	StatsServ string `mapstructure:"stats"`
}

type RedisConfig struct {
	Addr string `mapstructure:"addr"`
	Db   int    `mapstructure:"db"`
}

type Config struct {
	DBCfg      DBConfig          `mapstructure:"database" validate:"required"`
	AppCfg     AppConfig         `mapstructure:"app" validate:"required"`
	KafkaCfg   KafkaConfig       `mapstructure:"kafka" validate:"required"`
	LoggerCfg  LoggerConfig      `mapstructure:"logger" validate:"required"`
	GrpcServer GrpcServersConfig `mapstructure:"grpcservers" validate:"required"`
	RedisCfg   RedisConfig       `mapstructure:"redis" validate:"required"`
}

func LoadConfig(path string) (c *Config, err error) {
	v := viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}
