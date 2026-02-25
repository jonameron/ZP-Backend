# ZeitPass Quick Deployment Guide

**Server**: 77.42.64.36 (zp.11data.ai)

## 🚀 Fast Track Deployment (30 minutes)

### 1. Prepare Backend (Local - 5 min)

```bash
cd ZP-Backend

# Create production .env
cat > .env.production << EOF
DATABASE_URL=postgres://zeitpass:CHANGE_ME@localhost:5432/zeitpass?sslmode=disable
PORT=8080
ENVIRONMENT=production
FRONTEND_URL=https://zp.11data.ai
EOF

# Build
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o zeitpass-api ./cmd/api
```

### 2. Server Setup (Server - 10 min)

```bash
# SSH to server
ssh jonathanstaude@77.42.64.36

# Run setup script
curl -sSL https://raw.githubusercontent.com/YOUR_REPO/deploy/server-setup.sh | sudo bash

# OR manually:
sudo apt update && sudo apt install -y postgresql nginx git
# ... see DEPLOYMENT.md for full steps
```

### 3. Deploy Backend (Server - 5 min)

```bash
# Create directories
sudo mkdir -p /opt/zeitpass/backend
sudo useradd -r -m zeitpass || true

# Upload binary (from local machine)
scp zeitpass-api jonathanstaude@77.42.64.36:/tmp/
scp .env.production jonathanstaude@77.42.64.36:/tmp/.env

# Install (on server)
ssh jonathanstaude@77.42.64.36
sudo mv /tmp/zeitpass-api /opt/zeitpass/backend/
sudo mv /tmp/.env /opt/zeitpass/backend/
sudo chown -R zeitpass:zeitpass /opt/zeitpass/backend

# Set up systemd
sudo tee /etc/systemd/system/zeitpass.service > /dev/null << EOF
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

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable zeitpass
sudo systemctl start zeitpass
sudo systemctl status zeitpass
```

### 4. Deploy Frontend (Server - 5 min)

```bash
# Build locally
cd ZP-Frontend
npm run build

# Upload
scp -r dist/ jonathanstaude@77.42.64.36:/tmp/frontend-dist/

# Install (on server)
ssh jonathanstaude@77.42.64.36
sudo mkdir -p /opt/zeitpass/frontend
sudo mv /tmp/frontend-dist /opt/zeitpass/frontend/dist
sudo chown -R www-data:www-data /opt/zeitpass/frontend/dist
```

### 5. Configure Nginx (Server - 3 min)

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
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /health {
        proxy_pass http://localhost:8080/health;
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/zp.11data.ai /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 6. SSL with Let's Encrypt (Server - 2 min)

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d zp.11data.ai --non-interactive --agree-tos -m your@email.com
```

### 7. Test (2 min)

```bash
# Health check
curl https://zp.11data.ai/health

# API test
curl https://zp.11data.ai/api/v1/curation?uid=test

# Visit in browser
open https://zp.11data.ai
```

---

## ✅ Done!

Your ZeitPass app is now live at **https://zp.11data.ai**

## 📊 Quick Commands

```bash
# Backend logs
ssh jonathanstaude@77.42.64.36 "sudo journalctl -u zeitpass -f"

# Restart backend
ssh jonathanstaude@77.42.64.36 "sudo systemctl restart zeitpass"

# Nginx logs
ssh jonathanstaude@77.42.64.36 "sudo tail -f /var/log/nginx/access.log"

# Database access
ssh jonathanstaude@77.42.64.36 "sudo -u zeitpass psql zeitpass"
```

## 🔄 Update Backend

```bash
# Local
cd ZP-Backend
CGO_ENABLED=0 GOOS=linux go build -o zeitpass-api ./cmd/api
scp zeitpass-api jonathanstaude@77.42.64.36:/tmp/

# Server
ssh jonathanstaude@77.42.64.36
sudo systemctl stop zeitpass
sudo mv /tmp/zeitpass-api /opt/zeitpass/backend/
sudo systemctl start zeitpass
```

## 🔄 Update Frontend

```bash
# Local
cd ZP-Frontend
npm run build
scp -r dist/ jonathanstaude@77.42.64.36:/tmp/frontend-dist/

# Server
ssh jonathanstaude@77.42.64.36
sudo rm -rf /opt/zeitpass/frontend/dist
sudo mv /tmp/frontend-dist /opt/zeitpass/frontend/dist
sudo chown -R www-data:www-data /opt/zeitpass/frontend/dist
```

---

For detailed instructions, see **DEPLOYMENT.md**
