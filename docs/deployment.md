# Vendel - Deployment

Vendel can be deployed via Docker Compose (recommended) or as standalone binaries.

## Option 1: Docker Compose (Recommended)

### Preparation

* Have a remote server available.
* Configure your domain's DNS records to point to the server's IP.
* Install and configure [Docker](https://docs.docker.com/engine/install/) on the remote server (Docker Engine, not Docker Desktop).

### Environment Variables

```bash
cp .env.example .env
```

Set the required variables:

```bash
ENVIRONMENT=production
FIRST_SUPERUSER=admin@yourdomain.com
FIRST_SUPERUSER_PASSWORD=<strong-password>
WEBHOOK_ENCRYPTION_KEY=<random-32-char-string>
APP_URL=https://yourdomain.com
FRONTEND_URL=https://yourdomain.com
```

See `docs/development.md` for the full list of environment variables.

### Deploy

```bash
docker compose up -d
```

The Docker image uses a multi-stage build that compiles the frontend and Go backend into a single binary on Alpine (~50MB). No external database required — SQLite is embedded in PocketBase.

The `pb_data` volume persists the SQLite database across restarts.

### With modem agent

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

Note: `proxy_buffering off` is required for SSE connections used by the modem agent.

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

## Continuous Deployment (CD)

### GitHub Actions Runner

1. On your remote server, create a user for GitHub Actions:

```bash
sudo adduser github
sudo usermod -aG docker github
```

2. Switch to the `github` user and [install a GitHub Action self-hosted runner](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/adding-self-hosted-runners#adding-a-self-hosted-runner-to-a-repository).

3. When asked for labels, add a label for the environment (e.g., `production`).

4. Install it as a service:

```bash
cd /home/github/actions-runner
sudo ./svc.sh install github
sudo ./svc.sh start
```

### Configure Secrets

In your GitHub repository, configure the secrets expected by the workflows:

- `DOMAIN_PRODUCTION` / `DOMAIN_STAGING`
- `STACK_NAME_PRODUCTION` / `STACK_NAME_STAGING`
- `FIRST_SUPERUSER`, `FIRST_SUPERUSER_PASSWORD`
- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`
- `APP_URL_PRODUCTION` / `APP_URL_STAGING`
- `FRONTEND_URL_PRODUCTION` / `FRONTEND_URL_STAGING`
- `FIREBASE_SERVICE_ACCOUNT_JSON`
- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`
- `GH_OAUTH_CLIENT_ID`, `GH_OAUTH_CLIENT_SECRET`
- `QVAPAY_APP_ID`, `QVAPAY_APP_SECRET`
- `WEBHOOK_ENCRYPTION_KEY`
- `LITESTREAM_REPLICA_URL`, `LITESTREAM_ACCESS_KEY_ID`, `LITESTREAM_SECRET_ACCESS_KEY` (optional)

## Releases

Releases are managed independently from the monorepo using tag prefixes:

| Component | Tag format | Example |
|---|---|---|
| Server | `v*` | `v1.0.0` |
| Modem Agent | `modem-agent/v*` | `modem-agent/v0.3.0` |

Creating a tag triggers the corresponding GitHub Actions workflow, which builds binaries for all platforms and publishes a GitHub Release.

```bash
# Release a new server version
git tag v1.0.0
git push --tags

# Release a new modem agent version
git tag modem-agent/v0.1.0
git push --tags
```

## Production URLs

Replace `yourdomain.com` with your domain.

| Service | URL |
|---------|-----|
| App | `https://yourdomain.com` |
| PocketBase Admin | `https://yourdomain.com/_/` |
| API base URL | `https://yourdomain.com/api/` |
