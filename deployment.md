# Ender - Deployment

Puedes desplegar el proyecto usando Docker Compose en un servidor remoto.

## Preparación

* Tener un servidor remoto disponible.
* Configurar los registros DNS de tu dominio para apuntar a la IP del servidor.
* Instalar y configurar [Docker](https://docs.docker.com/engine/install/) en el servidor remoto (Docker Engine, no Docker Desktop).

## Variables de Entorno

Necesitas configurar algunas variables de entorno.

Configura el `ENVIRONMENT`, por defecto `local` (para desarrollo), pero al desplegar en un servidor pondrías `production`:

```bash
export ENVIRONMENT=production
```

Variables principales:

* `ENVIRONMENT`: Entorno de ejecución (`local`, `staging`, `production`).
* `FIRST_SUPERUSER`: El email del primer superusuario.
* `FIRST_SUPERUSER_PASSWORD`: La contraseña del primer superusuario.
* `FIREBASE_SERVICE_ACCOUNT_JSON`: JSON de cuenta de servicio de Firebase para FCM.
* `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET`: Credenciales OAuth de Google.
* `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET`: Credenciales OAuth de GitHub.
* `QVAPAY_APP_ID` / `QVAPAY_APP_SECRET`: Credenciales de QvaPay.
* `WEBHOOK_ENCRYPTION_KEY`: Clave AES para encriptar secretos de webhooks.
* `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD`: Configuración SMTP.
* `LITESTREAM_REPLICA_URL`: URL de réplica S3 para backup continuo (opcional).
* `LITESTREAM_ACCESS_KEY_ID` / `LITESTREAM_SECRET_ACCESS_KEY`: Credenciales S3 para Litestream.
* `SERVER_BASE_URL`: URL pública del backend (ej. `https://api.example.com`).
* `FRONTEND_HOST`: URL del frontend (ej. `https://app.example.com`).

## Desplegar con Docker Compose

Con las variables de entorno configuradas, puedes desplegar con Docker Compose:

```bash
docker compose up -d
```

La imagen Docker usa un build multi-stage que compila el frontend y el backend Go en un binario único sobre Alpine (~50MB). No requiere base de datos externa — SQLite está embebido en PocketBase.

El volumen `pb_data` persiste la base de datos SQLite entre reinicios.

## Despliegue Continuo (CD)

Puedes usar GitHub Actions para desplegar tu proyecto automáticamente.

### Instalar GitHub Actions Runner

* En tu servidor remoto, crea un usuario para GitHub Actions:

```bash
sudo adduser github
```

* Agrega permisos de Docker al usuario `github`:

```bash
sudo usermod -aG docker github
```

* Cambia temporalmente al usuario `github`:

```bash
sudo su - github
```

* [Instala un GitHub Action self-hosted runner siguiendo la guía oficial](https://docs.github.com/en/actions/hosting-your-own-runners/managing-self-hosted-runners/adding-self-hosted-runners#adding-a-self-hosted-runner-to-a-repository).

* Cuando te pregunte por labels, agrega un label para el entorno, ej. `production`.

* Instálalo como servicio:

```bash
exit
sudo su
cd /home/github/actions-runner
./svc.sh install github
./svc.sh start
./svc.sh status
```

### Configurar Secrets

En tu repositorio, configura secrets para las variables de entorno. Los workflows esperan estos secrets:

* `FIRST_SUPERUSER`
* `FIRST_SUPERUSER_PASSWORD`
* `FIREBASE_SERVICE_ACCOUNT_JSON`
* `QVAPAY_APP_ID` / `QVAPAY_APP_SECRET`
* `WEBHOOK_ENCRYPTION_KEY`
* SMTP credentials (if using email)
* Litestream credentials (if using backup)

## URLs

Reemplaza `example.com` con tu dominio.

### Producción

App: `https://example.com`

PocketBase Admin: `https://example.com/_/`

API base URL: `https://example.com/api/`

### Frontend independiente

El frontend también puede desplegarse por separado (ej. Vercel) apuntando `VITE_API_URL` al backend.
