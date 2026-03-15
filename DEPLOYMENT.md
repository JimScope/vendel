# Deployment Guide

Vendel can be deployed via Docker (recommended) or as standalone binaries.

## Option 1: Docker Compose (Recommended)

Best for servers where you want a single-command deploy.

### 1. Clone and configure

```bash
git clone https://github.com/JimScope/vendel.git
cd vendel
cp .env.example .env
```

Edit `.env` with your production values:

```bash
ENVIRONMENT=production
FIRST_SUPERUSER=admin@yourdomain.com
FIRST_SUPERUSER_PASSWORD=<strong-password>
WEBHOOK_ENCRYPTION_KEY=<random-32-char-string>
APP_URL=https://yourdomain.com
FRONTEND_URL=https://yourdomain.com
```

### 2. Start

```bash
docker compose up -d
```

The app runs on port `8090`. Place a reverse proxy (Nginx, Caddy, Traefik) in front for HTTPS.

### 3. With modem agent

If the server has USB modems attached:

```bash
# Add modem config to .env
MODEMS=dk_your_api_key:/dev/ttyUSB0:/dev/ttyUSB1

docker compose --profile modem up -d
```

## Option 2: Standalone Binaries

Best for minimal setups, VPS, or ARM devices (Raspberry Pi, etc.).

### 1. Download

Go to [GitHub Releases](https://github.com/JimScope/vendel/releases) and download the archive for your platform:

**Server** (includes backend + frontend):
- `vendel_linux_amd64.tar.gz` — Linux x86_64
- `vendel_linux_arm64.tar.gz` — Linux ARM64 (Raspberry Pi 4+, Oracle Cloud)
- `vendel_linux_arm.tar.gz` — Linux ARMv7
- `vendel_darwin_amd64.tar.gz` — macOS Intel
- `vendel_darwin_arm64.tar.gz` — macOS Apple Silicon
- `vendel_windows_amd64.zip` — Windows x86_64

**Modem Agent** (optional, separate release):
- `vendel-modem-agent_linux_amd64.tar.gz`
- etc.

### 2. Extract and run the server

```bash
tar xzf vendel_linux_amd64.tar.gz
cd vendel_linux_amd64

# The archive contains:
#   vendel          <- server binary
#   pb_public/      <- frontend assets

# Set required environment variables
export WEBHOOK_ENCRYPTION_KEY="your-random-32-char-key"
export FIRST_SUPERUSER="admin@yourdomain.com"
export FIRST_SUPERUSER_PASSWORD="your-password"
export APP_URL="https://yourdomain.com"
export FRONTEND_URL="https://yourdomain.com"

./vendel serve --http=0.0.0.0:8090
```

PocketBase will create a `pb_data/` directory for the SQLite database on first run.

### 3. Run as a systemd service

Create `/etc/systemd/system/vendel.service`:

```ini
[Unit]
Description=Vendel SMS Gateway
After=network.target

[Service]
Type=simple
User=vendel
WorkingDirectory=/opt/vendel
ExecStart=/opt/vendel/vendel serve --http=0.0.0.0:8090
EnvironmentFile=/opt/vendel/.env
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now vendel
```

### 4. Run the modem agent (optional)

On the same or a different machine with USB modems:

```bash
tar xzf vendel-modem-agent_linux_amd64.tar.gz

export VENDEL_URL="https://yourdomain.com"
export MODEMS="dk_your_api_key:/dev/ttyUSB0:/dev/ttyUSB1"

./vendel-modem-agent
```

The agent connects to the server via API key authentication and SSE for real-time message dispatch.

## Reverse Proxy

Vendel runs on port `8090`. You need a reverse proxy for HTTPS.

### Caddy (simplest)

```
yourdomain.com {
    reverse_proxy localhost:8090
}
```

### Nginx

```nginx
server {
    listen 443 ssl;
    server_name yourdomain.com;

    ssl_certificate     /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # SSE support (modem agent real-time connection)
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 3600s;
    }
}
```

Note the `proxy_buffering off` — required for SSE connections used by the modem agent.

## Backups

Vendel uses SQLite. The database lives in `pb_data/`.

### Manual backup

```bash
cp -r pb_data/ pb_data_backup_$(date +%Y%m%d)
```

### Litestream (continuous replication)

Vendel supports [Litestream](https://litestream.io/) for continuous SQLite replication to S3-compatible storage. Set these variables:

```bash
LITESTREAM_REPLICA_URL=s3://my-bucket/vendel/data
LITESTREAM_ACCESS_KEY_ID=your-key
LITESTREAM_SECRET_ACCESS_KEY=your-secret
```

## Releases

Releases are managed independently from the monorepo using tag prefixes:

| Component | Tag format | Example |
|---|---|---|
| Server | `v*` | `v1.0.0` |
| Modem Agent | `modem-agent/v*` | `modem-agent/v0.3.0` |

Creating a tag triggers the corresponding GitHub Actions workflow which builds binaries for all platforms and publishes a GitHub Release.

```bash
# Release a new server version
git tag v1.0.0
git push --tags

# Release a new modem agent version
git tag modem-agent/v0.1.0
git push --tags
```
