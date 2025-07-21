#!/bin/bash
source ../.env

if [ -z "$DB_USER" ] || [ -z "$DB_USER_PASSWORD" ] || [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_NAME" ]; then
    echo "Database configuration variables are not set in .env file"
    exit 1
fi

if [ -z "$DATABASE_URL" ]; then
    echo "DATABASE_URL is not set in .env file"
    exit 1
fi

# Set password for postgres user
export PGPASSWORD="$DB_POSTGRES_PASSWORD"

psql -U postgres -c "CREATE USER $DB_USER WITH PASSWORD '$DB_USER_PASSWORD';" 2>/dev/null || echo "User $DB_USER already exists"
psql -U postgres -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" 2>/dev/null || echo "Database $DB_NAME already exists"
psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;" 2>/dev/null

# Clear the password variable for security
unset PGPASSWORD

echo "Running database migrations..."
migrate -database "$DATABASE_URL?sslmode=disable" -path ./migrations up

# If migrations fail due to dirty state, offer to fix it
if [ $? -ne 0 ]; then
    echo "Migration failed. Attempting to fix dirty state..."
    migrate -database "$DATABASE_URL?sslmode=disable" -path ./migrations force 0
    migrate -database "$DATABASE_URL?sslmode=disable" -path ./migrations up
fi

# Verify tables were created
echo "Checking created tables..."
psql "$DATABASE_URL?sslmode=disable" -c "\dt"