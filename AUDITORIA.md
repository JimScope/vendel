# Auditoría de mejoras — Vendel

**Fecha:** 2026-09-15
**Alcance:** `backend/` (Go/PocketBase), `frontend/` (React 19 / Vite 8 / TS), `modem-agent/` (Go), `.github/` (CI/CD).
**Metodología:** revisión estática multi-agente + verificación manual contra el código real. La auditoría de bugs previa (`SEC-`/`BE-`/`FE-`/`MA-`) está **100% resuelta**; este documento la reemplaza con oportunidades de **mejora hacia adelante** (fiabilidad, observabilidad, rendimiento, testing, CI y pulido), no deuda conocida.

## Cómo usar este documento

Cada hallazgo tiene un ID estable (`REL-`, `OBS-`, `PERF-`, `TEST-`, `CI-`, `SEC-`, `DX-`), una casilla de progreso, su `file:line`, la causa raíz, el fix propuesto y el estado de verificación.

**Severidad:** 🔴 Crítica · 🟠 Alta · 🟡 Media · ⚪ Baja
**Esfuerzo:** S = puntual (<1h) · M = medio · L = amplio
**Verificación:** ✅ confirmado en código · ⚠️ plausible (no verificado línea a línea)

---

## Tabla de progreso

### Fiabilidad (dinero / pérdida de mensajes)
| ID | Sev | Esf | Título | ✔ |
|----|-----|-----|--------|---|
| REL-1 | 🔴 | M | Fuga de cuota facturable: SMS nunca despachado no la libera | ☐ |
| REL-2 | 🔴 | M | Mensajes externos `assigned` huérfanos ante error de proveedor | ☐ |
| REL-3 | 🟡 | S | Sin tope de destinatarios por request | ☐ |
| REL-4 | 🟡 | M | Drenaje fijo de 50 filas en retry/rescate | ☐ |
| REL-5 | 🟡 | S | Idempotencia de crédito degrada a nada si falta `txHash` | ☐ |

### Observabilidad — ⏭️ descartado (cubierto por los logs nativos de PocketBase)
> Por decisión del proyecto, la observabilidad **ya está cubierta** por el store de logs de PocketBase (`app.Logger()` → `_logs`, filtrable en el admin). No se añade instrumentación adicional. Los siguientes quedan como referencia, **sin acción**.

| ID | Sev | Esf | Título | Estado |
|----|-----|-----|--------|--------|
| OBS-1 | 🟠 | M | Eventos estructurados en transiciones terminales | ⏭️ N/A (logs existentes) |
| OBS-2 | 🟡 | S | Estado del circuit breaker en logs | ⏭️ N/A (logs existentes) |
| OBS-3 | ⚪ | S | `sendFCMTickle` usa `log.Printf` | ⏭️ N/A |

### Rendimiento
| ID | Sev | Esf | Título | ✔ |
|----|-----|-----|--------|---|
| PERF-1 | 🟡 | M | N+1 en resolución de dispositivos dentro de loops de retry | ☐ |
| PERF-2 | 🟡 | M | Queries sin límite en crons de suscripción y expansión de grupos | ☐ |
| PERF-3 | 🟡 | S | `BroadcastModemStatus` O(dispositivos × clientes SSE) | ☐ |

### Testing
| ID | Sev | Esf | Título | ✔ |
|----|-----|-----|--------|---|
| TEST-1 | 🟠 | M | Sin tests en los caminos de dinero (payment providers + balance) | ☐ |
| TEST-2 | 🟠 | M | Sin test de la carrera de reserva de cuota | ☐ |
| TEST-3 | 🟡 | M | Sin tests de orquestación de retry/rescate | ☐ |

### CI / Tooling de seguridad
| ID | Sev | Esf | Título | ✔ |
|----|-----|-----|--------|---|
| CI-1 | 🟠 | M | El frontend no corre en CI (lint/typecheck/build/E2E) | ☐ |
| CI-2 | 🟡 | S | Dependabot no cubre el ecosistema `gomod` | ☐ |
| CI-3 | 🟡 | S | Sin SAST/vuln scanning de Go (gosec/govulncheck) | ☐ |
| CI-4 | ⚪ | S | Sin reporte de cobertura en CI | ☐ |

### Seguridad (backend ya es fuerte; estos son huecos menores)
| ID | Sev | Esf | Título | ✔ |
|----|-----|-----|--------|---|
| SEC-1 | 🟡 | S | Secretos vivos en `.env` de working-tree sin guarda anti-commit | ☐ |
| SEC-2 | ⚪ | S | Sin pinning explícito de orígenes CORS | ☐ |

### Pulido / DX
| ID | Sev | Esf | Título | ✔ |
|----|-----|-----|--------|---|
| DX-1 | ⚪ | S | `frontend/README.md` instruye npm; el estándar es bun | ☐ |
| DX-2 | ⚪ | S | Drift de versión de Playwright (imagen v1.63 vs paquete 1.61) | ☐ |
| DX-3 | ⚪ | S | Parseo de montos con `fmt.Sscanf` que ignora errores | ☐ |

---

## 🔴 Fiabilidad

### REL-1 · 🔴 Crítica · Fuga de cuota facturable · ✅
- [ ] **Resuelto**

**Ubicación:** `backend/services/sms.go:30,53` · `sms.go:268` (`MarkMessageTerminal`) · `sms_retry.go:138,172`

**Causa:** `SendSMS` reserva cuota para todos los destinatarios al inicio (`sms.go:30`). Si el usuario no tiene dispositivo, `len(devices)==0`, se salta el dispatch (`sms.go:53`) y **retorna éxito** con los mensajes en `pending`. Después `rescuePendingMessages` los marca `failed` vía `MarkMessageTerminal`, que **no** llama a `ReleaseSMSQuota` (solo se llama en los caminos de error síncronos, `sms.go:63`). Resultado: el usuario queda facturado por SMS que nunca salieron.

**Fix:** centralizar la liberación de cuota dentro de `MarkMessageTerminal` cuando un mensaje transita a `failed`/terminal **sin haberse enviado** (idempotente, marcando el registro para no doble-liberar). Cierra la raíz común de REL-1 y REL-2.

---

### REL-2 · 🔴 Crítica · Mensajes externos `assigned` huérfanos · ✅
- [ ] **Resuelto**

**Ubicación:** `backend/services/sms_provider_dispatch.go:98-101`

**Causa:** en `dispatchOne`, si `provider.Send` (AEUM/externos) devuelve error de transporte (timeout, blip de red), hace `return` **sin cambiar el estado**. Los proveedores externos son push (no poll): no hay consumidor para `assigned` externo (`RetryFailedMessages` solo toma `failed`, los rescates toman `pending`/`sending`). El mensaje queda varado para siempre y su cuota consumida.

**Fix:** ante error de `provider.Send`, marcar el registro `failed` (lo recoge el retry cron) — combinado con REL-1, eso libera la cuota. Alternativa: rescate para `assigned + externo + stale`.

---

### REL-3 · 🟡 Media · Sin tope de destinatarios por request · ✅
- [ ] **Resuelto**

**Ubicación:** `backend/handlers/sms.go` (`resolveRecipients`) · `services/sms.go:220` (`createMessageRecords`)

**Causa:** solo se acota el largo del cuerpo (`MaxMessageBodyLength`), nunca el número de destinatarios. Un envío a un grupo enorme es una única transacción SQLite grande + un fan-out `FireAndForget` grande; la cuota mensual es el único freno.

**Fix:** guarda explícita de máximo de destinatarios por llamada (constante configurable).

---

### REL-4 · 🟡 Media · Drenaje fijo de 50 filas en retry/rescate · ⚠️
- [ ] **Resuelto**

**Ubicación:** `backend/services/sms_retry.go:48,118,157` · `schedule.go:21`

**Causa:** límite hardcodeado 50 por pasada; `RetryFailedMessages` corre cada 15 min → 200/hora máximo, ignorando el resto en silencio si hay backlog.

**Fix:** bucle acotado hasta agotar la cola (con techo de seguridad) o tamaño de página configurable.

---

### REL-5 · 🟡 Media · Idempotencia de crédito sin `txHash` · ⚠️
- [ ] **Resuelto**

**Ubicación:** `backend/services/balance.go:191,208`

**Causa:** el marcador de idempotencia solo se escribe si `txHash != ""`. Un webhook de crédito sin id de transacción doble-acreditaría en re-entrega. Hoy QvaPay/Stripe/TronDealer siempre lo aportan (latente), pero conviene rechazar eventos de crédito sin clave de idempotencia.

**Fix:** exigir clave de idempotencia en eventos que acreditan saldo; rechazar si falta.

---

## ⏭️ Observabilidad — descartado (cubierto por los logs de PocketBase)

> La observabilidad ya está cubierta por el store de logs nativo de PocketBase (`app.Logger()` → colección `_logs`, filtrable/consultable desde el admin UI). Por decisión del proyecto **no se añade instrumentación adicional** (ni Prometheus/OTel externo ni nuevos eventos). Los ítems OBS-1/OBS-2/OBS-3 quedan documentados como referencia pero **sin acción**.

---

## 🟡 Rendimiento

### PERF-1 · 🟡 Media · N+1 en resolución de dispositivos en retry · ✅
- [ ] **Resuelto**

**Ubicación:** `backend/services/sms_retry.go:81,127` → `assignAvailableDevice` → `resolveDevices` (`sms.go:95`)

**Causa:** una resolución de dispositivos (1-2 queries) por registro; un lote de 50 rescatados = ~100 queries resolviendo el mismo set por usuario.

**Fix:** cachear la resolución por `userId` dentro de una pasada del cron (espejo del `deviceTypeCache` ya usado en `partitionByProvider`).

---

### PERF-2 · 🟡 Media · Queries sin límite en crons · ✅
- [ ] **Resuelto**

**Ubicación:** `backend/services/subscription.go:132,158,193` · `handlers/sms.go:448` (expansión de grupos) · (menores: `webhook.go:489`, `sms.go:113`, `modem.go:51`)

**Causa:** `FindRecordsByFilter(..., 0, 0, ...)` sin límite. Los crons de suscripción y la expansión de contactos de grupo son del tamaño de los datos del usuario y pueden crecer sin techo.

**Fix:** paginar/limitar las queries de cron y la expansión de grupos.

---

### PERF-3 · 🟡 Media · `BroadcastModemStatus` cuadrático · ⚠️
- [ ] **Resuelto**

**Ubicación:** `backend/services/modem.go:60-71`

**Causa:** por cada dispositivo módem escanea todos los `SubscriptionsBroker().Clients()`; disparado por `FireAndForget` en connect/disconnect → O(dispositivos × clientes) por evento.

**Fix:** construir un set `topic-suscrito → bool` una vez por broadcast.

---

## 🟠 Testing

### TEST-1 · 🟠 Alta · Sin tests en los caminos de dinero · ✅
- [ ] **Resuelto**

**Ubicación:** `backend/services/payment/*` (sin `*_test.go`) · `services/balance.go`

**Causa:** cero cobertura en verificación de firmas (Stripe HMAC, QvaPay re-verificación, TronDealer), parseo de montos (`stripe.go:39`) e idempotencia de crédito (`creditAndActivate`). Es la superficie de mayor riesgo (bypass de firma, doble-crédito, redondeo).

**Fix:** table tests de `ParseWebhook`/`verifySignature` por proveedor + test de la transacción de idempotencia.

---

### TEST-2 · 🟠 Alta · Sin test de la carrera de reserva de cuota · ✅
- [ ] **Resuelto**

**Ubicación:** `backend/services/quota.go:184-198` (`ReserveSMSQuota`) · `ReleaseSMSQuota`

**Causa:** el UPDATE condicional que impide el over-send (el invariante de negocio más crítico) no tiene test.

**Fix:** test de concurrencia que afirme que dos reservas paralelas no superan el límite.

---

### TEST-3 · 🟡 Media · Sin tests de orquestación de retry/rescate · ✅
- [ ] **Resuelto**

**Ubicación:** `backend/services/sms_retry.go` · `sms_provider_dispatch.go`

**Causa:** gating de backoff, skip de fallo permanente, rescate pending/sending y el pool de dispatch (donde viven REL-1/REL-2) no tienen tests directos.

**Fix:** tests de las transiciones de rescate y del dispatch (incluyendo el fix de REL-2).

---

## 🟠 CI / Tooling de seguridad

### CI-1 · 🟠 Alta · El frontend no corre en CI · ✅
- [ ] **Resuelto**

**Ubicación:** `.github/workflows/test.yml` (path-filtrado a `backend/**`)

**Causa:** solo corren los tests de Go. No hay lint (biome), typecheck (`tsc`), `vite build` ni los 16 specs de Playwright en ningún push/PR. Builds/lint rotos del frontend mergean libremente.

**Fix:** workflow de frontend con **bun** (`bun install --frozen-lockfile`, biome, `tsc`, `bun run build`, Playwright).

---

### CI-2 · 🟡 Media · Dependabot sin `gomod` · ✅
- [ ] **Resuelto**

**Ubicación:** `.github/dependabot.yml`

**Causa:** cubre github-actions/bun/docker/docker-compose pero **no** los módulos Go de `backend/` ni `modem-agent/` → las deps de Go (incl. parches de seguridad) nunca se auto-actualizan.

**Fix:** añadir dos entradas `package-ecosystem: gomod` (backend y modem-agent).

---

### CI-3 · 🟡 Media · Sin SAST/vuln scanning de Go · ✅
- [ ] **Resuelto**

**Ubicación:** `.github/workflows/`

**Causa:** sin `gosec` ni `govulncheck` en un servicio que maneja pagos, cripto y HTTP sensible a SSRF.

**Fix:** job de CI con `govulncheck ./...` y (opcional) `gosec`.

---

### CI-4 · ⚪ Baja · Sin reporte de cobertura · ⚠️
- [ ] **Resuelto**

**Ubicación:** `.github/workflows/test.yml`

**Causa:** `go test -race` sin `-coverprofile`/codecov → sin visibilidad de tendencias de cobertura.

**Fix:** añadir `-coverprofile` y subir a codecov (o artefacto).

---

## 🟡 Seguridad

### SEC-1 · 🟡 Media · Secretos vivos en `.env` sin guarda · ⚠️
- [ ] **Resuelto**

**Ubicación:** `.env` (gitignored, no commiteado — verificado)

**Causa:** contiene clave privada de Firebase, `GITHUB_CLIENT_SECRET`, `SMTP_PASSWORD`, `WEBHOOK_ENCRYPTION_KEY`, `TRONDEALER_*` reales. Sin gitleaks/secret-scan en CI ni pre-commit, un `git add -f` o un desliz de tooling los filtra.

**Fix:** confirmar que son de dev (rotar si no) + añadir secret scanning (gitleaks) en CI.

---

### SEC-2 · ⚪ Baja · Sin pinning de orígenes CORS · ⚠️
- [ ] **Resuelto**

**Ubicación:** `backend/` (sin config CORS; usa el default de PocketBase)

**Causa:** riesgo bajo (auth por bearer/API-key, no cookies; SPA same-origin embebida), pero conviene pinnear orígenes permitidos explícitamente.

**Fix:** configurar `AllowOrigins` explícito.

---

## ⚪ Pulido / DX

### DX-1 · ⚪ Baja · `README.md` instruye npm en vez de bun · ✅
- [ ] **Resuelto**

**Ubicación:** `frontend/README.md`

**Causa:** instruye `npm install`/`npx playwright`, pero CI, `CLAUDE.md` y Dockerfiles usan bun ("bun.lock es el único lockfile"). Seguir el README genera un `package-lock.json` competidor.

**Fix:** reescribir a `bun install`/`bunx`.

---

### DX-2 · ⚪ Baja · Drift de versión de Playwright · ✅
- [ ] **Resuelto**

**Ubicación:** `frontend/Dockerfile.playwright` (imagen `v1.63.0-noble`) vs `package.json` (`@playwright/test 1.61.1`)

**Causa:** mismatch runner/browser → E2E flaky o divergente.

**Fix:** alinear ambas a la misma versión.

---

### DX-3 · ⚪ Baja · Parseo de montos con `fmt.Sscanf` · ✅
- [ ] **Resuelto**

**Ubicación:** `backend/services/payment/qvapay.go:84,127` · `trondealer.go:132`

**Causa:** `fmt.Sscanf(v, "%f", &amount)` ignora errores → silenciosamente 0. La guarda `expectedAmount > 0` mitiga, pero un monto malformado podría saltar la verificación.

**Fix:** `strconv.ParseFloat` con manejo de error en campos de dinero.

---

## Otros (menor prioridad / seguimiento)

- **`FireAndForget` sin drain en shutdown** (`sms.go:76,83,344`, `webhook.go:510`, `notification.go`, `middleware/apikey.go:58`): las goroutines recuperan panics pero al apagar el proceso se pierden entregas en vuelo; los crons de retry actúan de red de seguridad. Documentar y (opcional) `WaitGroup`/gauge de in-flight.
- **`deriveProviderChannel`** clasifica sender IDs alfanuméricos como RCS (`aws_aeum_events.go:220`) — solo cosmético.

---

## ✅ Ya sólido (no re-tocar)

- **SSRF de webhooks**: allowlist de esquema + bloqueo de IPs privadas + re-validación de IP en `DialContext` (anti DNS-rebinding) + re-validación en redirects + tope de tamaño (`webhook.go:30-131`).
- **Verificación de firma SNS**: HTTPS + regex de host AWS + `.pem` + SSRF en el fetch + cache de cert + pinning de topic-ARN fail-closed (`sns_signature.go`, `aws_aeum_events.go:64`).
- **Webhooks de pago**: Stripe HMAC + rechazo de replay por timestamp; QvaPay re-verificación server-side con match de monto; TronDealer HMAC; idempotencia vía `payment_idempotency` en transacción.
- **API keys** SHA-256 con reveal único (por diseño — **no** es hashing débil), webhooks salientes HMAC, redacción de PII en logs, circuit breaker por host, security headers + CSP, rate limits, login bloqueado hasta verificar email.
- **Carrera de over-send de cuota**: UPDATE condicional con re-check en el WHERE (`quota.go:184`).
- **Amplificación de retries de webhook**: resuelta actualizando el log existente; backoff con jitter anti-thundering-herd (`webhook.go:67`).
- **Rescate de "sending" stale** y su interacción con `SMSClaimBatchSize`/`SMSSendingStaleAfter` (`constants.go:64-81`), cuidadosamente razonado para evitar envíos duplicados.
- **Cifrado del cuerpo en reposo** con descifrado solo en el borde del proveedor; redacción de PII en logs de entrega.
- Devtools tree-shakeados fuera del bundle de prod; React StrictMode; `errorComponent`/`notFoundComponent` a nivel de ruta; skip-link + `sr-only`.

> **Falso positivo descartado:** el índice sobre `provider_message_id` **sí existe** (`migrations/1740000019_aws_aeum_provider.go:53`, `idx_sms_messages_provider_message_id`) — no requiere acción.

---

## 🚀 Roadmap de features (planeado en `docs/`)

- **SMPP / short code** (`docs/smpp-provider-plan.md`) — listo para implementar.
- **RCS** (`docs/rcs-research.md`) — investigado, sin código.
- **AWS End User Messaging** (`docs/aws-end-user-messaging-plan.md`) — aprobado; base ya en `services/smsprovider/`.

La abstracción `services/smsprovider/` (`Provider` con `Name`/`IsConfigured`/`Send`) está bien diseñada para añadir proveedores con un archivo + constructor singleton.
