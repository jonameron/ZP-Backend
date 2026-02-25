#!/bin/bash
# ZeitPass Server Setup Script
# Run this on your server: 77.42.64.36

set -e

echo "🚀 ZeitPass Server Setup"
echo "========================"

# Update system
echo "📦 Updating system packages..."
sudo apt update && sudo apt upgrade -y

# Install dependencies
echo "📦 Installing dependencies..."
sudo apt install -y postgresql postgresql-contrib nginx git curl build-essential

# Install Go 1.22
echo "📦 Installing Go 1.22..."
if ! command -v go &> /dev/null; then
    wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
    rm go1.22.0.linux-amd64.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
    export PATH=$PATH:/usr/local/go/bin
fi

go version

# Install Node.js (for frontend build)
echo "📦 Installing Node.js..."
if ! command -v node &> /dev/null; then
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
    sudo apt install -y nodejs
fi

node --version
npm --version

# Create zeitpass user
echo "👤 Creating zeitpass user..."
if ! id "zeitpass" &>/dev/null; then
    sudo useradd -r -m -s /bin/bash zeitpass
fi

# Create application directories
echo "📁 Creating application directories..."
sudo mkdir -p /opt/zeitpass/{backend,frontend}
sudo chown -R zeitpass:zeitpass /opt/zeitpass

# Setup PostgreSQL
echo "🗄️  Setting up PostgreSQL..."
sudo -u postgres psql -c "CREATE USER zeitpass WITH PASSWORD 'zeitpass_secure_password';" || true
sudo -u postgres psql -c "CREATE DATABASE zeitpass OWNER zeitpass;" || true
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE zeitpass TO zeitpass;" || true

# Configure PostgreSQL to accept local connections
sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = 'localhost'/g" /etc/postgresql/*/main/postgresql.conf
sudo systemctl restart postgresql

echo "✅ Server setup complete!"
echo ""
echo "Next steps:"
echo "1. Clone your repos to /opt/zeitpass/"
echo "2. Build the Go backend"
echo "3. Build the frontend"
echo "4. Configure Nginx"
echo "5. Set up the systemd service"
