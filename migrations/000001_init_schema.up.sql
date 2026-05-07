CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS nodes (
    id          BIGSERIAL PRIMARY KEY,
    repo        TEXT        NOT NULL,
    kind        TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    pkg         TEXT        NOT NULL,
    file_path   TEXT        NOT NULL,
    signature   TEXT,
    body_hash   TEXT,
    embedding   vector(1536),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (repo, kind, pkg, name)
);

CREATE TABLE IF NOT EXISTS edges (
    id          BIGSERIAL PRIMARY KEY,
    repo        TEXT        NOT NULL,
    from_id     BIGINT      NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    to_id       BIGINT      NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    kind        TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (from_id, to_id, kind)
);

CREATE TABLE IF NOT EXISTS metrics (
    id              BIGSERIAL PRIMARY KEY,
    node_id         BIGINT      NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    churn_rate      INT         NOT NULL DEFAULT 0,
    complexity      INT         NOT NULL DEFAULT 0,
    ownership       TEXT,
    coupling_score  FLOAT,
    measured_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (node_id)
);

CREATE TABLE IF NOT EXISTS llm_cache (
    id          BIGSERIAL PRIMARY KEY,
    node_id     BIGINT      NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    prompt_hash TEXT        NOT NULL,
    explanation TEXT        NOT NULL,
    model       TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    UNIQUE (node_id, prompt_hash)
);

CREATE INDEX IF NOT EXISTS idx_nodes_repo       ON nodes(repo);
CREATE INDEX IF NOT EXISTS idx_nodes_kind       ON nodes(kind);
CREATE INDEX IF NOT EXISTS idx_nodes_file_path  ON nodes(file_path);
CREATE INDEX IF NOT EXISTS idx_edges_from_id    ON edges(from_id);
CREATE INDEX IF NOT EXISTS idx_edges_to_id      ON edges(to_id);
CREATE INDEX IF NOT EXISTS idx_metrics_node_id  ON metrics(node_id);
CREATE INDEX IF NOT EXISTS idx_llm_cache_node   ON llm_cache(node_id);
