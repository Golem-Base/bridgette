CREATE TABLE IF NOT EXISTS l1_standard_bridge_eth_deposit_initiated (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    block_number UNSIGNED BIG INT NOT NULL,
    block_timestamp UNSIGNED BIG INT NOT NULL,
    tx_hash BLOB NOT NULL,
    from_address BLOB NOT NULL,
    to_address BLOB NOT NULL,
    amount REAL NOT NULL,
    event BLOB NOT NULL,
    matching_hash BLOB NOT NULL,
    matched_l2_standard_bridge_deposit_finalized_id INTEGER
);

CREATE TABLE IF NOT EXISTS l2_standard_bridge_deposit_finalized (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    block_number UNSIGNED BIG INT NOT NULL,
    block_timestamp UNSIGNED BIG INT NOT NULL,
    tx_hash BLOB NOT NULL,
    from_address BLOB NOT NULL,
    to_address BLOB NOT NULL,
    l1_token BLOB NOT NULL,
    amount REAL NOT NULL,
    event BLOB NOT NULL,
    matching_hash BLOB NOT NULL,
    matched_l1_standard_bridge_eth_deposit_initiated_id INTEGER
);

CREATE INDEX IF NOT EXISTS idx_l1_standard_bridge_eth_deposit_initiated_matching_hash ON l1_standard_bridge_eth_deposit_initiated(matching_hash);
CREATE INDEX IF NOT EXISTS idx_l2_standard_bridge_deposit_finalized_matching_hash ON l2_standard_bridge_deposit_finalized(matching_hash);

CREATE INDEX IF NOT EXISTS idx_l1_standard_bridge_eth_deposit_initiated_matched_l2_standard_bridge_deposit_finalized_id ON l1_standard_bridge_eth_deposit_initiated(matched_l2_standard_bridge_deposit_finalized_id);
CREATE INDEX IF NOT EXISTS idx_l2_standard_bridge_deposit_finalized_matched_l1_standard_bridge_eth_deposit_initiated_id ON l2_standard_bridge_deposit_finalized(matched_l1_standard_bridge_eth_deposit_initiated_id);