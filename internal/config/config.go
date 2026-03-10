package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	PostgresDB `yaml:"postgres_db"`
	HTTPServer `yaml:"http_server"`
	TCPServer  `yaml:"tcp_server"`
}

type PostgresDB struct {
	Host     string `yaml:"host" env-default:"localhost"`
	Port     string `yaml:"port" env-default:"5432"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true" env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"db_name" env-default:"postgres"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type TCPServer struct {
	Address string `yaml:"address" env-default:"localhost:9090"`
}

func MustLoad() *Config {
	var cfg Config

	ConfigPath := os.Getenv("CONFIG_PATH")
	if ConfigPath == "" {
		log.Fatal("CONFIG_PATH environment variable not set")
	}
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		log.Fatal("config does not exist")
	}

	err := cleanenv.ReadConfig(ConfigPath, &cfg)
	if err != nil {
		log.Fatalf("cannot read config %s", err)
	}

	return &cfg
}
