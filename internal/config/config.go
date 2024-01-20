package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	defaultRunHost              = "localhost"
	defaultRunPort              = "8080"
	defaultCompressLevel        = 5
	defaultLogLevel             = "info"
	defaultDatabaseConnTimeout  = 5 * time.Second
	defaultDatabaseConnAttempts = 3
	exampleDatabaseDSN          = "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable"
	exampleAccrualSystemAddress = "http://localhost:3560"
)

type Config struct {
	App
	DB
	HTTP
}

type App struct {
	LogLevel             string
	AccrualSystemAddress string
}

type DB struct {
	DatabaseURI          string
	DatabaseConnTimeout  time.Duration
	DatabaseConnAttempts int
}

type HTTP struct {
	RunAddress    string
	CompressLevel int
}

func New() Config {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	cfg := Config{}

	runAddressUsage := fmt.Sprintf("HTTP server endpoint, example: %q or %q",
		net.JoinHostPort(defaultRunHost, defaultRunPort), net.JoinHostPort("", defaultRunPort))
	flag.StringVar(&cfg.RunAddress, "a", net.JoinHostPort(defaultRunHost, defaultRunPort), runAddressUsage)

	dsnUsage := fmt.Sprintf("Database connection string, example: %q", exampleDatabaseDSN)
	flag.StringVar(&cfg.DatabaseURI, "d", "", dsnUsage)

	accrualSystemAddressUsage := fmt.Sprintf("Accrual system endpoint, example: %q", exampleAccrualSystemAddress)
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", accrualSystemAddressUsage)

	flag.Parse()

	if runAddress := os.Getenv("RUN_ADDRESS"); runAddress != "" {
		cfg.RunAddress = runAddress
	}

	if databaseURI := os.Getenv("DATABASE_URI"); databaseURI != "" {
		cfg.DatabaseURI = databaseURI
	}

	if accrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualSystemAddress != "" {
		cfg.AccrualSystemAddress = accrualSystemAddress
	}

	cfg.DatabaseConnAttempts = defaultDatabaseConnAttempts
	cfg.DatabaseConnTimeout = defaultDatabaseConnTimeout

	cfg.LogLevel = defaultLogLevel

	cfg.CompressLevel = defaultCompressLevel

	fmt.Printf("%+v\n\n", cfg)

	return cfg
}
