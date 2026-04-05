package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	JWT        `yaml:"jwt"`
	PostgresDB `yaml:"postgres_db"`
	HTTPServer `yaml:"http_server"`
	TCPServer  `yaml:"tcp_server"`
}

type JWT struct {
	Secret   string        `yaml:"secret" env-required:"true" env:"JWT_SECRET"`
	ExpiryIn time.Duration `yaml:"expiry_in" env-all:"1800s"`
}

type PostgresDB struct {
	Host     string `yaml:"host" env-all:"localhost"`
	Port     string `yaml:"port" env-all:"5432"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true" env:"POSTGRES_PASSWORD"`
	DBName   string `yaml:"db_name" env-all:"postgres"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-all:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-all:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-all:"60s"`
}

type TCPServer struct {
	Address     string        `yaml:"address" env-all:"localhost:9090"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-all:"60s"`
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
