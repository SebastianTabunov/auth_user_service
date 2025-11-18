#!/bin/sh
set -e

echo "🔧 Checking database migrations..."

# Database configuration
DB_HOST="${DB_HOST:-db}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-user}"
DB_NAME="${DB_NAME:-auth_service}"
DB_PASSWORD="${DB_PASSWORD:-password}"

export DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

# Wait for database
echo "⏳ Waiting for database to be ready..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"; do
  sleep 1
done
echo "✅ Database is ready!"

# Check if database is already initialized (users table exists)
if psql "$DATABASE_URL" -t -c "SELECT 1 FROM users LIMIT 1;" >/dev/null 2>&1; then
    echo "✅ Database already initialized, skipping migrations"
    exit 0
fi

echo "🔄 Applying migrations..."
if migrate -path /app/migrations -database "$DATABASE_URL" up; then
    echo "✅ Migrations completed successfully!"
else
    echo "❌ Migrations failed"
    exit 1
fi
