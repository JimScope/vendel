<p align="center">
  <img src="img/vendel-icon.png" alt="Vendel" width="80" height="80" />
</p>

<h1 align="center">Vendel</h1>

<p align="center">Plataforma SMS Gateway</p>

<p align="center">
  <a href="https://vendel.cc">Website</a> &middot;
  <a href="https://app.vendel.cc">Dashboard Demo</a> &middot;
  <a href="./README.md">Read in English</a>
</p>

Vendel es una plataforma full-stack para gestión y envío de SMS a través de dispositivos conectados. Permite enviar mensajes SMS usando dispositivos registrados (teléfonos Android o módems) como gateways, con gestión de cuotas, webhooks y soporte para múltiples usuarios.

<p align="center">
  <img src="img/homepage.png" alt="Vendel Homepage" width="800" />
</p>

## Stack Tecnológico

### Backend
- **Framework**: [PocketBase](https://pocketbase.io/) (Go)
- **Base de datos**: SQLite (embebido)
- **Autenticación**: JWT (built-in), OAuth2 (Google, GitHub)
- **Push Notifications**: Firebase Cloud Messaging (FCM)
- **Email**: Soporte SMTP integrado, Mailcatcher para dev local
- **Admin**: Dashboard de PocketBase en `/_/`

### Frontend
- **Framework**: React 19 con TypeScript
- **Build**: Vite
- **Estado**: TanStack Query + TanStack Router
- **Estilos**: Tailwind CSS + shadcn/ui
- **Formularios**: React Hook Form + Zod
- **Cliente API**: PocketBase JS SDK
- **Tests E2E**: Playwright

### Infraestructura
- **Contenedores**: Docker & Docker Compose
- **CI/CD**: GitHub Actions
- **Despliegue**: Binario único (~50MB imagen Docker)

## Funcionalidades Principales

### SMS
- Envío de SMS individuales y masivos
- Distribución round-robin entre dispositivos
- Cola de mensajes cuando no hay dispositivos online
- Tracking de estado (pending, queued, processing, sent, delivered, failed)
- Historial y reportes de SMS
- Soporte para SMS entrantes

### Dispositivos
- Registro de dispositivos con API keys únicas
- Gestión de tokens FCM para push notifications
- Monitoreo de estado de dispositivos

### Cuotas y Planes
- Múltiples planes de suscripción
- Tracking de cuota mensual de SMS
- Límites de dispositivos por plan
- Reset automático mensual de cuota (cron)

### Webhooks
- Suscripciones de webhooks configurables por tipo de evento
- Eventos soportados: `sms_received`, `sms_sent`, `sms_delivered`, `sms_failed`
- Payloads firmados con HMAC-SHA256 y JSON con claves ordenadas

### Pagos
- Abstracción de proveedores de pago (QvaPay)
- Gestión de ciclo de vida de suscripciones
- Flujos de pago por factura y autorización

### Integraciones
- API keys múltiples por usuario
- Códigos QR para onboarding de dispositivos
- API pública para sistemas externos

## Estructura del Proyecto

```
vendel/
├── backend/                    # Go + PocketBase API
│   ├── main.go                 # Setup de PocketBase, hooks, cron, rutas
│   ├── go.mod / go.sum
│   ├── handlers/               # Rutas API custom (sms, planes, webhooks)
│   ├── services/               # Lógica de negocio (SMS, FCM, cuota, suscripciones)
│   │   └── payment/            # Proveedor de pago (QvaPay)
│   ├── middleware/              # Auth por API key, modo mantenimiento
│   └── migrations/             # Definiciones de colecciones + datos semilla
├── frontend/                   # App React
│   ├── src/
│   │   ├── routes/             # Páginas (TanStack Router)
│   │   ├── components/         # Componentes React
│   │   ├── hooks/              # Hooks custom (PocketBase SDK)
│   │   └── lib/pocketbase.ts   # Cliente PocketBase
│   └── tests/                  # Tests Playwright
├── Dockerfile                  # Multi-stage (node + go + alpine)
├── docker-compose.yml
├── litestream.yml              # Config de replicación Litestream (opt-in)
├── entrypoint.sh               # Startup condicional (con/sin Litestream)
└── .env                        # Variables de entorno
```

## Inicio Rápido

### Requisitos Previos
- Docker y Docker Compose
- Go 1.23+ (para dev local del backend)
- Node.js 24+ (para dev local del frontend)

### Desarrollo con Docker Compose (Recomendado)

```bash
# Iniciar la app
docker compose up -d

# Ver logs
docker compose logs -f app
```

**Servicios disponibles:**
| Servicio | URL |
|----------|-----|
| App (API + Frontend) | http://localhost:8090 |
| PocketBase Admin | http://localhost:8090/_/ |
| Mailcatcher | http://localhost:1080 |

### Desarrollo Manual

#### Backend
```bash
cd backend

# Ejecutar servidor de desarrollo
go run . serve --http=0.0.0.0:8090

# Compilar binario
go build -o vendel .
./vendel serve --http=0.0.0.0:8090
```

#### Frontend
```bash
cd frontend

# Instalar dependencias
npm install

# Servidor de desarrollo
npm run dev

# Build
npm run build

# Tests E2E
npx playwright test
```

## Configuración

### Variables de Entorno

Crea un archivo `.env` en la raíz del proyecto:

```env
# Core
ENVIRONMENT=local
FIRST_SUPERUSER=admin@vendel.cc
FIRST_SUPERUSER_PASSWORD=changethis

# Firebase (push notifications)
FIREBASE_SERVICE_ACCOUNT_JSON=<json-de-firebase>

# OAuth (opcional)
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=

# Pago (QvaPay)
QVAPAY_APP_ID=
QVAPAY_APP_SECRET=

# Seguridad
WEBHOOK_ENCRYPTION_KEY=         # Clave AES para secretos de webhooks

# SMTP (por defecto localhost:1025 para mailcatcher en dev)
SMTP_HOST=
SMTP_PORT=
SMTP_USERNAME=
SMTP_PASSWORD=

# Backup (Litestream - opcional)
LITESTREAM_REPLICA_URL=         # ej. s3://my-bucket/vendel/data
LITESTREAM_ACCESS_KEY_ID=
LITESTREAM_SECRET_ACCESS_KEY=

# URLs
APP_URL=http://localhost:8090
FRONTEND_URL=http://localhost:5173
```

## Testing

### Frontend
```bash
# Tests E2E
npx playwright test

# Modo UI
npx playwright test --ui
```

## Despliegue

Ver [deployment.md](./deployment.md) para instrucciones detalladas de despliegue en producción.

## Documentación Adicional

- [Desarrollo](./development.md) - Guía de desarrollo local
- [Despliegue](./deployment.md) - Instrucciones de producción

## Licencia

MIT License
