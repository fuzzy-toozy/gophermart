package config

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/fuzzy-toozy/gophermart/internal/database"
	_ "github.com/jackc/pgx/stdlib"
)

type Config struct {
	ServerAddress     string
	StoreFilePath     string
	AccrualAddress    string
	SecretKey         []byte
	TokenLifetime     time.Duration
	DatabaseConfig    database.DBConfig
	MaxBodySize       uint64
	ReadTimeoutSec    time.Duration
	WriteTimeoutSec   time.Duration
	IdleTimeoutSec    time.Duration
	ProcessingInteval time.Duration
}

func BuildConfig() (*Config, error) {
	c := Config{}

	c.ReadTimeoutSec = 10 * time.Second
	c.WriteTimeoutSec = 10 * time.Second
	c.IdleTimeoutSec = 10 * time.Second
	c.TokenLifetime = 48 * time.Hour
	c.ProcessingInteval = 5 * time.Second
	c.DatabaseConfig.DriverName = "pgx"

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var secretKey string

	flag.StringVar(&c.DatabaseConfig.ConnURI, "d", "postgres://postgres:xxXX123@localhost:5432/postgres?sslmode=disable", "Database connection string")
	flag.StringVar(&secretKey, "k", "super_secret_key", "Secret key")
	flag.StringVar(&c.ServerAddress, "a", "localhost:8080", "Server address")
	flag.StringVar(&c.AccrualAddress, "r", "localhost:8080", "Accrual system address")

	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	if len(secretKey) != 0 {
		c.SecretKey = []byte(secretKey)
	}

	err = c.parseEnvVariables()
	if err != nil {
		return nil, err
	}

	c.AccrualAddress += "/api/orders/"

	return &c, err
}
func (c *Config) parseEnvVariables() error {
	type EnvConfig struct {
		ServerAddress  string `env:"RUN_ADDRESS"`
		AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
		DBConnURI      string `env:"DATABASE_URI"`
		SecretKey      string `env:"KEY"`
	}
	ecfg := EnvConfig{}
	err := env.Parse(&ecfg)
	if err != nil {
		return err
	}

	if len(ecfg.ServerAddress) > 0 {
		c.ServerAddress = ecfg.ServerAddress
	}

	if len(ecfg.AccrualAddress) > 0 {
		c.ServerAddress = ecfg.AccrualAddress
	}

	if len(ecfg.DBConnURI) > 0 {
		c.DatabaseConfig.ConnURI = ecfg.DBConnURI
	}

	if len(ecfg.SecretKey) > 0 {
		c.SecretKey = []byte(ecfg.SecretKey)
	}

	return nil
}
