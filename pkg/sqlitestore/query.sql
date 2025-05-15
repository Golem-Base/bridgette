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

