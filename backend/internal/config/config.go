package congfig

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name    string `mapstructure:"name"`
		Port    string `mapstructure:"port"`
		Version string `mapstructure:"version"`
	} `mapstructure:"app"`
	PostgreSQL struct{}
	JWT        struct {
		SecretKey string `mapstructure:"secretkey"`
	} `mapstructure:"jwt"`
	Redis    struct{}
	Slog     struct{}
	RabbitMQ struct{}
}

var AppConfig *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/.")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	AppConfig = &Config{}
	err = viper.Unmarshal(AppConfig)
	if err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}
	log.Printf("Config loaded successfully: %+v", AppConfig)
}
