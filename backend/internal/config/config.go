package congfig

import (
	"os"

	"github.com/spf13/viper"
	"backend/internal/logger"
)

type Config struct {
	App struct {
		Name    string `mapstructure:"name"`
		Port    string `mapstructure:"port"`
		Version string `mapstructure:"version"`
	} `mapstructure:"app"`
	PostgreSQL struct{
		DSN string `mapstructure:"dsn"`
	}`mapstructure:"postgresql"`
	JWT        struct {
		SecretKey string `mapstructure:"secretkey"`
	} `mapstructure:"jwt"`
	Redis    struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
	} `mapstructure:"redis"`
	Slog     struct{} `mapstructure:"slog"`
	RabbitMQ struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
	} `mapstructure:"rabbitmq"`
	Qdrant struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"Qdrant"`
}

var AppConfig *Config

func InitConfig(appLog *logger.AppLogger) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs/.")

	err := viper.ReadInConfig()
	if err != nil {
		appLog.Error(err, "Error reading config file")
		os.Exit(1)
	}

	AppConfig = &Config{}
	err = viper.Unmarshal(AppConfig)
	if err != nil {
		appLog.Error(err, "Error unmarshalling config")
		os.Exit(1)
	}
	appLog.Info("Config loaded successfully", "config", AppConfig)
}
