# Android agent: reliable pickup and reporting (pending)

Status: to do in the Android app. Backend mitigations shipped in v0.6.3 and
v0.6.4 (see "Backend mitigations" below).

## What happened (2026-11-24)

A message from Nexo ATS (`3zmk96bfgvfdih0`) to a Cuban number went through
these states on the only device, a Samsung A13:

| UTC | State | Meaning |
| --- | --- | --- |
| 21:19:10 | `assigned` | Accepted by the API, assigned to the device, FCM push sent |
| 21:38:29 | `sending` | The app claimed it via `/api/sms/pending`, 19 minutes later |
| 21:40:44 | — | The device talked to the backend again, but never reported a result |
| 21:57:41 | — | The app was opened by hand |
| 22:00:04 | `sent` | The retry cron re-queued it and the foreground app sent it at once |

On 2026-11-23 another message ended `failed` with "agent claimed the message
but never reported a result", and the successful tests took 7–13 minutes to
be sent.

## 2026-11-26: three more failures, found with adb logcat

A message from Nexo ATS (`ntrzglt1cm13dez`, sent 15:58:08 UTC) to the gateway
itself. Device logs (Samsung A13, Android 14, app 0.2.0):

| Local time (UTC-4) | What happened |
| --- | --- |
| 11:58:15 | FCM push arrived at once; `SmsSenderService` started and called `GET /api/sms/pending`. The request hung in the phone's network (it never reached Cloudflare) and failed after 30 s with `SocketTimeoutException`; the service logged "Service error" and **did not retry** |
| 12:00:54 | Samsung killed a background launch after 3 s (`AL_Kill : over 3 sec`) |
| 12:01:18 | The user opened the app. **The main screen called `GET /api/sms/pending`, which claims messages (assigned → sending), but nothing in that process sent them** |
| 12:03:22 | Android killed the app when the SMS permission was revoked; the claimed message stayed in `sending` |
| 12:15:01 | After the rescue, the background service claimed it again and called `sendTextMessage`; Android answered at once `getSubmitPdu() returned null` → `RESULT_ERROR_GENERIC_FAILURE` |

The last failure is encoding: the body had 124 characters with accents typed
as a combining mark (U+0301). Anything outside the GSM 03.38 alphabet forces
UCS-2, where one SMS holds 70 characters; the body needed 2 parts, but the app
sends with the single-part API.

## Problems in the app

1. **Late pickup.** The FCM push arrives late or not at all on this device
   (Cuban network, Samsung battery management), so messages wait until the
   app wakes up on its own.
2. **Claimed but never reported.** After `/api/sms/pending` moves a batch to
   `sending`, the app is killed (or the send fails silently) before it calls
   the status endpoint. A healthy send reports in well under a minute.

## Changes to make in the app

- **Do not claim from the main screen.** The UI must never call
  `/api/sms/pending` (it is a claim, not a read): use `/api/sms/messages` to
  show messages, or start `SmsSenderService` and let it claim and send.
- **Send multipart.** Use `SmsManager.divideMessage()` and
  `sendMultipartTextMessage()`, and report `sent` only when every part's sent
  intent succeeds (`failed` with the first error code otherwise).
- **Retry the pending fetch.** When `/api/sms/pending` fails (timeout, no
  network), retry with backoff inside the service, or reschedule a
  `PendingSyncWorker` with a network constraint, instead of giving up.

- **Poll while active.** While the app is in the foreground, or while a
  foreground service is running, call `/api/sms/pending` periodically (for
  example every 30–60 seconds) instead of relying only on FCM.
- **Foreground service for sending.** Claim and send inside a foreground
  service (with its notification), so Android does not kill the process
  between claiming the batch and reporting the result.
- **Always report.** Report every claimed message: `sent` when the
  `SmsManager` sent intent succeeds, `failed` with the result code otherwise
  (no service, radio off, generic failure, no credit). Never leave a claimed
  message without a report.
- **Ask to ignore battery optimizations.** On first run, guide the user to
  set the app to "Unrestricted" battery use
  (`ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS`), and show a warning in the
  app when it is still optimized.
- **Optional: claim only what can be sent now.** Claim a small batch right
  before sending it, so a killed process strands as few messages as
  possible (see `SMSClaimBatchSize`).

## Backend mitigations

- v0.6.3: the retry cron runs every 5 minutes (was 15).
- v0.6.4: a claimed message with no report is re-queued after 3 minutes (was
  10) by a cron that runs every minute, and an agent claims at most 5
  messages at a time so a slow modem still finishes within that window.
- v0.6.4: bodies assigned to an Android device are folded into the GSM-7
  alphabet (`FoldToGSM7`: á → a, ó → o; ñ, é, ü, ¿ and ¡ are kept; typographic
  quotes and dashes become ASCII), so a body up to 160 characters goes out as
  one plain SMS. Bodies over 160 GSM-7 characters still need the multipart fix.

## Until then

On the gateway phone:

- Set Vendel to "Unrestricted" in Settings › Apps › Vendel › Battery, and add
  it to Settings › Battery › Background usage limits › Never sleeping apps.
- Do not open the main screen while messages are pending (it claims them
  without sending) and do not swipe the app away from Recents; let the
  background service do the work.
- Keep SMS templates short and plain: up to 160 characters without accents
  beyond ñ/é/ü, or up to 70 characters otherwise (v0.6.4 folds accents for
  Android devices).
