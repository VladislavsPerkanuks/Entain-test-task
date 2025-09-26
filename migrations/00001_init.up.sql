CREATE TABLE users (
    id UUID PRIMARY KEY,
    balance DECIMAL(20, 2) NOT NULL DEFAULT 0.00
);

ALTER TABLE users ADD CONSTRAINT users_balance_non_negative CHECK (
    balance >= 0
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id),
    state VARCHAR(10) NOT NULL CHECK (state IN ('win', 'lose')),
    amount DECIMAL(20, 2) NOT NULL,
    source_type VARCHAR(20) NOT NULL CHECK (source_type IN ('game', 'server', 'payment')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE processed_transactions (
    transaction_id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id),
    processed_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO users (id, balance) VALUES
('550e8400-e29b-41d4-a716-446655440001', 0.00),
('550e8400-e29b-41d4-a716-446655440002', 0.00),
('550e8400-e29b-41d4-a716-446655440003', 0.00);
