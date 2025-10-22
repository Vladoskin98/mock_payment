CREATE TABLE IF NOT EXISTS payments (
    id SERIAL PRIMARY KEY,
    provider_name VARCHAR(100) NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    payment_date VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);