CREATE TABLE payments (
                          id UUID PRIMARY KEY,
                          order_id VARCHAR(255) NOT NULL,
                          transaction_id VARCHAR(255) UNIQUE NOT NULL,
                          amount BIGINT NOT NULL,
                          status VARCHAR(50) NOT NULL,
                          created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);