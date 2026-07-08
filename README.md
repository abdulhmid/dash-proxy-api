# API Source Proxy

Reverse proxy gateway with admin dashboard, API key auth, and activity logging.

## Architecture

- **API** — Go (Chi router, embedded SPA dashboard)
- **PostgreSQL** — Users, API keys, API sources config
- **MongoDB** — Activity logs

## Production Deployment

### Prerequisites

- Docker & Docker Compose v2
- Server with Docker installed (tested on Ubuntu 22.04+)

### 1. Clone & Prepare

```bash
git clone <repo-url> /opt/api-source-proxy
cd /opt/api-source-proxy
cp .env.example .env
```

### 2. Configure `.env`

```ini
# Security (CHANGE THESE!)
JWT_SECRET=<generate-random-64-char-string>
ADMIN_PASSWORD=<strong-admin-password>
ADMIN_EMAIL=admin@yourcompany.com

# Optional: behind corporate proxy
SERVER_PROXY=http://proxy.yourcompany.com:8080

# Database credentials (change for production)
DB_POSTGRES_PASSWORD=<strong-db-password>
DB_MONGO_URI=mongodb://mongodb:27017
```

Generate secrets:
```bash
# JWT secret
openssl rand -hex 32
```

### 3. Update `docker-compose.yaml` for Production

Replace the default Postgres credentials in the `postgres` service:

```yaml
postgres:
  environment:
    POSTGRES_DB: api_source_proxy
    POSTGRES_USER: apiproxy
    POSTGRES_PASSWORD: ${DB_POSTGRES_PASSWORD}
```

And update the `api` service environment:

```yaml
api:
  environment:
    DATABASE_POSTGRES_USER: apiproxy
    DATABASE_POSTGRES_PASSWORD: ${DB_POSTGRES_PASSWORD}
    JWT_SECRET: ${JWT_SECRET}
    ADMIN_PASSWORD: ${ADMIN_PASSWORD}
    ADMIN_EMAIL: ${ADMIN_EMAIL}
    SERVER_PROXY: ${SERVER_PROXY}
```

### 4. Deploy

```bash
# Start all services
docker compose up -d

# Check logs
docker compose logs -f api

# Verify health
curl http://localhost:8080/dashboard
# Expected: 200 (HTML page)
```

### 5. Apply Database Migrations

The `postgres` container auto-runs scripts from `migrations/postgres/` on first start. If deploying to existing DB:

```bash
docker compose exec postgres psql -U apiproxy -d api_source_proxy \
  -f /docker-entrypoint-initdb.d/001_create_users.up.sql

# Run remaining migrations manually via psql:
for f in migrations/postgres/*.up.sql; do
  echo "Running $f..."
  docker compose exec -T postgres psql -U apiproxy -d api_source_proxy < "$f"
done
```

### 6. NGINX Reverse Proxy (Recommended)

```nginx
server {
    listen 443 ssl;
    server_name api-proxy.yourcompany.com;

    ssl_certificate     /etc/ssl/certs/your.crt;
    ssl_certificate_key /etc/ssl/private/your.key;

    client_max_body_size 10M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support (if needed)
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### 7. Systemd (Auto-restart)

```ini
# /etc/systemd/system/docker-compose-api-proxy.service
[Unit]
Description=API Source Proxy
Requires=docker.service
After=docker.service

[Service]
WorkingDirectory=/opt/api-source-proxy
ExecStart=/usr/bin/docker compose up
ExecStop=/usr/bin/docker compose down
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
systemctl daemon-reload
systemctl enable --now docker-compose-api-proxy
```

### 8. Verify

| Endpoint | Description |
|----------|-------------|
| `https://api-proxy.yourcompany.com/dashboard` | Admin dashboard |
| `https://api-proxy.yourcompany.com/api/v1/auth/login` | Login API |

Default admin login: username from config (`admin`), password from `.env` (`ADMIN_PASSWORD`).

## Backup

```bash
# PostgreSQL
docker compose exec postgres pg_dump -U apiproxy api_source_proxy > backup_$(date +%Y%m%d).sql

# MongoDB
docker compose exec mongodb mongodump --out /tmp/backup
docker compose cp mongodb:/tmp/backup ./mongo_backup_$(date +%Y%m%d)
```

## Updating

```bash
cd /opt/api-source-proxy
git pull
docker compose up --build -d api
# Apply any new migrations (step 5)
```

## Troubleshooting

- **API crashes with "Resource already exists"**: The admin user already exists. Run `docker compose exec postgres psql -U apiproxy -d api_source_proxy -c "SELECT username, email FROM users;"` and ensure the configured admin username/email isn't conflicting with an existing user.
- **MongoDB connection refused**: Ensure `mongodb` service is healthy (`docker compose ps`).
- **Postgres connection refused**: Check `DATABASE_POSTGRES_HOST` — must be `postgres` (Docker service name), not `localhost`.
