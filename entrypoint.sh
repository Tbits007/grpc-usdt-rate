#!/bin/bash

set -ex

echo "Waiting for PostgreSQL to become ready..."
while ! pg_isready -h postgres -p 5432 -U postgres; do
  sleep 2
done

echo "Checking migrations status..."
goose -dir /app/migrations postgres "$POSTGRES_DSN" status

echo "Applying migrations..."
goose -dir /app/migrations postgres "$POSTGRES_DSN" up

echo "Starting application..."
exec ./app