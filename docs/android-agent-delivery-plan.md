# Android agent: reliable pickup and reporting (pending)

Status: to do in the Android app. The backend side (retry cron every 5 minutes)
is already changed.

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

## Two problems in the app

1. **Late pickup.** The FCM push arrives late or not at all on this device
   (Cuban network, Samsung battery management), so messages wait until the
   app wakes up on its own.
2. **Claimed but never reported.** After `/api/sms/pending` moves a batch to
   `sending`, the app is killed (or the send fails silently) before it calls
   the status endpoint. A healthy send reports in well under a minute.

## Changes to make in the app

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

## Until then

On the gateway phone, set Vendel to "Unrestricted" in Settings › Apps ›
Vendel › Battery and keep the app open during demos.
