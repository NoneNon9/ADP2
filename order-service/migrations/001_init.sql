CREATE TABLE orders (
                        id UUID PRIMARY KEY,
                        customer_id VARCHAR(255) NOT NULL,
                        item_name VARCHAR(255) NOT NULL,
                        amount BIGINT NOT NULL,
                        status VARCHAR(50) NOT NULL,
                        idempotency_key VARCHAR(255) UNIQUE,
                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);