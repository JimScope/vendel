# Vendel Android App — Implementation Plan

## Context

Vendel needs a native Android app (Kotlin) that turns a phone into an SMS gateway. The backend already supports the full flow: FCM tickle notifications, pending message fetch, status reporting, and incoming SMS forwarding. A Go modem-agent exists as reference. The app lives in `android/` in the monorepo.

**User decisions**: Min API 26, QR scan + manual setup, forward incoming SMS.

---

## Phase 1: Project Skeleton + Secure Config

Create the Gradle project and encrypted preferences.

**Files to create:**
- `android/build.gradle.kts` — root build file (AGP, Kotlin, Hilt, KSP, Google Services plugins)
- `android/settings.gradle.kts` — module inclusion
- `android/gradle.properties` — JVM args, AndroidX opt-in
- `android/gradle/libs.versions.toml` — version catalog (all deps in one place)
- `android/gradlew`, `android/gradlew.bat`, `android/gradle/wrapper/*` — Gradle wrapper
- `android/.gitignore` — standard Android + `google-services.json`
- `android/google-services.json.example` — placeholder with instructions
- `android/app/build.gradle.kts` — `namespace = "cc.vendel.gateway"`, minSdk 26, targetSdk 35
- `android/app/src/main/AndroidManifest.xml` — all permissions, services, receivers
- `android/app/src/main/java/cc/vendel/gateway/VendelApp.kt` — `@HiltAndroidApp`, WorkManager init
- `android/app/src/main/java/cc/vendel/gateway/data/preferences/SecurePreferences.kt` — EncryptedSharedPreferences wrapper (serverUrl, apiKey, deviceId, pendingFcmToken)
- `android/app/src/main/java/cc/vendel/gateway/data/repository/ConfigRepository.kt` — StateFlow<ConnectionConfig>
- `android/app/src/main/java/cc/vendel/gateway/di/AppModule.kt` — Hilt: prefs, config repo

**Key deps:** Jetpack Compose + Material 3, Hilt, Room, Retrofit + OkHttp + Moshi, CameraX + ML Kit, WorkManager, Firebase Messaging, Security Crypto

**Verify:** Project compiles with `./gradlew assembleDebug`

---

## Phase 2: Network Layer (Retrofit)

Mirror the modem-agent's `VendelClient` (ref: `modem-agent/client.go`).

**Files to create:**
- `data/remote/VendelApi.kt` — Retrofit interface:
  - `GET /api/sms/pending` → `PendingResponse`
  - `POST /api/sms/report` → `StatusReportRequest`
  - `POST /api/sms/incoming` → `IncomingSmsRequest`
  - `POST /api/sms/fcm-token` → `FcmTokenRequest`
- `data/remote/ApiKeyInterceptor.kt` — OkHttp interceptor adding `X-API-Key` header from SecurePreferences
- `data/remote/dto/PendingResponse.kt` — `{deviceId, messages: [{messageId, recipient, body}]}`
- `data/remote/dto/StatusReportRequest.kt` — `{messageId, status, errorMessage}`
- `data/remote/dto/IncomingSmsRequest.kt` — `{fromNumber, body, timestamp}`
- `data/remote/dto/FcmTokenRequest.kt` — `{fcmToken}`
- `di/NetworkModule.kt` — OkHttp + Moshi + Retrofit with dynamic base URL from SecurePreferences

**Verify:** Unit tests with MockWebServer for all 4 endpoints

---

## Phase 3: Room Database + Repository

Local persistence for offline resilience and message log UI.

**Files to create:**
- `data/local/entity/PendingReportEntity.kt` — queued status reports (messageId, status, errorMessage, createdAt)
- `data/local/entity/MessageLogEntity.kt` — local log (messageId, recipient, body, direction, status, errorMessage, timestamp)
- `data/local/dao/PendingReportDao.kt` — insert, getAll, delete, countFlow
- `data/local/dao/MessageLogDao.kt` — insert, updateStatus, observeAll, countByStatus, pruneOlderThan
- `data/local/VendelDatabase.kt` — Room database with both entities
- `data/repository/SmsRepository.kt` — orchestration:
  - `fetchAndProcessPending()` — GET pending → insert to log → return messages
  - `reportStatus(messageId, status, error)` — try POST, fallback to Room queue
  - `reportIncoming(from, body, timestamp)` — POST + insert to log
  - `flushQueuedReports()` — retry all PendingReportEntity rows
  - `updateFcmToken(token)` — POST to /api/sms/fcm-token
- `di/DatabaseModule.kt` — Room database + DAOs

**Verify:** Instrumented tests for Room DAOs, unit tests for SmsRepository with mocked deps

---

## Phase 4: Foreground Service + SMS Sending

The core runtime.

**Files to create:**
- `service/SmsSenderService.kt` — foreground service:
  1. Flush queued reports
  2. Fetch pending messages
  3. Send each SMS via `SmsManager.sendTextMessage` (sequential, 1s delay)
  4. For messages >160 chars: `sendMultipartTextMessage`
  5. Update notification with progress ("Sending 3/10...")
  6. Stop self when done
- `service/SmsSentReceiver.kt` — dynamic BroadcastReceiver per-message sentIntent → reportStatus("sent"/"failed")
- `service/SmsDeliveredReceiver.kt` — dynamic BroadcastReceiver per-message deliveryIntent → reportStatus("delivered")
- `service/BootReceiver.kt` — on BOOT_COMPLETED, enqueue PendingSyncWorker

**Key behaviors:**
- Sequential sending (like modem-agent) to avoid carrier throttling
- PendingIntent extras carry messageId for callback matching
- Multipart SMS: track all parts, report only when all complete
- Error codes mapped to human-readable messages

**Verify:** Robolectric tests for service lifecycle and send sequence

---

## Phase 5: FCM + WorkManager

Wake the app on push + periodic safety net.

**Files to create:**
- `service/VendelFirebaseService.kt` — `FirebaseMessagingService`:
  - `onMessageReceived`: if `data["type"] == "tickle"` → start SmsSenderService
  - `onNewToken`: POST to backend, fallback save to SecurePreferences.pendingFcmToken
- `worker/PendingSyncWorker.kt` — periodic (15 min), `NetworkType.CONNECTED` constraint:
  - Flush queued reports → fetch pending → start service if any
  - Also prune old message logs (>7 days)
- `worker/ReportFlushWorker.kt` — one-shot, enqueued on connectivity restored

**FCM tickle payload** (data-only, no notification body):
```json
{"data": {"type": "tickle", "count": "5"}, "android": {"priority": "high"}}
```

**Verify:** Unit tests for FCM message handling, WorkManager execution

---

## Phase 6: Incoming SMS Receiver

**Files to create:**
- `service/SmsReceiver.kt` — BroadcastReceiver for `SMS_RECEIVED`:
  - Reassemble multipart via `Telephony.Sms.Intents.getMessagesFromIntent`
  - `goAsync()` + coroutine for network call
  - POST to `/api/sms/incoming`
  - Backend deduplicates within 5-min window (ref: `backend/services/sms.go:142-177`)

**Verify:** Robolectric test with mock SMS PDUs

---

## Phase 7: UI — Setup Screen with QR Scanning

**Files to create:**
- `ui/theme/Color.kt` — Vendel palette from `frontend/src/index.css` (accent: #2dd4a8)
- `ui/theme/Type.kt` — Inter + Libre Baskerville + JetBrains Mono
- `ui/theme/Theme.kt` — Material 3 light/dark with Vendel colors
- `ui/navigation/VendelNavigation.kt` — NavHost (Setup → Status, Log, Settings)
- `ui/setup/SetupScreen.kt` — two modes:
  - **QR scan**: CameraX preview + ML Kit barcode → parse `{server_instance, api_key, version}` (ref: `frontend/src/components/Devices/AddDevice.tsx:92-99`)
  - **Manual input**: server URL + API key text fields
- `ui/setup/SetupViewModel.kt` — validate connection (fetchPending), save config, register FCM token
- `MainActivity.kt` — single Activity, Compose host, permission requests

**Verify:** Compose UI test for QR parsing and manual input validation

---

## Phase 8: UI — Status, Log, Settings

**Files to create:**
- `ui/status/StatusScreen.kt` + `StatusViewModel.kt`:
  - Connection indicator (green/red)
  - Stats cards: sent, failed, pending, queued reports (from Room flows)
  - "Sync Now" button
- `ui/log/MessageLogScreen.kt` + `MessageLogViewModel.kt`:
  - LazyColumn from `messageLogDao.observeAll()`
  - Direction icon, recipient, body preview, color-coded status badge, timestamp
- `ui/settings/SettingsScreen.kt` + `SettingsViewModel.kt`:
  - Server URL, device ID (read-only)
  - Reconnect / Disconnect buttons
  - Battery optimization guide (`ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS`)
  - Incoming SMS toggle, app version

**Verify:** Compose UI tests for stats display

---

## Phase 9: Polish + Hardening

- Permission request chain with rationale dialogs in MainActivity
- Multipart SMS part tracking (`ConcurrentHashMap<String, AtomicInteger>`)
- Rate limiting: 1s delay between sends
- ProGuard rules for Moshi, Room, Hilt, Firebase
- Network connectivity listener → enqueue ReportFlushWorker
- Update monorepo `.gitignore` for Android build artifacts
- `foregroundServiceType="connectedDevice"` for Android 14+

**Verify:** Full end-to-end test on real device: QR scan → FCM tickle → SMS sent → status reported → incoming SMS forwarded

---

## Critical Backend References

| File | What it defines |
|------|----------------|
| `backend/handlers/sms.go` | All 4 API endpoints (pending, report, incoming, fcm-token) |
| `backend/services/notification.go` | FCM tickle payload format |
| `modem-agent/client.go` | Reference VendelClient implementation to mirror |
| `frontend/src/components/Devices/AddDevice.tsx:92-99` | QR code JSON payload format |
| `frontend/src/index.css` | Design system color tokens |
| `backend/migrations/1740000000_initial.go` | Collection schemas (sms_devices, sms_messages) |
