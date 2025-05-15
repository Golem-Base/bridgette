package webui

import (
	"context"
	"database/sql"
	"encoding/hex"
	"time"
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

// GetMatchedDeposits returns a list of matched deposit pairs with time difference information
func GetMatchedDeposits(ctx context.Context, db *sql.DB, limit, offset int) ([]DepositPair, error) {
	query := `
	SELECT 
		l1.id,
		l1.from_address,
		l1.to_address,
		l1.amount,
		l1.block_number as l1_block_number,
		l2.block_number as l2_block_number,
		l1.block_timestamp as l1_timestamp,
		l2.block_timestamp as l2_timestamp,
		(l2.block_timestamp - l1.block_timestamp) as time_diff_seconds,
		l1.tx_hash as tx_hash_l1,
		l2.tx_hash as tx_hash_l2
	FROM 
		l1_standard_bridge_eth_deposit_initiated l1
	JOIN 
		l2_standard_bridge_deposit_finalized l2 
	ON 
		l1.matched_l2_standard_bridge_deposit_finalized_id = l2.id
	ORDER BY 
		l1.block_timestamp DESC
	LIMIT ? OFFSET ?
	`

	rows, err := db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deposits []DepositPair
	for rows.Next() {
		var d DepositPair
		var fromAddrBytes, toAddrBytes []byte
		var l1Timestamp, l2Timestamp int64
		var txHashL1, txHashL2 []byte

		if err := rows.Scan(
			&d.ID,
			&fromAddrBytes,
			&toAddrBytes,
			&d.Amount,
			&d.L1BlockNumber,
			&d.L2BlockNumber,
			&l1Timestamp,
			&l2Timestamp,
			&d.TimeDiffSeconds,
			&txHashL1,
			&txHashL2,
		); err != nil {
			return nil, err
		}

		d.FromAddress = "0x" + hex.EncodeToString(fromAddrBytes)
		d.ToAddress = "0x" + hex.EncodeToString(toAddrBytes)
		d.L1Timestamp = time.Unix(l1Timestamp, 0)
		d.L2Timestamp = time.Unix(l2Timestamp, 0)
		d.TxHashL1 = "0x" + hex.EncodeToString(txHashL1)
		d.TxHashL2 = "0x" + hex.EncodeToString(txHashL2)

		deposits = append(deposits, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deposits, nil
}

// GetTotalMatchedDeposits returns the total number of matched deposits
func GetTotalMatchedDeposits(ctx context.Context, db *sql.DB) (int, error) {
	query := `
	SELECT 
		COUNT(*)
	FROM 
		l1_standard_bridge_eth_deposit_initiated
	WHERE 
		matched_l2_standard_bridge_deposit_finalized_id IS NOT NULL
	`

	var count int
	err := db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetBridgeStats returns statistics about the bridge
func GetBridgeStats(ctx context.Context, db *sql.DB) (map[string]interface{}, error) {
	statsQuery := `
	SELECT 
		COUNT(*) as total_matched,
		AVG(l2.block_timestamp - l1.block_timestamp) as avg_time_diff,
		MIN(l2.block_timestamp - l1.block_timestamp) as min_time_diff,
		MAX(l2.block_timestamp - l1.block_timestamp) as max_time_diff,
		SUM(l1.amount) as total_bridged_eth
	FROM 
		l1_standard_bridge_eth_deposit_initiated l1
	JOIN 
		l2_standard_bridge_deposit_finalized l2 
	ON 
		l1.matched_l2_standard_bridge_deposit_finalized_id = l2.id
	`

	var totalMatched int
	var avgTimeDiff, minTimeDiff, maxTimeDiff, totalBridgedETH float64

	err := db.QueryRowContext(ctx, statsQuery).Scan(
		&totalMatched,
		&avgTimeDiff,
		&minTimeDiff,
		&maxTimeDiff,
		&totalBridgedETH,
	)
	if err != nil {
		return nil, err
	}

	// Get total deposits initiated but not yet finalized
	pendingQuery := `
	SELECT 
		COUNT(*) 
	FROM 
		l1_standard_bridge_eth_deposit_initiated
	WHERE 
		matched_l2_standard_bridge_deposit_finalized_id IS NULL
	`

	var pendingDeposits int
	err = db.QueryRowContext(ctx, pendingQuery).Scan(&pendingDeposits)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_matched":     totalMatched,
		"avg_time_diff":     avgTimeDiff,
		"min_time_diff":     minTimeDiff,
		"max_time_diff":     maxTimeDiff,
		"total_bridged_eth": totalBridgedETH,
		"pending_deposits":  pendingDeposits,
	}

	return stats, nil
}
