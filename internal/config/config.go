package config

import (
	"flag"
	"fmt"
	"log"
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
	defaultReadTimeout          = 10 * time.Second
	defaultReadHeaderTimeout    = 5 * time.Second
	defaultWriteTimeout         = 10 * time.Second
	defaultIdleTimeout          = 1 * time.Minute
	defaultShutdownTimeout      = 5 * time.Second
	defaultWorkerPollInterval   = 10 * time.Second
	defaultClientTimeout        = 5 * time.Second
)

var (
	defaultSigningKey = []byte("36626d331c8c44f2d72f348f36323743598e267e86b3e4aca27c5b433247ea72")
)

type Config struct {
	DB
	App
	HTTP
}

type App struct {
	LogLevel             string
	AccrualSystemAddress string
	SigningKey           []byte
	WorkerPollInterval   time.Duration
}

type DB struct {
	DatabaseURI          string
	DatabaseConnTimeout  time.Duration
	DatabaseConnAttempts int
}

type HTTP struct {
	RunAddress        string
	CompressLevel     int
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	ClientTimeout     time.Duration
}

func New() Config {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	cfg := Config{
		App: App{
			LogLevel:           defaultLogLevel,
			SigningKey:         defaultSigningKey,
			WorkerPollInterval: defaultWorkerPollInterval,
		},
		HTTP: HTTP{
			CompressLevel:     defaultCompressLevel,
			ReadTimeout:       defaultReadTimeout,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
			ShutdownTimeout:   defaultShutdownTimeout,
			ClientTimeout:     defaultClientTimeout,
		},
		DB: DB{
			DatabaseConnTimeout:  defaultDatabaseConnTimeout,
			DatabaseConnAttempts: defaultDatabaseConnAttempts,
		},
	}

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

	log.Println("loaded config:", fmt.Sprintf("%+v", cfg))

	return cfg
}
