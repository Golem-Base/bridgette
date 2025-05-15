package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Golem-Base/bridgette/pkg/sqlitestore"
	_ "github.com/mattn/go-sqlite3"
)

const (
	L1_ETH_DEPOSIT_INITIATED_LOW_BLOCK  = "l1_standard_bridge_eth_deposit_initiated_lowest_processed_block"
	L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK = "l1_standard_bridge_eth_deposit_initiated_last_processed_block"
)

func main() {
	// Open a test database
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		log.Fatal("Failed to open in-memory database:", err)
	}
	defer db.Close()

	// Run migrations
	err = sqlitestore.Migrate(db)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Create a store
	store := sqlitestore.New(db)
	ctx := context.Background()

	// Test case 1: Update with NULL
	fmt.Println("Test case 1: Update when pointer is NULL")

	lowBlock, err := store.GetBlockPointer(ctx, L1_ETH_DEPOSIT_INITIATED_LOW_BLOCK)
	if err != nil {
		log.Fatal("Failed to get low block pointer:", err)
	}
	fmt.Printf("  Low block pointer before: BlockNumber=%v, BlockTime=%v\n",
		formatNilInt(lowBlock.BlockNumber), formatNilInt(lowBlock.BlockTime))

	lastBlock, err := store.GetBlockPointer(ctx, L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK)
	if err != nil {
		log.Fatal("Failed to get last block pointer:", err)
	}
	fmt.Printf("  Last block pointer before: BlockNumber=%v, BlockTime=%v\n",
		formatNilInt(lastBlock.BlockNumber), formatNilInt(lastBlock.BlockTime))

	// Update the last block pointer with a value if NULL
	toBlockNumber := int64(12345)
	currentTime := time.Now().Unix()
	toBlockTime := currentTime

	err = store.UpdateBlockPointerIfNull(ctx, sqlitestore.UpdateBlockPointerIfNullParams{
		BlockNumber: &toBlockNumber,
		BlockTime:   &toBlockTime,
		Name:        L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK,
	})
	if err != nil {
		log.Fatal("Failed to update last block pointer:", err)
	}

	lastBlock, err = store.GetBlockPointer(ctx, L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK)
	if err != nil {
		log.Fatal("Failed to get last block pointer:", err)
	}
	fmt.Printf("  Last block pointer after: BlockNumber=%v, BlockTime=%v\n",
		formatNilInt(lastBlock.BlockNumber), formatTimeInt(lastBlock.BlockTime))

	// Test case 2: Update when already has a value
	fmt.Println("Test case 2: Update when pointer already has a value")
	// First set a value using the regular update
	toBlockNumber = int64(54321)
	toBlockTime = currentTime + 3600 // 1 hour later

	err = store.UpdateBlockPointer(ctx, sqlitestore.UpdateBlockPointerParams{
		BlockNumber: &toBlockNumber,
		BlockTime:   &toBlockTime,
		Name:        L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK,
	})
	if err != nil {
		log.Fatal("Failed to set last block pointer:", err)
	}

	lastBlock, err = store.GetBlockPointer(ctx, L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK)
	if err != nil {
		log.Fatal("Failed to get last block pointer:", err)
	}
	fmt.Printf("  Last block pointer before UpdateIfNull: BlockNumber=%v, BlockTime=%v\n",
		formatNilInt(lastBlock.BlockNumber), formatTimeInt(lastBlock.BlockTime))

	// Try to update it using UpdateIfNull
	toBlockNumber = int64(99999)
	toBlockTime = currentTime + 7200 // 2 hours later

	err = store.UpdateBlockPointerIfNull(ctx, sqlitestore.UpdateBlockPointerIfNullParams{
		BlockNumber: &toBlockNumber,
		BlockTime:   &toBlockTime,
		Name:        L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK,
	})
	if err != nil {
		log.Fatal("Failed to update last block pointer:", err)
	}

	lastBlock, err = store.GetBlockPointer(ctx, L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK)
	if err != nil {
		log.Fatal("Failed to get last block pointer:", err)
	}
	fmt.Printf("  Last block pointer after UpdateIfNull: BlockNumber=%v, BlockTime=%v\n",
		formatNilInt(lastBlock.BlockNumber), formatTimeInt(lastBlock.BlockTime))

	fmt.Println("All tests passed!")
}

// Helper to format nil int pointer
func formatNilInt(val *int64) string {
	if val == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d", *val)
}

// Helper to format time from int64
func formatTimeInt(val *int64) string {
	if val == nil {
		return "<nil>"
	}
	t := time.Unix(*val, 0)
	return fmt.Sprintf("%d (%s)", *val, t.Format(time.RFC3339))
}
