# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-04-29

### Added
- `GET /api/devices` — paginated list of the authenticated user's registered devices. Supports `page`, `per_page` (max 200), and `device_type` filters. Auth: JWT or `vk_` API key.
- `GET /api/sms/messages` — paginated SMS message history. Supports `page`, `per_page`, `status`, `device_id`, `batch_id`, `recipient`, `from`, and `to` (ISO8601) filters. Auth: JWT or `vk_` API key.
- `GET /api/plans/quota` extended with `scheduled_sms_count`, `integrations_count`, `max_scheduled_sms`, and `max_integrations` so clients can show the four resource counters consistently.

### Removed
- Deprecated quota response keys `scheduled_sms_active` and `integrations_created`. Clients should switch to the new `scheduled_sms_count` and `integrations_count` names introduced in this release.

## Earlier versions

For releases prior to 0.4.0 (`v0.1.0-beta.x`, `v0.2.0`, `v0.3.0`, `v0.3.1`, `v0.3.2`), see the published GitHub releases at <https://github.com/JimScope/vendel/releases>.
