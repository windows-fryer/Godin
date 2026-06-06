CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE SCHEMA IF NOT EXISTS eris;

CREATE TABLE IF NOT EXISTS eris.guilds (
    guild_id bigint PRIMARY KEY,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deleted', 'failed')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS eris.services (
    service_id uuid PRIMARY KEY,
    guild_id bigint NOT NULL REFERENCES eris.guilds(guild_id) ON DELETE RESTRICT,
    bot_token text NOT NULL,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('provisioning', 'active', 'failed', 'deleted')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS eris.channels (
    channel_id bigint PRIMARY KEY,
    guild_id bigint NOT NULL REFERENCES eris.guilds(guild_id) ON DELETE CASCADE,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'failed', 'deleted')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS eris.webhooks (
    webhook_id bigint PRIMARY KEY,
    channel_id bigint NOT NULL REFERENCES eris.channels(channel_id) ON DELETE CASCADE,
    webhook_token text NOT NULL,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'failed', 'deleted')),
    failure_count integer NOT NULL DEFAULT 0 CHECK (failure_count >= 0),
    last_used_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS eris.files (
    file_id uuid PRIMARY KEY,
    service_id uuid NOT NULL REFERENCES eris.services(service_id) ON DELETE RESTRICT,
    file_name text NOT NULL,
    content_type text,
    size_bytes bigint,
    uploaded_bytes bigint NOT NULL DEFAULT 0 CHECK (uploaded_bytes >= 0),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'uploading', 'complete', 'failed', 'deleted')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz
);

CREATE TABLE IF NOT EXISTS eris.sessions (
    session_id uuid PRIMARY KEY,
    service_id uuid NOT NULL REFERENCES eris.services(service_id) ON DELETE CASCADE,
    file_id uuid NOT NULL UNIQUE REFERENCES eris.files(file_id) ON DELETE CASCADE,
    file_chunk_size integer NOT NULL CHECK (file_chunk_size > 0),
    expiration_time timestamptz NOT NULL,
    status text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'complete', 'expired', 'cancelled')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS eris.file_parts (
    file_part_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id uuid NOT NULL REFERENCES eris.files(file_id) ON DELETE CASCADE,
    channel_id bigint NOT NULL REFERENCES eris.channels(channel_id) ON DELETE RESTRICT,
    message_id text NOT NULL,
    message_url text NOT NULL,
    message_expiration timestamptz NOT NULL,
    part_index integer NOT NULL CHECK (part_index >= 0),
    size_bytes bigint NOT NULL CHECK (size_bytes > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (file_id, part_index)
);

CREATE INDEX IF NOT EXISTS sessions_open_expiration_idx ON eris.sessions(expiration_time) WHERE status = 'open';
CREATE INDEX IF NOT EXISTS file_parts_file_part_index_idx ON eris.file_parts(file_id, part_index);
CREATE INDEX IF NOT EXISTS webhooks_active_channel_idx ON eris.webhooks(channel_id) WHERE status = 'active';
