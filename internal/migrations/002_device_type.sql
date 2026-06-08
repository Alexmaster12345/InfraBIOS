-- Add device_type and os columns to servers table
ALTER TABLE servers
    ADD COLUMN IF NOT EXISTS device_type TEXT NOT NULL DEFAULT 'server',
    ADD COLUMN IF NOT EXISTS os          TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS servers_device_type_idx ON servers(device_type);
