#!/bin/bash

set -e

echo "Running DB migrations...."
goose -dir ./migrations postgres "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}" up

echo "Migrations completed successfully"