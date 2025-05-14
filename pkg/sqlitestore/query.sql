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
)