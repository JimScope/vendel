# Release Notes

## v2.0.0 - PocketBase Rewrite

### Breaking Changes
- Backend rewritten from FastAPI (Python) to PocketBase (Go)
- Database changed from PostgreSQL to SQLite (embedded)
- API endpoints changed from `/api/v1/` to PocketBase conventions
- Frontend API client changed from auto-generated OpenAPI to PocketBase JS SDK

### Added
- PocketBase admin dashboard at `/_/`
- Built-in OAuth2 support (Google, GitHub)
- Built-in email verification and password reset
- Cron jobs for quota reset, renewal checks, SMS retry
- HMAC-SHA256 webhook signatures
- Payment provider abstraction (QvaPay, Tropipay)
- Subscription lifecycle management

### Removed
- PostgreSQL dependency
- QStash (replaced by goroutines + cron)
- Maileroo (PocketBase handles email)
- Adminer (replaced by PocketBase admin)
- Auto-generated OpenAPI client
- Pre-commit hooks (Python-specific)

### Infrastructure
- Single binary deployment (~50MB Docker image)
- Docker Compose reduced from 5 services to 2
- Environment variables reduced from 40+ to ~18

---

## v1.x - FastAPI

### Features
- Setup de logs
- Integración con QStash para cola de mensajes
- Códigos QR para integraciones externas y dispositivos
- Envío de SMS masivo (múltiples destinatarios)
- Gestión de cuotas y planes
- API keys externas para integraciones
