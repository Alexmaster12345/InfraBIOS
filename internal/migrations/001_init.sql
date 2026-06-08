-- InfraBIOS — initial schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- profiles first (servers reference it)
CREATE TABLE bios_profiles (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT        NOT NULL UNIQUE,
    description TEXT        NOT NULL DEFAULT '',
    settings    JSONB       NOT NULL DEFAULT '{}',
    tags        TEXT[]      NOT NULL DEFAULT '{}',
    created_by  TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE servers (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    hostname      TEXT        NOT NULL UNIQUE,
    ip_address    TEXT        NOT NULL DEFAULT '',
    vendor        TEXT        NOT NULL DEFAULT '',
    model         TEXT        NOT NULL DEFAULT '',
    serial_number TEXT        NOT NULL DEFAULT '',
    bios_version  TEXT        NOT NULL DEFAULT '',
    bmc_version   TEXT        NOT NULL DEFAULT '',
    profile_id    UUID        REFERENCES bios_profiles(id) ON DELETE SET NULL,
    status        TEXT        NOT NULL DEFAULT 'active',
    last_seen     TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON servers(status);
CREATE INDEX ON servers(profile_id);

CREATE TABLE bios_settings (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id     UUID        NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    settings      JSONB       NOT NULL DEFAULT '{}',
    agent_version TEXT        NOT NULL DEFAULT '',
    collected_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON bios_settings(server_id);
CREATE INDEX ON bios_settings(collected_at DESC);

CREATE TABLE firmware_inventory (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id        UUID        NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    component        TEXT        NOT NULL,
    current_version  TEXT        NOT NULL DEFAULT '',
    approved_version TEXT        NOT NULL DEFAULT '',
    status           TEXT        NOT NULL DEFAULT 'unknown',
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(server_id, component)
);

CREATE INDEX ON firmware_inventory(server_id);

CREATE TABLE drift_events (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id      UUID        NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    setting_key    TEXT        NOT NULL,
    expected_value JSONB,
    actual_value   JSONB,
    detected_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at    TIMESTAMPTZ,
    status         TEXT        NOT NULL DEFAULT 'open'
);

CREATE INDEX ON drift_events(server_id);
CREATE INDEX ON drift_events(status);
CREATE INDEX ON drift_events(detected_at DESC);

CREATE TABLE change_requests (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id    UUID        REFERENCES servers(id) ON DELETE CASCADE,
    profile_id   UUID        REFERENCES bios_profiles(id) ON DELETE SET NULL,
    requested_by TEXT        NOT NULL,
    type         TEXT        NOT NULL,
    payload      JSONB       NOT NULL DEFAULT '{}',
    status       TEXT        NOT NULL DEFAULT 'pending',
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_by  TEXT,
    reviewed_at  TIMESTAMPTZ,
    applied_at   TIMESTAMPTZ
);

CREATE INDEX ON change_requests(status);
CREATE INDEX ON change_requests(server_id);

CREATE TABLE snapshots (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id   UUID        NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    settings    JSONB       NOT NULL DEFAULT '{}',
    firmware    JSONB       NOT NULL DEFAULT '{}',
    taken_by    TEXT        NOT NULL DEFAULT '',
    taken_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    description TEXT        NOT NULL DEFAULT ''
);

CREATE INDEX ON snapshots(server_id);
CREATE INDEX ON snapshots(taken_at DESC);

CREATE TABLE compliance_reports (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    server_id    UUID        NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    profile_id   UUID        NOT NULL REFERENCES bios_profiles(id) ON DELETE CASCADE,
    compliant    BOOLEAN     NOT NULL,
    violations   JSONB       NOT NULL DEFAULT '[]',
    score        FLOAT       NOT NULL DEFAULT 100.0,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON compliance_reports(server_id);
CREATE INDEX ON compliance_reports(generated_at DESC);

CREATE TABLE jobs (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    type         TEXT        NOT NULL,
    target       JSONB       NOT NULL DEFAULT '{}',
    status       TEXT        NOT NULL DEFAULT 'queued',
    progress     INT         NOT NULL DEFAULT 0,
    result       JSONB,
    created_by   TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at   TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX ON jobs(status);
CREATE INDEX ON jobs(created_at DESC);
