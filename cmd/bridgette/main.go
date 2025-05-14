package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

// // ETHBridgeInitiated (index_topic_1 address from, index_topic_2 address to, uint256 amount, bytes extraData)
// // 0x2849b43074093a05396b6f2a937dee8565b15a48a7b3d4bffb732a5017380af5
// var ethBridgeInitiatedEvent = common.HexToHash("0x2849b43074093a05396b6f2a937dee8565b15a48a7b3d4bffb732a5017380af5")

// L1 - sending ETH
// ETHDepositInitiated (index_topic_1 address from, index_topic_2 address to, uint256 amount, bytes extraData)
var ethDepositInitiatedEvent = common.HexToHash("0x35d79ab81f2b2017e19afb5c5571778877782d7a8786f5907f93b0f4702f4f23")

var l2StandardBridgeAddress = common.HexToAddress("0x4200000000000000000000000000000000000010")

// L2 - receiving ETH
// DepositFinalized (index_topic_1 address l1Token, index_topic_2 address l2Token, index_topic_3 address from, address to, uint256 amount, bytes extraData)
var ethDepositFinalizedEvent = common.HexToHash("0xb0444523268717a02698be47d0803aa7468c00acbed2f8bd93a0459cde61dd89")

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

			// fromBlock := uint64(8311163 - 200)

			// logsChan := make(chan types.Log, 200)
			// sub, err := l1Client.SubscribeFilterLogs(
			// 	ctx,
			// 	ethereum.FilterQuery{
			// 		Addresses: []common.Address{bridgeAddress},
			// 		Topics:    [][]common.Hash{{ethBridgeInitiatedEvent}},
			// 		FromBlock: big.NewInt(int64(fromBlock)),
			// 	},
			// 	logsChan,
			// )

			// go func() {
			// 	select {
			// 	case err := <-sub.Err():
			// 		log.Error("subscription error", "error", err)
			// 		stop()
			// 	case <-ctx.Done():
			// 		log.Info("context done")
			// 		return
			// 	}
			// }()

			eg, egCtx := errgroup.WithContext(ctx)

			eg.Go(func() error {

				log := log.With("chain", "l1")

				l1OutpuPath := filepath.Join("output", "l1")

				fromBlock, err := l1Client.BlockNumber(egCtx)
				if err != nil {
					return fmt.Errorf("failed to get current block number: %w", err)
				}

				for fromBlock > 0 {

					toBlock := fromBlock - 1

					if fromBlock > 10_000 {
						fromBlock -= 10_000
					} else {
						fromBlock = 0
					}

					log.Info("filtering logs", "from_block", fromBlock, "to_block", toBlock)

					logs, err := l1Client.FilterLogs(egCtx, ethereum.FilterQuery{
						Addresses: []common.Address{bridgeAddress},
						Topics:    [][]common.Hash{{ethDepositInitiatedEvent}},
						FromBlock: big.NewInt(int64(fromBlock)),
						ToBlock:   big.NewInt(int64(toBlock)),
					})
					if err != nil {
						return fmt.Errorf("failed to filter logs: %w", err)
					}

					for _, log := range logs {
						fileName := filepath.Join(l1OutpuPath, fmt.Sprintf("%018d-%04d-%04d.json", log.BlockNumber, log.TxIndex, log.Index))

						file, err := os.Create(fileName)
						if err != nil {
							return fmt.Errorf("failed to create file: %w", err)
						}

						err = json.NewEncoder(file).Encode(log)
						if err != nil {
							return fmt.Errorf("failed to encode log: %w", err)
						}

						err = file.Close()
						if err != nil {
							return fmt.Errorf("failed to close file: %w", err)
						}

					}

					// transactionsHashes := map[common.Hash]bool{}

					// for _, log := range logs {
					// 	transactionsHashes[log.TxHash] = true
					// }

					// for txHash := range transactionsHashes {
					// 	receipt, err := l1Client.TransactionReceipt(egCtx, txHash)
					// 	if err != nil {
					// 		return fmt.Errorf("failed to get transaction receipt: %w", err)
					// 	}

					// 	fileName := filepath.Join(l1OutpuPath, fmt.Sprintf("%018d-%04d.json", receipt.BlockNumber.Uint64(), receipt.TransactionIndex))

					// 	file, err := os.Create(fileName)
					// 	if err != nil {
					// 		return fmt.Errorf("failed to create file: %w", err)
					// 	}

					// 	err = json.NewEncoder(file).Encode(receipt)
					// 	if err != nil {
					// 		return fmt.Errorf("failed to encode receipt: %w", err)
					// 	}

					// 	err = file.Close()
					// 	if err != nil {
					// 		return fmt.Errorf("failed to close file: %w", err)
					// 	}

					// }

					if len(logs) == 0 {
						log.Info("no logs found", "from_block", fromBlock)
						break
					} else {
						log.Info("got logs", "from_block", fromBlock, "count", len(logs))
					}

				}

				return nil
			})

			eg.Go(func() error {

				log := log.With("chain", "l2")

				l2OutpuPath := filepath.Join("output", "l2")

				fromBlock, err := l2Client.BlockNumber(egCtx)
				if err != nil {
					return fmt.Errorf("failed to get current block number: %w", err)
				}

				for fromBlock > 0 {

					toBlock := fromBlock - 1

					if fromBlock > 10_000 {
						fromBlock -= 10_000
					} else {
						fromBlock = 0
					}

					log.Info("filtering logs", "from_block", fromBlock, "to_block", toBlock)

					logs, err := l2Client.FilterLogs(egCtx, ethereum.FilterQuery{
						Addresses: []common.Address{l2StandardBridgeAddress},
						Topics:    [][]common.Hash{{ethDepositFinalizedEvent}},
						FromBlock: big.NewInt(int64(fromBlock)),
						ToBlock:   big.NewInt(int64(toBlock)),
					})
					if err != nil {
						return fmt.Errorf("failed to filter logs: %w", err)
					}

					for _, log := range logs {
						fileName := filepath.Join(l2OutpuPath, fmt.Sprintf("%018d-%04d-%04d.json", log.BlockNumber, log.TxIndex, log.Index))

						file, err := os.Create(fileName)
						if err != nil {
							return fmt.Errorf("failed to create file: %w", err)
						}

						err = json.NewEncoder(file).Encode(log)
						if err != nil {
							return fmt.Errorf("failed to encode log: %w", err)
						}

						err = file.Close()
						if err != nil {
							return fmt.Errorf("failed to close file: %w", err)
						}

					}

					if len(logs) == 0 {
						log.Info("no logs found", "from_block", fromBlock)
					} else {
						log.Info("got logs", "from_block", fromBlock, "count", len(logs))
					}

				}

				return nil
			})
			return eg.Wait()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error("Error running bridge monitor", "error", err)
		os.Exit(1)
	}
}
