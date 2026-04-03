# Vendel Roadmap

## 1. Contact Lists / Groups
Import and manage contact lists with groups. Enable bulk SMS to groups instead of typing numbers manually.

- [ ] `contacts` collection (name, phone, tags/groups, user)
- [ ] `contact_groups` collection (name, user)
- [ ] Import from CSV
- [ ] Group selector in SMS send flow
- [ ] API endpoint for CRUD

## 2. SMS Retry with Device Failover
When a device fails to send, automatically re-assign the message to another available device instead of marking it as `failed`.

- [ ] Retry logic in `ProcessSMSAck` on failure
- [ ] Max retry count (configurable, default 2)
- [ ] Device cooldown after consecutive failures
- [ ] Failover only to devices with recent heartbeat

## 3. MMS / Image Support
Send images and media alongside SMS, depending on device capabilities.

- [ ] Research Android SMS/MMS API limitations
- [ ] `media` file field on `sms_messages`
- [ ] FCM payload with media URL
- [ ] Android app MMS sending support
- [ ] Modem agent: evaluate MMS AT commands (device-dependent)

## 4. SMS Templates with Variables
Reusable message templates with variable substitution (`{{name}}`, `{{code}}`).

- [ ] Templates UI (create, edit, preview)
- [ ] Variable interpolation at send time
- [ ] Template selector in SMS compose
- [ ] API support: `template_id` + `variables` in send request

## 5. Scheduled SMS with Recurrence
Extend scheduled SMS to support recurring sends via cron expressions with full UI support.

- [ ] Frontend cron builder (visual, not raw expression)
- [ ] Recurrence preview (next 5 runs)
- [ ] Pause/resume scheduled SMS
- [ ] Integration with contact groups (scheduled broadcast)
