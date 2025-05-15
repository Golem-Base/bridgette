-- name: InsertL1StandardBridgeETHDepositInitiated :one
INSERT INTO l1_standard_bridge_eth_deposit_initiated (
    block_number,
    block_timestamp,
    tx_hash,
    from_address,
    to_address,
    amount,
    event,
    matching_hash
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING id;

-- name: GetBlockPointer :one
SELECT block_number FROM BLOCK_POINTERS WHERE name = ? LIMIT 1;

-- name: UpdateBlockPointer :exec
UPDATE BLOCK_POINTERS SET block_number = ? WHERE name = ?;

-- name: InsertL2StandardBridgeDepositFinalized :one
INSERT INTO l2_standard_bridge_deposit_finalized (
    block_number,
    block_timestamp,
    tx_hash,
    from_address,
    to_address,
    l1_token,
    amount,
    event,
    matching_hash
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING id;

-- name: FindMatchingL2Deposits :many
SELECT 
    id,
    block_timestamp,
    matching_hash
FROM l2_standard_bridge_deposit_finalized
WHERE 
    matching_hash = ? AND
    block_timestamp >= ? AND
    matched_l1_standard_bridge_eth_deposit_initiated_id IS NULL
ORDER BY block_timestamp ASC
LIMIT 1;

-- name: UpdateL2DepositWithMatch :exec
UPDATE l2_standard_bridge_deposit_finalized
SET 
    matched_l1_standard_bridge_eth_deposit_initiated_id = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateL1DepositWithMatch :exec
UPDATE l1_standard_bridge_eth_deposit_initiated
SET 
    matched_l2_standard_bridge_deposit_finalized_id = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: FindMatchingL1Deposits :many
SELECT 
    id,
    block_timestamp,
    matching_hash
FROM l1_standard_bridge_eth_deposit_initiated
WHERE 
    matching_hash = ? AND
    block_timestamp <= ? AND
    matched_l2_standard_bridge_deposit_finalized_id IS NULL
ORDER BY block_timestamp DESC
LIMIT 1;

-- Web UI Queries

-- name: GetMatchedDeposits :many
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
LIMIT ? OFFSET ?;

-- name: GetTotalMatchedDeposits :one
SELECT 
    COUNT(*)
FROM 
    l1_standard_bridge_eth_deposit_initiated
WHERE 
    matched_l2_standard_bridge_deposit_finalized_id IS NOT NULL;

-- name: GetBridgeStats :one
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
    l1.matched_l2_standard_bridge_deposit_finalized_id = l2.id;

-- name: GetPendingDeposits :one
SELECT 
    COUNT(*) 
FROM 
    l1_standard_bridge_eth_deposit_initiated
WHERE 
    matched_l2_standard_bridge_deposit_finalized_id IS NULL;

-- name: GetLatestL1Block :one
SELECT 
    block_number,
    block_timestamp
FROM 
    l1_standard_bridge_eth_deposit_initiated
ORDER BY 
    block_number DESC
LIMIT 1;

-- name: GetLatestL2Block :one
SELECT 
    block_number,
    block_timestamp
FROM 
    l2_standard_bridge_deposit_finalized
ORDER BY 
    block_number DESC
LIMIT 1;

