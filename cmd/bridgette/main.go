package main

import (
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := struct {
		l1ExecutionURL string
		l2ExecutionURL string
		dbPath         string
		addr           string
	}{}

	app := &cli.App{
		Name:  "bridgette",
		Usage: "A tool for monitoring of the Optimism Bridge",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "l1-execution-url",
				Usage:       "The URL of the L1 execution layer",
				EnvVars:     []string{"L1_EXECUTION_URL"},
				Required:    true,
				Destination: &cfg.l1ExecutionURL,
			},
			&cli.StringFlag{
				Name:        "l2-execution-url",
				Usage:       "The URL of the L2 execution layer",
				EnvVars:     []string{"L2_EXECUTION_URL"},
				Required:    true,
				Destination: &cfg.l2ExecutionURL,
			},
			&cli.StringFlag{
				Name:        "db-path",
				Usage:       "The path to the database",
				EnvVars:     []string{"DB_PATH"},
				Required:    true,
				Destination: &cfg.dbPath,
			},
			&cli.StringFlag{
				Name:     "addr",
				Usage:    "The address to listen on",
				EnvVars:  []string{"ADDR"},
				Value:    ":8084",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			log.Info("Starting bridge monitor", "l1-execution-url", cfg.l1ExecutionURL, "l2-execution-url", cfg.l2ExecutionURL, "db-path", cfg.dbPath)
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error("Error running bridge monitor", "error", err)
		os.Exit(1)
	}
}
