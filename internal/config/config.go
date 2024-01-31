package config

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ivas1ly/gophermart/internal/utils/jwt"
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

var (
	defaultSigningKey = []byte("36626d331c8c44f2d72f348f36323743598e267e86b3e4aca27c5b433247ea72")
)

type Config struct {
	App
	HTTP
	DB
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
	jwt.SigningKey = defaultSigningKey

	cfg.CompressLevel = defaultCompressLevel

	fmt.Printf("%+v\n\n", cfg)

	return cfg
}
