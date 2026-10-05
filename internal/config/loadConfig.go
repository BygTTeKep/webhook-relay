package config

import "github.com/spf13/viper"

type DBConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Name     string `mapstructure:"name"`
	Password string `mapstructure:"password"`
}
type AppConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
}

type Config struct {
	DBCfg  DBConfig
	AppCfg AppConfig
}

// фукнкция загрузки конфига
// path путь до конфига
func LoadConfig(path string) (cfg *Config, err error) {

	viper.AddConfigPath(path)
	viper.SetConfigType("yml")

	err = viper.ReadInConfig()
	if err != nil {
		return &Config{},err
	}
	err = viper.Unmarshal(&cfg)
	return
}