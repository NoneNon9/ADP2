#!/bin/bash
set -e

# 1. Create the separate databases for each bounded context
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE orderdb;
    CREATE DATABASE paymentdb;
EOSQL

# 2. Run Order Service Migrations
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "orderdb" -f /migrations/order_init.sql

# 3. Run Payment Service Migrations
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "paymentdb" -f /migrations/payment_init.sql

echo "Databases and schemas initialized successfully!"