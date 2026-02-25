#!/bin/bash
# ZeitPass Deployment Script
# Run this locally to deploy to server

set -e

SERVER="77.42.64.36"
USER="jonathanstaude"  # Update with your SSH user
DEPLOY_PATH="/opt/zeitpass"

echo "🚀 ZeitPass Deployment Script"
echo "=============================="

# Build backend locally
echo "🔨 Building Go backend..."
cd "$(dirname "$0")/.."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o zeitpass-api ./cmd/api

echo "✅ Backend built successfully"

# Create deployment package
echo "📦 Creating deployment package..."
tar -czf zeitpass-backend.tar.gz zeitpass-api .env.example

# Upload to server
echo "📤 Uploading to server..."
scp zeitpass-backend.tar.gz ${USER}@${SERVER}:/tmp/

# Deploy on server
echo "🚀 Deploying on server..."
ssh ${USER}@${SERVER} << 'ENDSSH'
set -e

# Extract
cd /tmp
tar -xzf zeitpass-backend.tar.gz

# Stop service if running
sudo systemctl stop zeitpass || true

# Move files
sudo mkdir -p /opt/zeitpass/backend
sudo mv zeitpass-api /opt/zeitpass/backend/
sudo chown -R zeitpass:zeitpass /opt/zeitpass/backend

# Setup .env if not exists
if [ ! -f /opt/zeitpass/backend/.env ]; then
    sudo cp /tmp/.env.example /opt/zeitpass/backend/.env
    echo "⚠️  Please configure /opt/zeitpass/backend/.env"
fi

# Restart service
sudo systemctl start zeitpass
sudo systemctl status zeitpass --no-pager

echo "✅ Deployment complete!"

# Cleanup
rm /tmp/zeitpass-backend.tar.gz

ENDSSH

# Cleanup local files
rm zeitpass-backend.tar.gz

echo "✅ Deployment finished successfully!"
echo "🌐 Check https://zp.11data.ai"
