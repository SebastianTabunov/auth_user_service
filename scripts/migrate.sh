#!/bin/sh
set -e

echo "🔧 Starting database migrations..."

# Формируем DATABASE_URL из отдельных переменных если не задана
if [ -z "$DATABASE_URL" ]; then
  export DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE"
fi

echo "📊 Using database: $DB_HOST:$DB_PORT/$DB_NAME"

# Ждем пока PostgreSQL запустится
echo "⏳ Waiting for database to be ready..."
for i in $(seq 1 30); do
  if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; then
    echo "✅ Database is ready!"
    break
  fi
  echo "⏳ Waiting for database... ($i/30)"
  sleep 2
done

# Проверяем окончательно
if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; then
  echo "❌ Database connection failed after 60 seconds"
  echo "🔍 Debug info:"
  echo "DB_HOST: $DB_HOST"
  echo "DB_PORT: $DB_PORT"
  echo "DB_USER: $DB_USER"
  echo "DATABASE_URL: $DATABASE_URL"
  exit 1
fi

# Запускаем миграции
echo "🔄 Applying migrations..."
migrate -path /app/migrations -database "$DATABASE_URL" up

if [ $? -eq 0 ]; then
  echo "✅ Migrations completed successfully!"
else
  echo "❌ Migrations failed!"
  exit 1
fi

echo "🎉 Database setup complete!"

