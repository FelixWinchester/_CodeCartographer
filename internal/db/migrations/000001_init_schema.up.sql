-- Расширение для работы с векторами (нужно для similarity search в блоке 7)
CREATE EXTENSION IF NOT EXISTS vector;

-- Узлы графа: пакеты, файлы, функции
CREATE TABLE nodes (
    id          BIGSERIAL PRIMARY KEY,
    repo_path   TEXT NOT NULL,                -- путь к репозиторию
    kind        TEXT NOT NULL,                -- "package" | "file" | "function"
    name        TEXT NOT NULL,                -- имя узла
    file_path   TEXT NOT NULL,                -- путь к файлу относительно репозитория
    start_line  INT,                          -- строка начала (для функций)
    end_line    INT,                          -- строка конца (для функций)
    embedding   vector(1536),                 -- вектор для similarity search
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Рёбра графа: зависимости между узлами
CREATE TABLE edges (
    id          BIGSERIAL PRIMARY KEY,
    repo_path   TEXT NOT NULL,
    from_id     BIGINT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    to_id       BIGINT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL,                -- "import" | "call"
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Git метрики по файлам
CREATE TABLE file_metrics (
    id          BIGSERIAL PRIMARY KEY,
    repo_path   TEXT NOT NULL,
    file_path   TEXT NOT NULL,
    churn_rate  INT NOT NULL DEFAULT 0,       -- сколько раз файл менялся
    owner       TEXT,                         -- кто менял больше всего
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (repo_path, file_path)
);

-- Change coupling: файлы которые всегда меняются вместе
CREATE TABLE change_coupling (
    id          BIGSERIAL PRIMARY KEY,
    repo_path   TEXT NOT NULL,
    file_a      TEXT NOT NULL,
    file_b      TEXT NOT NULL,
    coupling_score FLOAT NOT NULL DEFAULT 0, -- как часто меняются вместе (0..1)
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (repo_path, file_a, file_b)
);

-- Индексы
CREATE INDEX idx_nodes_repo_path ON nodes(repo_path);
CREATE INDEX idx_nodes_kind ON nodes(kind);
CREATE INDEX idx_edges_from_id ON edges(from_id);
CREATE INDEX idx_edges_to_id ON edges(to_id);
CREATE INDEX idx_file_metrics_repo_path ON file_metrics(repo_path);