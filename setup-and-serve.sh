#!/bin/bash
set -e  # Exit if any step fails

echo "🚀 Running init..."
docker-compose run --rm init

echo "⚙️ Starting config server..."
docker-compose up config  # run in background

echo "🛑 Stopping config server..."
docker-compose stop config

echo "📥 Importing data..."
docker-compose run --rm import

echo "🌐 Starting main server..."
docker-compose up server