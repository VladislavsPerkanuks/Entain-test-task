CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    balance DECIMAL(20, 2) NOT NULL DEFAULT 0.00
);

ALTER TABLE users ADD CONSTRAINT users_balance_non_negative CHECK (
    balance >= 0
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users (id),
    state VARCHAR(10) NOT NULL CHECK (state IN ('win', 'lose')),
    amount DECIMAL(20, 2) NOT NULL,
    source_type VARCHAR(20) NOT NULL CHECK (
        source_type IN ('game', 'server', 'payment')
    ),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
INSERT INTO users (balance) VALUES
(100.00),
(200.00),
(50.00);
