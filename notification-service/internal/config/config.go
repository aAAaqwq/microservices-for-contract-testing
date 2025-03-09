package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config 应用配置
type Config struct {
	MongoDB  MongoDBConfig `yaml:"mongodb-notification"`
	Redis    RedisConfig   `yaml:"redis-notification"`
	Port     string        `yaml:"notification-port"`
	Services ServiceConfig `yaml:"services"`
	Email    EmailConfig   `yaml:"email"`
}

// MongoDBConfig MongoDB配置
type MongoDBConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// ServiceConfig 其他服务的配置
type ServiceConfig struct {
	UserServiceURL    string `yaml:"user-service-url"`
	OrderServiceURL   string `yaml:"order-service-url"`
	PaymentServiceURL string `yaml:"payment-service-url"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
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

