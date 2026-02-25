# Deploy ZeitPass to Server (77.42.64.36)

## Quick Deployment Guide

Everything is ready! Follow these steps to deploy to your server.

---

## Prerequisites

1. GitHub repos created:
   - `SD-Ram/ZP-Backend`
   - `SD-Ram/ZP-Frontend` (already exists)
   - `SD-Ram/ZP-Onboarding` (already exists)

2. DNS: `zp.11data.ai` → `77.42.64.36`

3. SSH access to server

---

## Step 1: Create GitHub Repo for Backend

```bash
# On GitHub, create new repo: SD-Ram/ZP-Backend
# Then locally:
cd /Users/jonathanstaude/Documents/current_work/ZP-Backend
git remote add origin git@github.com:SD-Ram/ZP-Backend.git
git push -u origin main
```

## Step 2: Push Frontend Changes

```bash
cd /Users/jonathanstaude/Documents/current_work/ZP-Frontend
git push origin main

cd /Users/jonathanstaude/Documents/current_work/ZP-Onboarding
git push origin main
```

## Step 3: SSH to Server

```bash
ssh jonathanstaude@77.42.64.36
# Or via Tailscale:
ssh jonathanstaude@100.97.209.102
```

## Step 4: Run Server Setup (On Server)

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y postgresql postgresql-contrib nginx git curl build-essential

# Install Go 1.22
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Install Node.js 20
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# Create application user
sudo useradd -r -m -s /bin/bash zeitpass

# Create directories
sudo mkdir -p /opt/zeitpass/{backend,frontend}
sudo chown -R zeitpass:zeitpass /opt/zeitpass

# Setup PostgreSQL
sudo -u postgres psql << EOF
CREATE USER zeitpass WITH PASSWORD 'zeitpass_secure_password_change_me';
CREATE DATABASE zeitpass OWNER zeitpass;
GRANT ALL PRIVILEGES ON DATABASE zeitpass TO zeitpass;
\q
EOF
```

## Step 5: Clone and Build Backend (On Server)

```bash
# Clone repo
cd /opt/zeitpass/backend
sudo -u zeitpass git clone https://github.com/SD-Ram/ZP-Backend.git .

# Create .env
sudo -u zeitpass tee .env > /dev/null << 'EOF'
DATABASE_URL=postgres://zeitpass:zeitpass_secure_password_change_me@localhost:5432/zeitpass?sslmode=disable
PORT=8080
ENVIRONMENT=production
FRONTEND_URL=https://zp.11data.ai
EOF

# Build
sudo -u zeitpass /usr/local/go/bin/go build -ldflags="-s -w" -o zeitpass-api ./cmd/api
```

## Step 6: Set Up systemd Service (On Server)

```bash
sudo tee /etc/systemd/system/zeitpass.service > /dev/null << 'EOF'
[Unit]
Description=ZeitPass Go API
After=network.target postgresql.service

[Service]
Type=simple
User=zeitpass
WorkingDirectory=/opt/zeitpass/backend
EnvironmentFile=/opt/zeitpass/backend/.env
ExecStart=/opt/zeitpass/backend/zeitpass-api
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable zeitpass
sudo systemctl start zeitpass
sudo systemctl status zeitpass
```

## Step 7: Build and Deploy Frontend (On Server)

```bash
# Clone ZP-Frontend
cd /tmp
git clone https://github.com/SD-Ram/ZP-Frontend.git
cd ZP-Frontend

# Create production .env
cat > .env.production << 'EOF'
VITE_API_BASE_URL=https://zp.11data.ai/api/v1
EOF

# Build
npm install
npm run build

# Deploy
sudo rm -rf /opt/zeitpass/frontend/dist
sudo mv dist /opt/zeitpass/frontend/
sudo chown -R www-data:www-data /opt/zeitpass/frontend/dist

# Cleanup
cd .. && rm -rf ZP-Frontend
```

## Step 8: Configure Nginx (On Server)

```bash
sudo tee /etc/nginx/sites-available/zp.11data.ai > /dev/null << 'EOF'
server {
    listen 80;
    server_name zp.11data.ai;
    root /opt/zeitpass/frontend/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /health {
        proxy_pass http://localhost:8080/health;
        access_log off;
    }

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
}
EOF

sudo ln -sf /etc/nginx/sites-available/zp.11data.ai /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

## Step 9: Set Up SSL (On Server)

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d zp.11data.ai --non-interactive --agree-tos -m your@email.com
```

## Step 10: Test Deployment

```bash
# Health check
curl https://zp.11data.ai/health

# API test
curl https://zp.11data.ai/api/v1/curation?uid=test

# Visit in browser
open https://zp.11data.ai
```

---

## All-in-One Deploy Script

Save this as `deploy-all.sh` and run on the server:

```bash
#!/bin/bash
set -e

echo "🚀 ZeitPass Complete Deployment"

# 1. System setup
sudo apt update && sudo apt install -y postgresql nginx git curl build-essential

# 2. Install Go
if ! command -v go &> /dev/null; then
    wget -q https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
fi

# 3. Install Node
if ! command -v node &> /dev/null; then
    curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
    sudo apt install -y nodejs
fi

# 4. Create user
sudo useradd -r -m -s /bin/bash zeitpass || true
sudo mkdir -p /opt/zeitpass/{backend,frontend}

# 5. Setup database
sudo -u postgres psql -c "CREATE USER zeitpass WITH PASSWORD 'CHANGE_PASSWORD';" || true
sudo -u postgres psql -c "CREATE DATABASE zeitpass OWNER zeitpass;" || true

# 6. Clone and build backend
cd /opt/zeitpass/backend
sudo -u zeitpass git clone https://github.com/SD-Ram/ZP-Backend.git . || (cd /opt/zeitpass/backend && sudo -u zeitpass git pull)

cat > .env << 'EOF'
DATABASE_URL=postgres://zeitpass:CHANGE_PASSWORD@localhost:5432/zeitpass?sslmode=disable
PORT=8080
ENVIRONMENT=production
FRONTEND_URL=https://zp.11data.ai
EOF

sudo -u zeitpass /usr/local/go/bin/go build -o zeitpass-api ./cmd/api

# 7. Setup systemd
sudo cp deploy/zeitpass.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable zeitpass
sudo systemctl restart zeitpass

# 8. Build frontend
cd /tmp
rm -rf ZP-Frontend
git clone https://github.com/SD-Ram/ZP-Frontend.git
cd ZP-Frontend
echo 'VITE_API_BASE_URL=https://zp.11data.ai/api/v1' > .env.production
npm install
npm run build
sudo rm -rf /opt/zeitpass/frontend/dist
sudo mv dist /opt/zeitpass/frontend/
sudo chown -R www-data:www-data /opt/zeitpass/frontend/dist

# 9. Configure Nginx
sudo cp /opt/zeitpass/backend/deploy/nginx.conf /etc/nginx/sites-available/zp.11data.ai
sudo ln -sf /etc/nginx/sites-available/zp.11data.ai /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

# 10. SSL
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d zp.11data.ai --non-interactive --agree-tos -m admin@11data.ai

echo "✅ Deployment complete!"
echo "🌐 Visit: https://zp.11data.ai"
```

---

## Troubleshooting

### Backend won't start
```bash
sudo journalctl -u zeitpass -n 50
```

### Frontend not loading
```bash
sudo tail -f /var/log/nginx/error.log
```

### Database connection issues
```bash
sudo -u zeitpass psql zeitpass
```

---

**Ready to deploy? Just need to:**
1. Create GitHub repo for backend
2. Push all code
3. SSH to server and run the script!
