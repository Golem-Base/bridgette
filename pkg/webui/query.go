package webui

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/Golem-Base/bridgette/pkg/sqlitestore"
)

// DepositPair represents a matched pair of L1 and L2 deposit events
type DepositPair struct {
	ID              int64
	FromAddress     string
	ToAddress       string
	Amount          float64
	L1BlockNumber   int64
	L2BlockNumber   int64
	L1Timestamp     time.Time
	L2Timestamp     time.Time
	TimeDiffSeconds int64
	TxHashL1        string
	TxHashL2        string
}

// UnmatchedDeposit represents an unmatched L1 deposit event
type UnmatchedDeposit struct {
	ID               int64
	FromAddress      string
	ToAddress        string
	Amount           float64
	L1BlockNumber    int64
	L1Timestamp      time.Time
	TimeSinceSeconds int64
	TxHashL1         string
}

// GetMatchedDeposits returns a list of matched deposit pairs with time difference information
func GetMatchedDeposits(ctx context.Context, db *sql.DB, limit, offset int) ([]DepositPair, error) {
	queries := sqlitestore.New(db)

	rows, err := queries.GetMatchedDeposits(ctx, sqlitestore.GetMatchedDepositsParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, err
	}

	var deposits []DepositPair
	for _, row := range rows {
		// Convert TimeDiffSeconds from interface{} to int64
		var timeDiff int64
		switch v := row.TimeDiffSeconds.(type) {
		case int64:
			timeDiff = v
		case float64:
			timeDiff = int64(v)
		case nil:
			timeDiff = 0 // Default to 0 if null
		}

		deposit := DepositPair{
			ID:              row.ID,
			FromAddress:     "0x" + hex.EncodeToString(row.FromAddress),
			ToAddress:       "0x" + hex.EncodeToString(row.ToAddress),
			Amount:          row.Amount,
			L1BlockNumber:   row.L1BlockNumber,
			L2BlockNumber:   row.L2BlockNumber,
			L1Timestamp:     time.Unix(row.L1Timestamp, 0),
			L2Timestamp:     time.Unix(row.L2Timestamp, 0),
			TimeDiffSeconds: timeDiff,
			TxHashL1:        "0x" + hex.EncodeToString(row.TxHashL1),
			TxHashL2:        "0x" + hex.EncodeToString(row.TxHashL2),
		}
		deposits = append(deposits, deposit)
	}

	return deposits, nil
}

// GetTotalMatchedDeposits returns the total number of matched deposits
func GetTotalMatchedDeposits(ctx context.Context, db *sql.DB) (int, error) {
	queries := sqlitestore.New(db)

	count, err := queries.GetTotalMatchedDeposits(ctx)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// GetUnmatchedDeposits returns a list of unmatched L1 deposit events
func GetUnmatchedDeposits(ctx context.Context, db *sql.DB, limit, offset int) ([]UnmatchedDeposit, error) {
	queries := sqlitestore.New(db)

	rows, err := queries.GetUnmatchedDeposits(ctx, sqlitestore.GetUnmatchedDepositsParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, err
	}

	var deposits []UnmatchedDeposit
	for _, row := range rows {
		// Convert TimeSinceSeconds from interface{} to int64
		var timeSince int64
		switch v := row.TimeSinceSeconds.(type) {
		case int64:
			timeSince = v
		case float64:
			timeSince = int64(v)
		case nil:
			timeSince = 0 // Default to 0 if null
		}

		deposit := UnmatchedDeposit{
			ID:               row.ID,
			FromAddress:      "0x" + hex.EncodeToString(row.FromAddress),
			ToAddress:        "0x" + hex.EncodeToString(row.ToAddress),
			Amount:           row.Amount,
			L1BlockNumber:    row.L1BlockNumber,
			L1Timestamp:      time.Unix(row.L1Timestamp, 0),
			TimeSinceSeconds: timeSince,
			TxHashL1:         "0x" + hex.EncodeToString(row.TxHashL1),
		}
		deposits = append(deposits, deposit)
	}

	return deposits, nil
}

// GetTotalUnmatchedDeposits returns the total number of unmatched deposits
func GetTotalUnmatchedDeposits(ctx context.Context, db *sql.DB) (int, error) {
	queries := sqlitestore.New(db)

	count, err := queries.GetTotalUnmatchedDeposits(ctx)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// GetBridgeStats returns statistics about the bridge
func GetBridgeStats(ctx context.Context, db *sql.DB) (map[string]interface{}, error) {
	queries := sqlitestore.New(db)

	// Get main bridge stats
	stats, err := queries.GetBridgeStats(ctx)
	if err != nil {
		return nil, err
	}

	// Get pending deposits count
	pendingDeposits, err := queries.GetPendingDeposits(ctx)
	if err != nil {
		return nil, err
	}

	// Get latest L1 block info
	latestL1Block, err := queries.GetLatestL1Block(ctx)
	if err != nil {
		// Handle case when no L1 blocks exist yet
		if err == sql.ErrNoRows {
			// Create a zero struct
			latestL1Block = sqlitestore.GetLatestL1BlockRow{}
		} else {
			return nil, err
		}
	}

	// Get latest L2 block info
	latestL2Block, err := queries.GetLatestL2Block(ctx)
	if err != nil {
		// Handle case when no L2 blocks exist yet
		if err == sql.ErrNoRows {
			// Create a zero struct
			latestL2Block = sqlitestore.GetLatestL2BlockRow{}
		} else {
			return nil, err
		}
	}

	// Calculate time since last block
	currentTime := time.Now().Unix()
	var l1TimeSince, l2TimeSince float64

	// Only calculate time since if we have valid timestamps
	if latestL1Block.BlockTimestamp != nil && *latestL1Block.BlockTimestamp > 0 {
		l1TimeSince = float64(currentTime - *latestL1Block.BlockTimestamp)
	}
	if latestL2Block.BlockTimestamp != nil && *latestL2Block.BlockTimestamp > 0 {
		l2TimeSince = float64(currentTime - *latestL2Block.BlockTimestamp)
	}

	// Handle nullable fields
	var avgTimeDiff, minTimeDiff, maxTimeDiff, totalBridgedEth float64

	if stats.AvgTimeDiff != nil {
		avgTimeDiff = *stats.AvgTimeDiff
	}

	// Convert from interface{} types
	if stats.MinTimeDiff != nil {
		switch v := stats.MinTimeDiff.(type) {
		case float64:
			minTimeDiff = v
		case int64:
			minTimeDiff = float64(v)
		}
	}

	if stats.MaxTimeDiff != nil {
		switch v := stats.MaxTimeDiff.(type) {
		case float64:
			maxTimeDiff = v
		case int64:
			maxTimeDiff = float64(v)
		}
	}

	if stats.TotalBridgedEth != nil {
		totalBridgedEth = *stats.TotalBridgedEth
	}

	// Get block numbers, handling null values
	var l1BlockNum, l2BlockNum int
	if latestL1Block.BlockNumber != nil {
		l1BlockNum = int(*latestL1Block.BlockNumber)
	}
	if latestL2Block.BlockNumber != nil {
		l2BlockNum = int(*latestL2Block.BlockNumber)
	}

	// Create a map with the results
	result := map[string]interface{}{
		"total_matched":     int(stats.TotalMatched),
		"avg_time_diff":     avgTimeDiff,
		"min_time_diff":     minTimeDiff,
		"max_time_diff":     maxTimeDiff,
		"total_bridged_eth": totalBridgedEth,
		"pending_deposits":  int(pendingDeposits),
		"latest_l1_block":   l1BlockNum,
		"latest_l2_block":   l2BlockNum,
		"l1_time_since":     l1TimeSince,
		"l2_time_since":     l2TimeSince,
	}

	return result, nil
}
