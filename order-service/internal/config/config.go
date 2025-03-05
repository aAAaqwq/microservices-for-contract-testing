package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config 应用配置
type Config struct {
	DB       DBConfig `yaml:"mysql-order"`
	Port     string   `yaml:"order-port"`
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
	UserServiceURL         string `yaml:"user-service-url"`
	PaymentServiceURL      string `yaml:"payment-service-url"`
	NotificationServiceURL string `yaml:"notification-service-url"`
	OrderServiceURL        string `yaml:"order-service-url"`
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


