# Godin rewrite architecture notes

The rewrite should treat PostgreSQL as the source of truth for service, upload session, file, and file part state. Discord is a provider that stores opaque file parts, not the core domain model.

## Core boundaries

- HTTP handlers should decode requests, read path values, call application services, and encode responses.
- Application services should own lifecycle rules such as service provisioning, upload session expiration, append-only file part writes, and upload completion.
- Repositories should own SQL and transaction details.
- Provider clients should own Discord-specific operations such as guild provisioning, webhook selection, message upload, URL refresh, and cleanup.

## Storage invariants

- Services have explicit lifecycle states: `provisioning`, `active`, `failed`, and `deleted`.
- Upload sessions have explicit lifecycle states: `open`, `complete`, `expired`, and `cancelled`.
- Files have explicit lifecycle states: `pending`, `uploading`, `complete`, `failed`, and `deleted`.
- File parts are append-only and must be unique by `(file_id, part_index)`.
- Upload sessions must store `expiration_time`; client-facing expiry values must always match durable database state.

## Provider safety rules

- Discord resources created by this service must use a Godin-owned prefix and must be the only resources cleaned up by Godin.
- Discord attachment URLs are temporary hints. Durable records should keep stable provider references such as channel ID, message ID, part index, and URL expiration.
- Raw provider credentials should eventually move to an encrypted credential store or a secret manager reference instead of ordinary service rows.
