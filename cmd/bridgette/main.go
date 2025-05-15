package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"os/signal"

	"github.com/Golem-Base/bridgette/pkg/logparser"
	"github.com/Golem-Base/bridgette/pkg/sqlitestore"
	"github.com/Golem-Base/bridgette/pkg/webui"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/mattn/go-sqlite3"
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

const L1_ETH_DEPOSIT_INITIATED_LOW_BLOCK = "l1_standard_bridge_eth_deposit_initiated_lowest_processed_block"
const L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK = "l1_standard_bridge_eth_deposit_initiated_last_processed_block"
const L2_ETH_DEPOSIT_FINALIZED_LOW_BLOCK = "l2_standard_bridge_eth_deposit_finalized_lowest_processed_block"
const L2_ETH_DEPOSIT_FINALIZED_LAST_BLOCK = "l2_standard_bridge_eth_deposit_finalized_last_processed_block"

// Helper function to convert Wei to ETH (1 ETH = 10^18 Wei)
func weiToEth(wei *big.Int) float64 {
	if wei == nil {
		return 0
	}
	// Create a big float from wei
	weiFloat := new(big.Float).SetInt(wei)
	// Create 10^18 as a big float
	ethUnit := new(big.Float).SetInt(big.NewInt(1e18))
	// Divide wei by 10^18 to get ETH
	ethFloat := new(big.Float).Quo(weiFloat, ethUnit)
	// Convert to float64
	eth, _ := ethFloat.Float64()
	return eth
}

func main() {

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := struct {
		l1ExecutionURL  string
		l2ExecutionURL  string
		dbURL           string
		addr            string
		l1BridgeAddress string
		webUIAddr       string
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
				Name:        "db-url",
				Usage:       "The URL of the database",
				EnvVars:     []string{"DB_URL"},
				Destination: &cfg.dbURL,
				Value:       "file:./store/bridgette.db?_txlock=immediate&_auto_vacuum=2&_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=true",
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
			&cli.StringFlag{
				Name:        "web-ui-addr",
				Usage:       "Address for the web UI",
				EnvVars:     []string{"WEB_UI_ADDR"},
				Value:       ":8085",
				Destination: &cfg.webUIAddr,
			},
		},
		Action: func(c *cli.Context) error {

			// Open database
			db, err := sql.Open("sqlite3", cfg.dbURL)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()
			log.Info("database opened", "url", cfg.dbURL)

			err = sqlitestore.Migrate(db)
			if err != nil {
				return fmt.Errorf("failed to migrate database: %w", err)
			}

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

			autocommitStore := sqlitestore.New(db)

			eg, egCtx := errgroup.WithContext(ctx)

			// Add web UI server
			webServer := webui.NewServer(db, log.With("component", "webui"), cfg.webUIAddr)
			eg.Go(func() error {
				return webServer.Start(egCtx)
			})

			eg.Go(func() error {

				log := log.With("chain", "l1")

				fromBlock, err := l1Client.BlockNumber(egCtx)
				if err != nil {
					return fmt.Errorf("failed to get current block number: %w", err)
				}

				lowestProcessedBlock, err := autocommitStore.GetBlockPointer(egCtx, L1_ETH_DEPOSIT_INITIATED_LOW_BLOCK)
				if err != nil {
					return fmt.Errorf("failed to get lowest processed block: %w", err)
				}

				if lowestProcessedBlock.BlockNumber != nil {
					fromBlock = uint64(*lowestProcessedBlock.BlockNumber)
				}

				for fromBlock > 0 {

					toBlock := fromBlock - 1

					if fromBlock > 10_000 {
						fromBlock -= 10_000
					} else {
						fromBlock = 0
					}

					// Update the last processed block pointer if it's NULL
					toBlockNumber := int64(toBlock)

					// Get block time for the latest block directly from RPC client
					toBlockHeader, err := l1Client.HeaderByNumber(egCtx, big.NewInt(int64(toBlock)))
					if err != nil {
						return fmt.Errorf("failed to get header for toBlock: %w", err)
					}
					toBlockTime := int64(toBlockHeader.Time)

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

					// Get block time for the lowest block
					lowBlockHeader, err := l1Client.HeaderByNumber(egCtx, big.NewInt(int64(fromBlock)))
					if err != nil {
						return fmt.Errorf("failed to get header for lowest block: %w", err)
					}
					lowBlockTime := int64(lowBlockHeader.Time)

					blockTimes := make(map[uint64]uint64)

					for _, log := range logs {
						header, err := l1Client.HeaderByNumber(egCtx, big.NewInt(int64(log.BlockNumber)))
						if err != nil {
							return fmt.Errorf("failed to get header: %w", err)
						}
						blockTimes[log.BlockNumber] = header.Time
					}

					// Start a transaction at the beginning of each loop iteration
					tx, err := db.Begin()
					if err != nil {
						return fmt.Errorf("failed to begin transaction: %w", err)
					}
					// Create a transaction-wrapped store
					txStore := sqlitestore.New(tx).WithTx(tx)

					defer func() {
						// If we exit with error, ensure we roll back the transaction
						if err != nil {
							tx.Rollback()
						}
					}()

					for _, lg := range logs {
						// Parse the event data
						event, err := logparser.ParseL1StandardBridgeETHDepositInitiatedEvent(&lg)
						if err != nil {
							return fmt.Errorf("failed to parse log: %w", err)
						}

						eventJSON, err := json.Marshal(lg)
						if err != nil {
							return fmt.Errorf("failed to marshal event: %w", err)
						}

						// Insert log data into database and get the ID directly
						lastInsertID, err := txStore.InsertL1StandardBridgeETHDepositInitiated(egCtx, sqlitestore.InsertL1StandardBridgeETHDepositInitiatedParams{
							BlockNumber:    int64(lg.BlockNumber),
							BlockTimestamp: int64(blockTimes[lg.BlockNumber]),
							TxHash:         lg.TxHash.Bytes(),
							FromAddress:    event.From.Bytes(),
							ToAddress:      event.To.Bytes(),
							Amount:         weiToEth(event.Amount), // Convert Wei to ETH
							Event:          eventJSON,
							MatchingHash:   event.DepositMatchingHash().Bytes(),
						})
						if err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to insert log: %w", err)
						}

						// Find matching L2 deposit finalized event
						// Matching criteria:
						// 1. Matching hash must match
						// 2. L2 event timestamp must be >= L1 event timestamp (happened at same time or later)
						// 3. L2 event must not be already matched
						// We order by timestamp to get the earliest matching L2 event (lowest time difference)
						matchingL2Deposits, err := txStore.FindMatchingL2Deposits(egCtx, sqlitestore.FindMatchingL2DepositsParams{
							MatchingHash:   event.DepositMatchingHash().Bytes(),
							BlockTimestamp: int64(blockTimes[lg.BlockNumber]),
						})
						if err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to find matching L2 deposits: %w", err)
						}

						// If we found a matching L2 deposit, update the match
						if len(matchingL2Deposits) > 0 {
							match := matchingL2Deposits[0]

							// Update the L2 deposit with the L1 deposit ID
							err = txStore.UpdateL2DepositWithMatch(egCtx, sqlitestore.UpdateL2DepositWithMatchParams{
								MatchedL1StandardBridgeEthDepositInitiatedID: &lastInsertID,
								ID: match.ID,
							})
							if err != nil {
								tx.Rollback()
								return fmt.Errorf("failed to update L2 deposit with match: %w", err)
							}

							// Update the L1 deposit with the L2 deposit ID
							matchedID := match.ID // Create a variable to get a pointer
							err = txStore.UpdateL1DepositWithMatch(egCtx, sqlitestore.UpdateL1DepositWithMatchParams{
								MatchedL2StandardBridgeDepositFinalizedID: &matchedID,
								ID: lastInsertID,
							})
							if err != nil {
								tx.Rollback()
								return fmt.Errorf("failed to update L1 deposit with match: %w", err)
							}

							// Use the correct log variable
							log.Info("matched deposits",
								"l1_deposit_id", lastInsertID,
								"l2_deposit_id", match.ID,
								"time_difference_seconds", match.BlockTimestamp-int64(blockTimes[lg.BlockNumber]))
						}
					}

					if len(logs) == 0 {
						log.Info("no logs found", "from_block", fromBlock)
					} else {
						log.Info("got logs", "from_block", fromBlock, "count", len(logs))
					}

					blockNumber := int64(fromBlock)

					err = txStore.UpdateBlockPointer(egCtx, sqlitestore.UpdateBlockPointerParams{
						BlockNumber: &blockNumber,
						BlockTime:   &lowBlockTime,
						Name:        L1_ETH_DEPOSIT_INITIATED_LOW_BLOCK,
					})
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to update block pointer: %w", err)
					}

					err = txStore.UpdateBlockPointerIfNull(egCtx, sqlitestore.UpdateBlockPointerIfNullParams{
						BlockNumber: &toBlockNumber,
						BlockTime:   &toBlockTime,
						Name:        L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK,
					})
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to update last block pointer: %w", err)
					}

					// Commit the transaction
					err = tx.Commit()
					if err != nil {
						return fmt.Errorf("failed to commit transaction: %w", err)
					}
				}

				return nil
			})

			eg.Go(func() error {

				log := log.With("chain", "l2")

				fromBlock, err := l2Client.BlockNumber(egCtx)
				if err != nil {
					return fmt.Errorf("failed to get current block number: %w", err)
				}

				lowestProcessedBlock, err := autocommitStore.GetBlockPointer(egCtx, L2_ETH_DEPOSIT_FINALIZED_LOW_BLOCK)
				if err != nil {
					return fmt.Errorf("failed to get lowest processed block: %w", err)
				}

				if lowestProcessedBlock.BlockNumber != nil {
					fromBlock = uint64(*lowestProcessedBlock.BlockNumber)
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

					blockTimes := make(map[uint64]uint64)

					for _, log := range logs {
						header, err := l2Client.HeaderByNumber(egCtx, big.NewInt(int64(log.BlockNumber)))
						if err != nil {
							return fmt.Errorf("failed to get header: %w", err)
						}
						blockTimes[log.BlockNumber] = header.Time
					}

					// Get block time for the lowest block
					lowBlockHeader, err := l2Client.HeaderByNumber(egCtx, big.NewInt(int64(fromBlock)))
					if err != nil {
						return fmt.Errorf("failed to get header for lowest block: %w", err)
					}
					lowBlockTime := int64(lowBlockHeader.Time)

					// Update the last processed block pointer if it's NULL
					toBlockNumber := int64(toBlock)

					// Get block time for the latest block directly from RPC client
					toBlockHeader, err := l2Client.HeaderByNumber(egCtx, big.NewInt(int64(toBlock)))
					if err != nil {
						return fmt.Errorf("failed to get header for toBlock: %w", err)
					}
					toBlockTime := int64(toBlockHeader.Time)

					// Start a transaction at the beginning of each loop iteration
					tx, err := db.Begin()
					if err != nil {
						return fmt.Errorf("failed to begin transaction: %w", err)
					}
					// Create a transaction-wrapped store
					txStore := sqlitestore.New(tx).WithTx(tx)

					defer func() {
						// If we exit with error, ensure we roll back the transaction
						if err != nil {
							tx.Rollback()
						}
					}()

					for _, lg := range logs {

						event, err := logparser.ParseL2StandardBridgeDepositFinalizedEvent(&lg)
						if err != nil {
							return fmt.Errorf("failed to parse log: %w", err)
						}

						eventJSON, err := json.Marshal(lg)
						if err != nil {
							return fmt.Errorf("failed to marshal event: %w", err)
						}

						// Insert log data into database instead of file
						l2DepositID, err := txStore.InsertL2StandardBridgeDepositFinalized(egCtx, sqlitestore.InsertL2StandardBridgeDepositFinalizedParams{
							BlockNumber:    int64(lg.BlockNumber),
							BlockTimestamp: int64(blockTimes[lg.BlockNumber]),
							TxHash:         lg.TxHash.Bytes(),
							FromAddress:    event.From.Bytes(),
							ToAddress:      event.To.Bytes(),
							L1Token:        event.L1Token.Bytes(),
							Amount:         weiToEth(event.Amount), // Convert Wei to ETH
							Event:          eventJSON,
							MatchingHash:   event.DepositMatchingHash().Bytes(),
						})
						if err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to insert log: %w", err)
						}

						// Find matching L1 deposit initiated event
						// Matching criteria:
						// 1. Matching hash must match
						// 2. L1 event timestamp must be <= L2 event timestamp (happened before or at the same time)
						// 3. L1 event must not be already matched
						// We order by timestamp DESC to get the latest matching L1 event (lowest time difference)
						matchingL1Deposits, err := txStore.FindMatchingL1Deposits(egCtx, sqlitestore.FindMatchingL1DepositsParams{
							MatchingHash:   event.DepositMatchingHash().Bytes(),
							BlockTimestamp: int64(blockTimes[lg.BlockNumber]),
						})
						if err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to find matching L1 deposits: %w", err)
						}

						// If we found a matching L1 deposit, update the match
						if len(matchingL1Deposits) > 0 {
							match := matchingL1Deposits[0]

							// Update the L1 deposit with the L2 deposit ID
							l2ID := l2DepositID // Create a variable to get a pointer
							err = txStore.UpdateL1DepositWithMatch(egCtx, sqlitestore.UpdateL1DepositWithMatchParams{
								MatchedL2StandardBridgeDepositFinalizedID: &l2ID,
								ID: match.ID,
							})
							if err != nil {
								tx.Rollback()
								return fmt.Errorf("failed to update L1 deposit with match: %w", err)
							}

							// Update the L2 deposit with the L1 deposit ID
							err = txStore.UpdateL2DepositWithMatch(egCtx, sqlitestore.UpdateL2DepositWithMatchParams{
								MatchedL1StandardBridgeEthDepositInitiatedID: &match.ID,
								ID: l2DepositID,
							})
							if err != nil {
								tx.Rollback()
								return fmt.Errorf("failed to update L2 deposit with match: %w", err)
							}

							log.Info("matched deposits (from L2)",
								"l1_deposit_id", match.ID,
								"l2_deposit_id", l2DepositID,
								"time_difference_seconds", int64(blockTimes[lg.BlockNumber])-match.BlockTimestamp)
						}
					}

					if len(logs) == 0 {
						log.Info("no logs found", "from_block", fromBlock)
					} else {
						log.Info("got logs", "from_block", fromBlock, "count", len(logs))
					}

					blockNumber := int64(fromBlock)

					err = txStore.UpdateBlockPointer(egCtx, sqlitestore.UpdateBlockPointerParams{
						Name:        L2_ETH_DEPOSIT_FINALIZED_LOW_BLOCK,
						BlockNumber: &blockNumber,
						BlockTime:   &lowBlockTime,
					})
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to update block pointer: %w", err)
					}

					err = txStore.UpdateBlockPointerIfNull(egCtx, sqlitestore.UpdateBlockPointerIfNullParams{
						BlockNumber: &toBlockNumber,
						BlockTime:   &toBlockTime,
						Name:        L2_ETH_DEPOSIT_FINALIZED_LAST_BLOCK,
					})
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("failed to update last block pointer: %w", err)
					}

					// Commit the transaction
					err = tx.Commit()
					if err != nil {
						return fmt.Errorf("failed to commit transaction: %w", err)
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
