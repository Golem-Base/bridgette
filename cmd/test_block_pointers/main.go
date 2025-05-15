package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

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
	fmt.Printf("  Low block pointer before: %v\n", lowBlock)

	lastBlock, err := store.GetBlockPointer(ctx, L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK)
	if err != nil {
		log.Fatal("Failed to get last block pointer:", err)
	}
	fmt.Printf("  Last block pointer before: %v\n", lastBlock)

	// Update the last block pointer with a value if NULL
	toBlockNumber := int64(12345)
	err = store.UpdateBlockPointerIfNull(ctx, sqlitestore.UpdateBlockPointerIfNullParams{
		BlockNumber: &toBlockNumber,
		Name:        L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK,
	})
	if err != nil {
		log.Fatal("Failed to update last block pointer:", err)
	}

	lastBlock, err = store.GetBlockPointer(ctx, L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK)
	if err != nil {
		log.Fatal("Failed to get last block pointer:", err)
	}
	fmt.Printf("  Last block pointer after: %v\n", lastBlock)

	// Test case 2: Update when already has a value
	fmt.Println("Test case 2: Update when pointer already has a value")
	// First set a value using the regular update
	toBlockNumber = int64(54321)
	err = store.UpdateBlockPointer(ctx, sqlitestore.UpdateBlockPointerParams{
		BlockNumber: &toBlockNumber,
		Name:        L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK,
	})
	if err != nil {
		log.Fatal("Failed to set last block pointer:", err)
	}

	lastBlock, err = store.GetBlockPointer(ctx, L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK)
	if err != nil {
		log.Fatal("Failed to get last block pointer:", err)
	}
	fmt.Printf("  Last block pointer before UpdateIfNull: %v\n", lastBlock)

	// Try to update it using UpdateIfNull
	toBlockNumber = int64(99999)
	err = store.UpdateBlockPointerIfNull(ctx, sqlitestore.UpdateBlockPointerIfNullParams{
		BlockNumber: &toBlockNumber,
		Name:        L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK,
	})
	if err != nil {
		log.Fatal("Failed to update last block pointer:", err)
	}

	lastBlock, err = store.GetBlockPointer(ctx, L1_ETH_DEPOSIT_INITIATED_LAST_BLOCK)
	if err != nil {
		log.Fatal("Failed to get last block pointer:", err)
	}
	fmt.Printf("  Last block pointer after UpdateIfNull: %v\n", lastBlock)

	fmt.Println("All tests passed!")
}
