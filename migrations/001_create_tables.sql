CREATE TABLE IF NOT EXISTS monitored_addresses (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    address VARCHAR(42) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_monitored_addresses_user_id ON monitored_addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_monitored_addresses_address ON monitored_addresses(address);
CREATE INDEX IF NOT EXISTS idx_monitored_addresses_active ON monitored_addresses(is_active);

CREATE TABLE IF NOT EXISTS processing_state (
    instance_id VARCHAR(255) PRIMARY KEY,
    last_processed_block BIGINT DEFAULT 0,
    stats_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS processed_transactions_log (
    id SERIAL PRIMARY KEY,
    transaction_hash VARCHAR(66) NOT NULL,
    block_number BIGINT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    source_address VARCHAR(42),
    destination_address VARCHAR(42),
    amount DECIMAL(78, 0),
    fees DECIMAL(78, 0),
    gas_used BIGINT,
    gas_price DECIMAL(78, 0),
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    kafka_published BOOLEAN DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_processed_transactions_hash ON processed_transactions_log(transaction_hash);
CREATE INDEX IF NOT EXISTS idx_processed_transactions_block ON processed_transactions_log(block_number);
CREATE INDEX IF NOT EXISTS idx_processed_transactions_user ON processed_transactions_log(user_id);
CREATE INDEX IF NOT EXISTS idx_processed_transactions_time ON processed_transactions_log(processed_at);

CREATE TABLE IF NOT EXISTS failed_transactions (
    id SERIAL PRIMARY KEY,
    transaction_hash VARCHAR(66) NOT NULL,
    block_number BIGINT NOT NULL,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_retry_at TIMESTAMP WITH TIME ZONE,
    resolved BOOLEAN DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_failed_transactions_resolved ON failed_transactions(resolved);
CREATE INDEX IF NOT EXISTS idx_failed_transactions_retry ON failed_transactions(retry_count, max_retries);

INSERT INTO monitored_addresses (user_id, address) VALUES
    ('user_1', '0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045'), -- Vitalik
    ('user_2', '0x47ac0Fb4F2D84898e4D9E7b4DaB3C24507a6D503'), -- Binance
    ('user_3', '0x28C6c06298d514Db089934071355E5743bf21d60'), -- Binance
    ('user_4', '0x21a31Ee1afC51d94C2eFcCAa2092aD1028285549'), -- Binance
    ('user_5', '0x3f5CE5FBFe3E9af3971dD833D26bA9b5C936f0bE')  -- Binance
ON CONFLICT (address) DO NOTHING;

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language plpgsql;

CREATE TRIGGER update_monitored_addresses_updated_at
    BEFORE UPDATE ON monitored_addresses 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_processing_state_updated_at 
    BEFORE UPDATE ON processing_state 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();