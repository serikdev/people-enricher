#!/bin/bash

set -e

echo "Running DB migrations...."

source .env
goose -dir ./migrations postgres "$DATABASE_URL" up

echo "Migrations completed successfully"