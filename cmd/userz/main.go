package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"

	"github.com/leophys/userz"
	"github.com/leophys/userz/http"
	"github.com/leophys/userz/store/pg"
)

const (
	defaultPGHost      = "localhost"
	defaultPGPort      = 5432
	defaultHTTPPort    = 6000
	defaultGRPCPort    = 7000
	defaultMetricsPort = 25000
	defaultHTTPRoute   = "/api"
)

var (
	flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Usage:   "Set logging to debug level (defaults to info)",
			EnvVars: []string{"DEBUG"},
		},
		&cli.BoolFlag{
			Name:  "console",
			Usage: "Enable pretty (and slower) logging",
		},
		&cli.IntFlag{
			Name:    "http-port",
			Usage:   "The port on which the HTTP API will be exposed",
			EnvVars: []string{"HTTP_PORT"},
			Value:   defaultHTTPPort,
			Action:  validatePort,
		},
		&cli.IntFlag{
			Name:    "grpc-port",
			Usage:   "The port on which the gRPC API will be exposed",
			EnvVars: []string{"GRPC_PORT"},
			Value:   defaultGRPCPort,
			Action:  validatePort,
		},
		&cli.IntFlag{
			Name:    "metrics-port",
			Usage:   "The port on which the metrics will be exposed (healthchech and prometheus)",
			EnvVars: []string{"METRICS_PORT"},
			Value:   defaultMetricsPort,
			Action:  validatePort,
		},
		&cli.StringFlag{
			Name:    "pgurl",
			Usage:   "The url to connect to the postgres database (if specified, supercedes all other postgres flags)",
			EnvVars: []string{"POSTGRES_URL"},
		},
		&cli.StringFlag{
			Name:    "pguser",
			Usage:   "The user to connect to the postgres database",
			EnvVars: []string{"POSTGRES_USER"},
		},
		&cli.StringFlag{
			Name:    "pghost",
			Usage:   "The host to connect to the postgres database",
			EnvVars: []string{"POSTGRES_HOST"},
			Value:   defaultPGHost,
		},
		&cli.StringFlag{
			Name:    "pgpassword",
			Usage:   "The password to connect to the postgres database",
			EnvVars: []string{"POSTGRES_PASSWORD"},
		},
		&cli.IntFlag{
			Name:    "pgport",
			Usage:   "The port to connect to the postgres database",
			EnvVars: []string{"POSTGRES_PORT"},
			Value:   defaultPGPort,
			Action:  validatePort,
		},
		&cli.StringFlag{
			Name:    "pgdbname",
			Usage:   "The dbname to connect to the postgres database",
			EnvVars: []string{"POSTGRES_DBNAME"},
		},
		&cli.BoolFlag{
			Name:    "pgssl",
			Usage:   "Wether to connect to the postgres database in strict ssl mode",
			EnvVars: []string{"POSTGRES_SSL"},
		},
	}
)

func init() {
	zerolog.TimeFieldFormat = ""
	zerolog.TimestampFunc = time.Now
}

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt, syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	app := &cli.App{
		Name:   "userz",
		Usage:  "manage a list of users",
		Flags:  flags,
		Action: run,
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		os.Exit(1)
	}
}

func run(c *cli.Context) error {
	logger := setupLogger(c)
	ctx := logger.WithContext(c.Context)

	pgURL, err := getPostgresURL(c)
	if err != nil {
		logger.Err(err).Msg("Failed to get postgres URL")
		return err
	}

	if err := pg.Migrate(pgURL); err != nil {
		logger.Err(err).Msg("Failed to perform migrations")
		return err
	}

	var store userz.Store

	store, err = pg.NewPGStore(ctx, pgURL)
	if err != nil {
		logger.Err(err).Msg("Failed to initialize store")
		return err
	}

	api := httpapi.New(defaultHTTPRoute, store)

	go listenAndServe(ctx, api, c.Int("http-port"))

	<-ctx.Done()

	return nil
}

func validatePort(c *cli.Context, p int) error {
	maxPort := 1<<16 - 2
	if p < 0 || p > maxPort {
		return fmt.Errorf("port must be in the [0, %d] range", maxPort)
	}
	return nil
}

func setupLogger(c *cli.Context) *zerolog.Logger {
	level := zerolog.InfoLevel

	if c.Bool("debug") {
		level = zerolog.DebugLevel
	}

	var w io.Writer = os.Stdout

	if c.Bool("console") {
		w = &zerolog.ConsoleWriter{
			Out: w,
		}
	}

	logger := zerolog.New(w).Level(level)

	logger.Debug().Msg("Debug level set")

	return &logger
}

func getPostgresURL(c *cli.Context) (string, error) {
	if url := c.String("pgurl"); url != "" {
		return url, nil
	}

	user := c.String("pguser")
	if user == "" {
		return "", fmt.Errorf("pguser is mandatory")
	}

	password := c.String("pgpassword")
	if password == "" {
		return "", fmt.Errorf("pgpassword is mandatory")
	}

	dbname := c.String("pgdbname")
	if dbname == "" {
		return "", fmt.Errorf("pgdbname is mandatory")
	}

	sslmode := "disable"
	if c.Bool("pgssl") {
		sslmode = "require"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, c.String("pghost"), c.Int("pgport"), dbname, sslmode,
	), nil
}

func listenAndServe(ctx context.Context, router chi.Router, port int) {
	logger := zerolog.Ctx(ctx)

	addr := fmt.Sprintf(":%d", port)

	logger.Info().Msgf("Serving HTTP APIs on '%s'", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Info().Msg("HTTP API shut down")
		} else {
			logger.Err(err).Msg("HTTP API failed to be served")
		}
	}
}
