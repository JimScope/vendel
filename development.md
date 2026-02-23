# Ender - Development

## Docker Compose

* Inicia el stack local con Docker Compose:

```bash
docker compose up -d
```

* Ahora puedes abrir tu navegador e interactuar con estas URLs:

App (API + Frontend): <http://localhost:8090>

PocketBase Admin Dashboard: <http://localhost:8090/_/>

MailCatcher: <http://localhost:1080>

Para revisar los logs:

```bash
docker compose logs -f app
```

## Mailcatcher

Para desarrollo local con Mailcatcher, inicia el perfil `dev`:

```bash
docker compose --profile dev up -d
```

Mailcatcher captura todos los emails enviados por PocketBase durante el desarrollo local. Todos los emails capturados se pueden ver en <http://localhost:1080>.

## Desarrollo Local

### Backend

```bash
cd backend

# Ejecutar servidor de desarrollo
go run . serve --http=0.0.0.0:8090

# Verificar compilación
go build ./...

# Compilar binario
go build -o ender .
./ender serve --http=0.0.0.0:8090
```

El backend corre en `http://localhost:8090`. El admin dashboard de PocketBase está disponible en `http://localhost:8090/_/`.

Las colecciones se definen en `migrations/1740000000_initial.go`. Con `ENVIRONMENT != production`, las auto-migraciones están habilitadas — los cambios hechos en el admin UI se reflejan automáticamente en archivos de migración.

### Frontend

```bash
cd frontend

# Instalar dependencias
npm install

# Servidor de desarrollo
npm run dev

# Build
npm run build

# Lint
npm run lint
```

El frontend corre en `http://localhost:5173` y se conecta al backend via PocketBase JS SDK.

## Variables de Entorno

El archivo `.env` contiene todas las configuraciones. Ver `.env.example` para referencia.

Variables principales:

| Variable | Descripción | Default |
|----------|-------------|---------|
| `ENVIRONMENT` | Entorno de ejecución | `local` |
| `FIRST_SUPERUSER` | Email del superusuario | `admin@ender.app` |
| `FIRST_SUPERUSER_PASSWORD` | Contraseña del superusuario | `changethis` |
| `FIREBASE_SERVICE_ACCOUNT_JSON` | JSON de Firebase para FCM | - |
| `GOOGLE_CLIENT_ID` | OAuth Google | - |
| `GITHUB_CLIENT_ID` | OAuth GitHub | - |
| `QVAPAY_APP_ID` | App ID de QvaPay | - |
| `QVAPAY_APP_SECRET` | App Secret de QvaPay | - |
| `WEBHOOK_ENCRYPTION_KEY` | Clave AES para secretos de webhooks | - |
| `SMTP_HOST` | Host SMTP | `localhost` |
| `SMTP_PORT` | Puerto SMTP | `1025` |
| `LITESTREAM_REPLICA_URL` | URL S3 para backup (opcional) | - |
| `APP_URL` | URL de la app | `http://localhost:8090` |
| `FRONTEND_URL` | URL del frontend | `http://localhost:5173` |

## URLs de Desarrollo

| Servicio | URL |
|----------|-----|
| App (API + Frontend) | http://localhost:8090 |
| PocketBase Admin | http://localhost:8090/_/ |
| Frontend (dev server) | http://localhost:5173 |
| Mailcatcher | http://localhost:1080 |
