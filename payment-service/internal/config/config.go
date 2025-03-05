package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config 应用配置
type Config struct {
	DB       DBConfig      `yaml:"postgres-payment"`
	Port     string        `yaml:"payment-port"`
	Services ServiceConfig `yaml:"services"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

// ServiceConfig 其他服务的配置
type ServiceConfig struct {
	OrderServiceURL        string `yaml:"order-service-url"`
	NotificationServiceURL string `yaml:"notification-service-url"`
}

// LoadConfig 加载yaml配置
func LoadConfig(path string) (*Config, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
