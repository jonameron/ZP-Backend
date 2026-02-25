#!/bin/bash
# ZeitPass Deployment Script for server 100.97.209.102 / 77.42.64.36
# Domain: zp.11data.ai
# Port: 8081 (to avoid conflict with courtofagents on 8080)
# Run this script on the server with sudo access

set -e

echo "=== ZeitPass Deployment Script ==="
echo "This will deploy ZeitPass backend to this server"
echo ""

# Step 1: Install PostgreSQL
echo ">>> Step 1: Installing PostgreSQL..."
sudo apt update
sudo apt install -y postgresql postgresql-contrib

# Step 2: Start PostgreSQL
echo ">>> Step 2: Starting PostgreSQL..."
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Step 3: Create database user and database
echo ">>> Step 3: Setting up database..."
sudo -u postgres psql << 'EOF'
CREATE USER zeitpass WITH PASSWORD 'zeitpass_prod_2024!';
CREATE DATABASE zeitpass OWNER zeitpass;
GRANT ALL PRIVILEGES ON DATABASE zeitpass TO zeitpass;
\q
EOF

# Step 4: Create zeitpass system user
echo ">>> Step 4: Creating zeitpass system user..."
sudo useradd -r -m -s /bin/bash zeitpass 2>/dev/null || echo "User zeitpass already exists"

# Step 5: Create directories
echo ">>> Step 5: Creating directories..."
sudo mkdir -p /opt/zeitpass/backend
sudo mkdir -p /opt/zeitpass/frontend/dist
sudo mkdir -p /opt/zeitpass/backend/logs
sudo chown -R zeitpass:zeitpass /opt/zeitpass

# Step 6: Clone backend repo
echo ">>> Step 6: Cloning backend repository..."
cd /opt/zeitpass/backend
sudo -u zeitpass git clone https://github.com/jonameron/ZP-Backend.git . 2>/dev/null || (cd /opt/zeitpass/backend && sudo -u zeitpass git pull)

# Step 7: Create .env file
echo ">>> Step 7: Creating .env file..."
sudo tee /opt/zeitpass/backend/.env > /dev/null << 'EOF'
DATABASE_URL=postgres://zeitpass:zeitpass_prod_2024!@localhost:5432/zeitpass?sslmode=disable
PORT=8081
ENVIRONMENT=production
FRONTEND_URL=https://zp.11data.ai
EOF
sudo chown zeitpass:zeitpass /opt/zeitpass/backend/.env
sudo chmod 600 /opt/zeitpass/backend/.env

# Step 8: Build backend
echo ">>> Step 8: Building backend..."
cd /opt/zeitpass/backend
export PATH=$HOME/.local/go/bin:$PATH
sudo -u zeitpass bash -c 'export PATH=/home/jon/.local/go/bin:$PATH && cd /opt/zeitpass/backend && go build -ldflags="-s -w" -o zeitpass-api ./cmd/api'

# Step 9: Create systemd service
echo ">>> Step 9: Creating systemd service..."
sudo tee /etc/systemd/system/zeitpass.service > /dev/null << 'EOF'
[Unit]
Description=ZeitPass Go API Server
After=network.target postgresql.service
Wants=postgresql.service

[Service]
Type=simple
User=zeitpass
Group=zeitpass
WorkingDirectory=/opt/zeitpass/backend
Environment="PATH=/usr/local/bin:/usr/bin:/bin"
EnvironmentFile=/opt/zeitpass/backend/.env
ExecStart=/opt/zeitpass/backend/zeitpass-api
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=zeitpass-api

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/zeitpass/backend/logs

# Resource limits
LimitNOFILE=65535
LimitNPROC=4096

[Install]
WantedBy=multi-user.target
EOF

# Step 10: Start zeitpass service
echo ">>> Step 10: Starting zeitpass service..."
sudo systemctl daemon-reload
sudo systemctl enable zeitpass
sudo systemctl restart zeitpass
sleep 2
sudo systemctl status zeitpass --no-pager || true

# Step 11: Create nginx config for zp.11data.ai
echo ">>> Step 11: Creating nginx config..."
sudo tee /etc/nginx/sites-available/zp.11data.ai > /dev/null << 'EOF'
# ZeitPass API (zp.11data.ai)
# Port 8081 to avoid conflict with courtofagents on 8080

server {
    listen 80;
    listen [::]:80;
    server_name zp.11data.ai;

    # Serve frontend static files
    root /opt/zeitpass/frontend/dist;
    index index.html;

    # Frontend - SPA routing
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Backend API proxy (port 8081)
    location /api/ {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Health check endpoint
    location /health {
        proxy_pass http://localhost:8081/health;
        access_log off;
    }

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1000;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json application/javascript;
}
EOF

# Step 12: Enable site and reload nginx
echo ">>> Step 12: Enabling nginx site..."
sudo ln -sf /etc/nginx/sites-available/zp.11data.ai /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx

# Step 13: Test the deployment
echo ">>> Step 13: Testing deployment..."
sleep 2
curl -s http://localhost:8081/health || echo "Health check failed - checking logs..."
sudo journalctl -u zeitpass -n 20 --no-pager || true

echo ""
echo "=== Deployment Complete ==="
echo "Backend running on port 8081"
echo "Nginx configured for zp.11data.ai"
echo ""
echo "Next steps:"
echo "1. Test: curl http://localhost:8081/health"
echo "2. SSL: sudo certbot --nginx -d zp.11data.ai"
echo "3. Deploy frontend to /opt/zeitpass/frontend/dist"
echo ""
