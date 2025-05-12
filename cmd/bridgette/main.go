package main

import (
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
)

// ETHBridgeInitiated (index_topic_1 address from, index_topic_2 address to, uint256 amount, bytes extraData)
// 0x2849b43074093a05396b6f2a937dee8565b15a48a7b3d4bffb732a5017380af5

var ethBridgeInitiatedEvent = common.HexToHash("0x2849b43074093a05396b6f2a937dee8565b15a48a7b3d4bffb732a5017380af5")

func main() {

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := struct {
		l1ExecutionURL  string
		l2ExecutionURL  string
		dbPath          string
		addr            string
		l1BridgeAddress string
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
				Name:        "addr",
				Usage:       "The address to listen on",
				EnvVars:     []string{"ADDR"},
				Value:       ":8084",
				Destination: &cfg.addr,
			},
			&cli.StringFlag{
				Name:        "l1-bridge-address",
				Usage:       "The address of the L1 bridge",
				EnvVars:     []string{"L1_BRIDGE_ADDRESS"},
				Value:       "0x54d6c1435ac7b90a5d46d01ee2f22ed6ff270ed3",
				Destination: &cfg.l1BridgeAddress,
			},
		},
		Action: func(c *cli.Context) error {

			ctx, stop := signal.NotifyContext(c.Context, os.Interrupt, os.Kill)
			defer stop()

			l1Client, err := ethclient.Dial(cfg.l1ExecutionURL)
			if err != nil {
				return fmt.Errorf("failed to dial L1 execution layer: %w", err)
			}
			defer l1Client.Close()

			l2Client, err := ethclient.Dial(cfg.l2ExecutionURL)
			if err != nil {
				return fmt.Errorf("failed to dial L2 execution layer: %w", err)
			}
			defer l2Client.Close()

			bridgeAddress := common.HexToAddress(cfg.l1BridgeAddress)

			log := log.With("l1_bridge_address", bridgeAddress)

			fromBlock := uint64(8311163 - 200)

			logsChan := make(chan types.Log, 200)
			sub, err := l1Client.SubscribeFilterLogs(
				ctx,
				ethereum.FilterQuery{
					Addresses: []common.Address{bridgeAddress},
					Topics:    [][]common.Hash{{ethBridgeInitiatedEvent}},
					FromBlock: big.NewInt(int64(fromBlock)),
				},
				logsChan,
			)

			go func() {
				select {
				case err := <-sub.Err():
					log.Error("subscription error", "error", err)
					stop()
				case <-ctx.Done():
					log.Info("context done")
					return
				}
			}()

			// fromBlock, err := l1Client.BlockNumber(ctx)
			// if err != nil {
			// 	return fmt.Errorf("failed to get current block number: %w", err)
			// }

			// fromBlock = currentBlock.Uint64()

			if err != nil {
				return fmt.Errorf("failed to subscribe to filter logs: %w", err)
			}
			log.Info("subscribed to filter logs")

			for {

				select {
				case l := <-logsChan:
					log.Info("got log", "block_number", l.BlockNumber, "tx_hash", l.TxHash, "address", l.Address, "topics", l.Topics)
				case <-ctx.Done():
					log.Info("context done")
					return nil
				}

				// log.Info("got log", "block_number", l.BlockNumber, "tx_hash", l.TxHash, "address", l.Address, "topics", l.Topics)

				// logs, err := l1Client.FilterLogs(ctx, ethereum.FilterQuery{
				// 	Addresses: []common.Address{bridgeAddress},
				// 	Topics:    [][]common.Hash{{ethBridgeInitiatedEvent}},
				// 	FromBlock: big.NewInt(int64(fromBlock)),
				// })
				// if err != nil {
				// 	return fmt.Errorf("failed to filter logs: %w", err)
				// }

				// if len(logs) == 0 {
				// 	log.Info("no logs found", "from_block", fromBlock)
				// 	break
				// }

				// log.Info("got logs", "from_block", fromBlock, "count", len(logs))

				// fromBlock = logs[len(logs)-1].BlockNumber + 1

			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error("Error running bridge monitor", "error", err)
		os.Exit(1)
	}
}
