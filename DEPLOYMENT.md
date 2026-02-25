# ZeitPass Deployment Guide

Complete guide to deploy ZeitPass to production server at `77.42.64.36` / `zp.11data.ai`

## Prerequisites

- SSH access to server (77.42.64.36)
- Domain `zp.11data.ai` pointing to server IP
- Git repositories for ZP-Frontend and ZP-Backend

---

## Step 1: Server Initial Setup

SSH into your server:

```bash
ssh jonathanstaude@77.42.64.36
# Or via Tailscale:
ssh jonathanstaude@100.97.209.102
```

Run the setup script:

```bash
# Upload setup script
scp deploy/server-setup.sh jonathanstaude@77.42.64.36:/tmp/
ssh jonathanstaude@77.42.64.36

# Run setup
chmod +x /tmp/server-setup.sh
sudo /tmp/server-setup.sh
```

This will:
- Install PostgreSQL, Nginx, Go, Node.js
- Create zeitpass user
- Set up database
- Create application directories

---

## Step 2: Configure Database

Still on the server, set a secure database password:

```bash
# Generate a secure password
openssl rand -base64 32

# Update PostgreSQL password
sudo -u postgres psql
ALTER USER zeitpass WITH PASSWORD 'YOUR_SECURE_PASSWORD_HERE';
\q
```

---

## Step 3: Deploy Backend

### Option A: Using Deploy Script (Recommended)

From your local machine:

```bash
cd ZP-Backend

# Update .env.example with production values
cat > .env.production << EOF
DATABASE_URL=postgres://zeitpass:YOUR_SECURE_PASSWORD@localhost:5432/zeitpass?sslmode=disable
PORT=8080
ENVIRONMENT=production
FRONTEND_URL=https://zp.11data.ai
EOF

# Deploy
./deploy/deploy.sh
```

### Option B: Manual Deployment

```bash
# Build locally
cd ZP-Backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o zeitpass-api ./cmd/api

# Upload
scp zeitpass-api jonathanstaude@77.42.64.36:/tmp/
scp .env.production jonathanstaude@77.42.64.36:/tmp/.env

# Deploy on server
ssh jonathanstaude@77.42.64.36
sudo mkdir -p /opt/zeitpass/backend
sudo mv /tmp/zeitpass-api /opt/zeitpass/backend/
sudo mv /tmp/.env /opt/zeitpass/backend/
sudo chown -R zeitpass:zeitpass /opt/zeitpass/backend
```

---

## Step 4: Set Up Systemd Service

On the server:

```bash
# Copy systemd service file
sudo cp /opt/zeitpass/backend/deploy/zeitpass.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload

# Enable and start service
sudo systemctl enable zeitpass
sudo systemctl start zeitpass

# Check status
sudo systemctl status zeitpass

# View logs
sudo journalctl -u zeitpass -f
```

---

## Step 5: Deploy Frontend

### Build Frontend Locally

```bash
cd ZP-Frontend  # or consolidated app

# Create production .env
cat > .env.production << EOF
VITE_API_URL=https://zp.11data.ai/api/v1
EOF

# Build
npm install
npm run build

# Upload
scp -r dist/ jonathanstaude@77.42.64.36:/tmp/frontend-dist/
```

### Deploy on Server

```bash
ssh jonathanstaude@77.42.64.36

sudo mkdir -p /opt/zeitpass/frontend
sudo rm -rf /opt/zeitpass/frontend/dist
sudo mv /tmp/frontend-dist /opt/zeitpass/frontend/dist
sudo chown -R www-data:www-data /opt/zeitpass/frontend/dist
```

---

## Step 6: Configure Nginx

On the server:

```bash
# Copy Nginx config
sudo cp /opt/zeitpass/backend/deploy/nginx.conf /etc/nginx/sites-available/zp.11data.ai

# Enable site
sudo ln -s /etc/nginx/sites-available/zp.11data.ai /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

Test HTTP access:

```bash
curl http://zp.11data.ai/health
```

---

## Step 7: Set Up SSL with Let's Encrypt

```bash
# Install certbot
sudo apt install -y certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d zp.11data.ai

# Certbot will automatically:
# - Generate SSL certificates
# - Update Nginx configuration
# - Set up auto-renewal

# Test auto-renewal
sudo certbot renew --dry-run
```

---

## Step 8: Migrate Data from Airtable

### Export from Airtable

1. Log into your Airtable base
2. Go to each table (Users, Events, Booking, ActivityLog)
3. Export as CSV

### Import to PostgreSQL

On the server:

```bash
# Connect to database
sudo -u zeitpass psql zeitpass

# Create initial admin user (example)
INSERT INTO users (user_id, first_name, email, plan, total_events, guest_access)
VALUES ('SivRam618', 'Sivarama', 'siv@example.com', 'Social', 5, true);

# Import events (example - adjust for your CSV)
\copy events(event_id, title, vendor, category, status, ...) FROM '/path/to/events.csv' DELIMITER ',' CSV HEADER;

# Exit
\q
```

Or create a Go migration script to import from Airtable API directly.

---

## Step 9: Verification Checklist

- [ ] Backend health check: `curl https://zp.11data.ai/health`
- [ ] API responds: `curl https://zp.11data.ai/api/v1/curation?uid=TEST`
- [ ] Frontend loads: Visit `https://zp.11data.ai`
- [ ] SSL certificate valid (green lock in browser)
- [ ] Systemd service running: `sudo systemctl status zeitpass`
- [ ] Database connected: Check logs for "Database connected"
- [ ] Nginx access logs: `sudo tail -f /var/log/nginx/access.log`

---

## Maintenance Commands

### Backend

```bash
# Restart API
sudo systemctl restart zeitpass

# View logs
sudo journalctl -u zeitpass -f

# Stop/Start
sudo systemctl stop zeitpass
sudo systemctl start zeitpass
```

### Frontend

```bash
# Deploy new version
scp -r dist/ jonathanstaude@77.42.64.36:/tmp/frontend-dist/
ssh jonathanstaude@77.42.64.36
sudo rm -rf /opt/zeitpass/frontend/dist
sudo mv /tmp/frontend-dist /opt/zeitpass/frontend/dist
sudo chown -R www-data:www-data /opt/zeitpass/frontend/dist
```

### Database

```bash
# Backup
sudo -u postgres pg_dump zeitpass > zeitpass_backup_$(date +%Y%m%d).sql

# Restore
sudo -u postgres psql zeitpass < zeitpass_backup.sql

# Connect
sudo -u zeitpass psql zeitpass
```

### Nginx

```bash
# Test config
sudo nginx -t

# Reload
sudo systemctl reload nginx

# Restart
sudo systemctl restart nginx

# Logs
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

---

## Troubleshooting

### Backend won't start

```bash
# Check logs
sudo journalctl -u zeitpass -n 50

# Common issues:
# 1. Database connection - check .env DATABASE_URL
# 2. Port already in use - check if another process uses port 8080
# 3. Permissions - ensure zeitpass user owns files
```

### Frontend not loading

```bash
# Check Nginx error log
sudo tail -f /var/log/nginx/error.log

# Verify files exist
ls -la /opt/zeitpass/frontend/dist/

# Check permissions
sudo chown -R www-data:www-data /opt/zeitpass/frontend/dist
```

### Database connection issues

```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Test connection
psql -U zeitpass -d zeitpass -h localhost

# Check pg_hba.conf allows local connections
sudo cat /etc/postgresql/*/main/pg_hba.conf
```

---

## Monitoring

### Set Up Monitoring (Optional)

```bash
# Install Prometheus Node Exporter
wget https://github.com/prometheus/node_exporter/releases/download/v1.7.0/node_exporter-1.7.0.linux-amd64.tar.gz
tar xvfz node_exporter-1.7.0.linux-amd64.tar.gz
sudo mv node_exporter-1.7.0.linux-amd64/node_exporter /usr/local/bin/
sudo useradd -rs /bin/false node_exporter

# Create systemd service for node_exporter
# Monitor at http://77.42.64.36:9100/metrics
```

### Simple Health Check Script

```bash
#!/bin/bash
# /opt/zeitpass/health-check.sh

curl -f https://zp.11data.ai/health || \
  echo "ZeitPass health check failed!" | mail -s "ZeitPass Alert" your@email.com
```

Add to crontab:

```bash
crontab -e
# Run every 5 minutes
*/5 * * * * /opt/zeitpass/health-check.sh
```

---

## Rollback Procedure

If deployment fails:

```bash
# Backend
sudo systemctl stop zeitpass
# Restore previous binary from backup
sudo cp /opt/zeitpass/backend/zeitpass-api.backup /opt/zeitpass/backend/zeitpass-api
sudo systemctl start zeitpass

# Frontend
sudo rm -rf /opt/zeitpass/frontend/dist
# Restore from backup
sudo cp -r /opt/zeitpass/frontend/dist.backup /opt/zeitpass/frontend/dist

# Database
sudo -u postgres psql zeitpass < /path/to/backup.sql
```

---

## Security Hardening (Recommended)

```bash
# Set up firewall
sudo ufw allow 22/tcp   # SSH
sudo ufw allow 80/tcp   # HTTP
sudo ufw allow 443/tcp  # HTTPS
sudo ufw enable

# Fail2ban for SSH protection
sudo apt install fail2ban
sudo systemctl enable fail2ban

# Regular updates
sudo apt update && sudo apt upgrade -y

# Database: Restrict to localhost only
# Edit postgresql.conf: listen_addresses = 'localhost'
```

---

## Performance Tuning

### PostgreSQL

```sql
-- /etc/postgresql/*/main/postgresql.conf
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
work_mem = 4MB
min_wal_size = 1GB
max_wal_size = 4GB
```

### Nginx

```nginx
# /etc/nginx/nginx.conf
worker_processes auto;
worker_connections 1024;
client_max_body_size 10M;
```

---

## Contact

For issues, contact: jonathan@11data.ai

**Server Access**:
- IP: 77.42.64.36
- Tailscale: 100.97.209.102
- Domain: zp.11data.ai
