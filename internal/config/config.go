package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig `envPrefix:"DB_"`
	Server   ServerConfig   `envPrefix:"SERVER_"`
}

type DatabaseConfig struct {
	Port     string `env:"PORT" envDefault:"5432"`
	Host     string `env:"HOST" envDefault:"localhost"`
	Name     string `env:"NAME" envDefault:"postgres"`
	User     string `env:"USER" envDefault:"postgres"`
	Password string `env:"PASS" envDefault:"postgres"`
	SSLMode  string `env:"SSL" envDefault:"disable"`
}

type ServerConfig struct {
	BaseURL string `env:"BASE_URL" envDefault:"http://localhost:8000"`
	Port string	`env:"PORT" envDefault:"8000"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Printf(".env not found, using system variables")
	}

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) DBString() string {
	return fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?sslmode=%v", c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name, c.Database.SSLMode)
}
