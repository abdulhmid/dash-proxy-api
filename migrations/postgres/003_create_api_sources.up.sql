CREATE TABLE IF NOT EXISTS api_sources (
    id UUID PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    base_url TEXT NOT NULL,
    username VARCHAR(255) NOT NULL DEFAULT '',
    auth_headers TEXT DEFAULT '',
    extra_params TEXT DEFAULT '',
    timeout_ms INTEGER NOT NULL DEFAULT 30000,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_api_sources_name ON api_sources(name);
