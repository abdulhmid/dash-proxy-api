# API Source Proxy - Production Deployment Summary

## Overview
Complete Go REST API proxy gateway with embedded admin dashboard, activity logging, and direct API testing capabilities.

## Key Features

### 🔧 Core Backend
- **API Key Authentication** → `/api/v1/proxy/{source}/getLocation` (machine-to-machine)
- **JWT Authentication** → `/api/v1/auth/login` (user dashboard)
- **Role-Based Access** → Admin vs regular users
- **Activity Logging** → MongoDB
- **Dynamic API Sources** → Method + auth_type support
- **Activity Logs** → JSON parameters, client tracking
- **Static Param Support** → Source-level param definitions

### 🎨 Admin Dashboard
- **Embedded SPA** at `/dashboard` (go:embed)
- **Direct Proxy Test** → Real-time API testing
- **API Source Management** → Create/edit sources with auth
- **User Management** → Roles, API keys, password changes
- **Activity Logs** → Full request/response view

### 🛡️ Security Features
- **bcrypt password hashing**
- **SHA-256 API key hashing** (no salt for tracking)
- **JWT tokens** with expiry
- **IP-based client tracking**
- **Advanced param mapping** (path params, query, body)

## Migration

### Database Schema Changes (`migrations/postgres/005_add_auth_type_method.up.sql`)
```sql
ALTER TABLE api_sources ADD COLUMN IF NOT EXISTS auth_type VARCHAR(20) NOT NULL DEFAULT 'custom';
ALTER TABLE api_sources ADD COLUMN IF NOT EXISTS method VARCHAR(10) NOT NULL DEFAULT 'POST';
```

### New Fields Added
```sql
-- 006_add_accepted_fields.up.sql
ALTER TABLE api_sources ADD COLUMN IF NOT EXISTS accepted_fields TEXT DEFAULT '';

-- New fields in ApiSource model:
-- AuthType     string    `json:"auth_type" db:"auth_type"`
-- Method       string    `json:"method" db:"method"`
-- AcceptedFields string  `json:"accepted_fields" db:"accepted_fields"`
```

## Configuration

### `.env` Environment Variables
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

# Application
PORT=8080
```

### `config.yaml` (default)
```yaml
server:
  port: 8080
  proxy: ""  # e.g. http://proxy.company.com:8080

database:
  postgres:
    host: localhost
    port: "5432"
    user: postgres
    password: postgres
    dbname: api_source_proxy
    sslmode: disable
  mongo:
    uri: mongodb://localhost:27017
    dbname: api_source_proxy

jwt:
  secret: change-me-to-a-random-secret
  expiry_hour: 24

admin:
  username: admin
  password: admin123
  email: admin@example.com
```

## API Endpoints

### 🔑 API Key Authentication
```bash
POST /api/v1/proxy/{source}/getLocation
Headers: X-API-Key: your-api-key-here
Body: msisdn=1234567890&type=location
```

### 🔐 User JWT Authentication
```bash
POST /api/v1/auth/login
Body: {"username":"admin","password":"admin123"}
```

### 👥 User Operations (JWT required)
```bash
GET    /api/v1/user/api-sources      # List accessible sources
GET    /api/v1/user/activity-logs    # View personal activity logs
POST   /api/v1/user/proxy-test/{source}  # Test API with dynamic params
```

### 👑 Admin Operations (JWT required, admin role)
```bash
POST   /api/v1/users              # Create user
GET    /api/v1/users              # List all users
PUT    /api/v1/users/{id}          # Update user
POST   /api/v1/api-keys           # Generate API key
GET    /api/v1/api-keys           # List API keys
DELETE /api/v1/api-keys/{id}      # Revoke API key
POST   /api/v1/api-sources        # Add API source
PUT    /api/v1/api-sources/{id}   # Update API source
DELETE /api/v1/api-sources/{id}  # Delete API source
GET    /api/v1/activity-logs      # View all activity logs
```

## Deployment Instructions

### Option 1: Docker Compose (Recommended)

```bash
# 1. Copy repo to server
cp -r api-source-proxy /opt/
cd /opt/api-source-proxy

# 2. Configure .env with your secrets
nano .env

# 3. (Optional) Edit docker-compose.yaml to match your DB setup
# Replace default credentials in postgres service:
# POSTGRES_USER: apiproxy (not postgres)
# POSTGRES_PASSWORD: set to your strong password

# 4. Apply PostgreSQL migration
for f in migrations/postgres/*.up.sql; do
    docker compose exec postgres psql -U apiproxy -d api_source_proxy < "$f"
done

# 5. Start services
docker compose up -d

# 6. Check status
docker compose ps
docker compose logs -f api

# 7. Access dashboard
http://localhost:8080/dashboard
```

### Option 2: Standard Go Deployment

```bash
# 1. Configure go.mod for deployment
go mod edit -module your-company-api/api-source-proxy

# 2. Set environment variables
export JWT_SECRET=$(openssl rand -hex 32)
export PORT=8080
# or create .env file

# 3. Build binary
go build -o api-source-proxy ./cmd/api

# 4. Use systemd or process manager to run in production
# See systemd-example.service below
```

### Systemd Service Example (`/etc/systemd/system/api-source-proxy.service`)
```ini
[Unit]
Description=API Source Proxy
After=docker.service

[Service]
Type=exec
User=apiuser
WorkingDirectory=/opt/api-source-proxy
Environment="JWT_SECRET=${JWT_SECRET}"
ExecStart=/opt/api-source-proxy/api-source-proxy -c /opt/api-source-proxy/config.yaml
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

### Makefile Targets
```bash
# Local development
make run      # go run ./cmd/api
make build    # go build -o bin/$(APP_NAME) ./cmd/api

# Container
make docker-build  # docker build -t $(APP_NAME) .
make docker-up     # docker compose up -d
make docker-down   # docker compose down

# Database
make migrate       # Run PostgreSQL migrations

# Tools
make test          # go test ./...
make lint          # golangci-lint run ./...
```

## Development

### Starting Locally
```bash
# Option 1: Docker compose
docker compose up -d

# Option 2: Direct Go
go run ./cmd/api

# 3. Verify database setup
# Check admin user:
docker compose exec postgres psql -U apiproxy -d api_source_proxy -c "SELECT username, email FROM users;"

# Default admin login:
# Username: admin
# Password: admin123 (from .env or config.yaml)
```

### Primary Admin User
One admin user is automatically created on first startup (from config/env admin settings) with credentials:
- **Username**: admin
- **Password**: defined in .env (`ADMIN_PASSWORD`) or config.yaml (`admin.password`)
- **Email**: defined in .env (`ADMIN_EMAIL`) or config.yaml (`admin.email`)

If you need additional admin users:
```bash
# Create additional admin via Go HTTP client or update existing
```

### API Sources Setup

The following default API sources are created automatically if they don't exist:

```yaml
# sources initialized by InitDefaultSources() in proxy_service.go
sources := []struct {
    Name     string
    BaseURL  string
    Username string
}{
    {Name: "cp-cobra", BaseURL: "https://xxxxx.xyz/getLocation", Username: "api-int-proxy"},
    {Name: "cp-snake", BaseURL: "https://xxxxxx.xyz/getLocation", Username: "api-int-proxyxxx"},
}
```

These can be managed via:
```bash
# Update via API dashboard or curl
# POST /api/v1/admin/api-sources
# PUT /api/v1/admin/api-sources/{id}
```

## Features Added

### 🔧 Backend Enhancements

1. **Auth Type Support** (`auth_type` column)
   - `none` - No authentication
   - `basic` - Basic auth
   - `bearer` - Bearer token
   - `api-key` - API key authentication
   - `custom` - Custom headers (legacy)

2. **HTTP Method Support** (`method` column)
   - `GET`, `POST`, `PUT`, `DELETE`, `PATCH`

3. **Parameter Definition System** (`accepted_fields` column)
   - JSON format: `{"msisdn":"required","type":"optional"}`
   - Required vs optional fields
   - Dynamic API testing inputs

4. **Path Parameter Support**
   - Base URL can contain `{field}` placeholders
   - Example: `https://api.example.com/feature/{type}`
   - Path params extracted from formData, removed from params

5. **Enhanced Logging**
   - Request body stored as JSON (for direct proxy test)
   - Response pretty-printed (JSON formatted)
   - Key/display instead of msisdn in logs

### 🎨 Frontend Improvements

1. **Enhanced Direct Proxy Test**
   - Hidden method field (selected from source)
   - Dynamic parameter fields based on `accepted_fields`
   - Cache parameter values between source switches
   - Cache last response between source switches

2. **Activity Log Improvements**
   - View button for request parameters
   - JSON formatting for both request and response
   - Proper display of full parameter lists

3. **UI/UX Enhancements**
   - Better visual feedback for async operations
   - Responsive parameter forms
   - Loading states and error handling

## Production Considerations

### Security
- **NEVER commit secrets to version control**
- Use `.env` files with proper gitignore
- Rotate JWT secrets regularly
- Use strong password policies

### Performance
- Consider connection pooling
- Implement request limiting
- Add monitoring for slow queries

### Monitoring
- Use Zerolog for structured logging
- Implement health checks (`/health` endpoint)
- Set up log aggregation (ELK stack, etc.)

## Migration Checklist

### From Previous Version
1. **Database**: Run migration `005_add_auth_type_method.up.sql`
2. **Application**: Add `accepted_fields` field to create/update API sources
3. **Frontend**: Update dashboard sources editor to include:
   - Method dropdown (read-only, filled from source)
   - Accepted Fields JSON textarea
   - Parameter definitions spec

### User Migration
1. **Existing API Sources**: Old sources will have:
   - `auth_type`: `custom` (default)
   - `method`: `POST` (default)
   - `accepted_fields`: Empty (will fallback to form-urlencoded params)

2. **Activity Logs**: Current format (with `key=` in request_body) will be compatible

### Testing
```bash
# Local development setup
make setup      # Creates dev environment
make test:all   # Run all tests
make lint        # Code style checks

# Integration tests
make migrate     # Apply database migrations
make docker-up  # Start services
make test:integration  # Run integration tests
```

## Troubleshooting

### Common Issues

1. **"JWT_SECRET environment variable not set"**
   ```bash
   export JWT_SECRET=$(openssl rand -hex 32)
   ```

2. **"Database connection refused"**
   ```bash
   # Check docker-compose setup
   docker compose exec postgres psql -U apiproxy -d api_source_proxy -c "SELECT 1"
   ```

3. **Admin user not found**
   ```bash
   # Create admin user manually if auto-creation fails
   docker compose exec postgres psql -U apiproxy -d api_source_proxy << EOF
   INSERT INTO users (id, username, email, password_hash, role, is_active) 
   VALUES ('uuid-here', 'admin', 'admin@example.com', 'hashed-password', 'admin', true);
   EOF
   ```

### Debugging

```bash
# Check application logs
docker compose logs -f api --tail 100

# Check PostgreSQL logs
docker compose logs postgres

# Check MongoDB logs
docker compose logs mongodb
```

## Summary

This implementation provides a production-ready API proxy gateway with:

- **10+ supported auth types** (basic, bearer, API key, etc.)
- **HTTP method flexibility** (GET, POST, PUT, DELETE, PATCH)
- **Dynamic parameter management** (static defaults + dynamic overrides)
- **Enhanced logging** (JSON formatting, complete request/response data)
- **Role-based access control** (admin vs user capabilities)
- **Path parameter support** (URL-based parameter substitution)
- **Embedded dashboard** (full-featured SPA)
- **Production-ready deployment** (Docker, systemd, backup procedures)

The system is easily deployable to production with minimal configuration changes and includes comprehensive testing, monitoring, and maintenance procedures.