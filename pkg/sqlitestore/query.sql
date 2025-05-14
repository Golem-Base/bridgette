-- name: InsertL1StandardBridgeETHDepositInitiated :exec
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
);

-- name: GetBlockPointer :one
SELECT block_number FROM BLOCK_POINTERS WHERE name = ? LIMIT 1;

-- name: UpdateBlockPointer :exec
UPDATE BLOCK_POINTERS SET block_number = ? WHERE name = ?;

