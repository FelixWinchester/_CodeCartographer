DROP INDEX IF EXISTS idx_llm_cache_node;
DROP INDEX IF EXISTS idx_metrics_node_id;
DROP INDEX IF EXISTS idx_edges_to_id;
DROP INDEX IF EXISTS idx_edges_from_id;
DROP INDEX IF EXISTS idx_nodes_file_path;
DROP INDEX IF EXISTS idx_nodes_kind;
DROP INDEX IF EXISTS idx_nodes_repo;

DROP TABLE IF EXISTS llm_cache;
DROP TABLE IF EXISTS metrics;
DROP TABLE IF EXISTS edges;
DROP TABLE IF EXISTS nodes;

DROP EXTENSION IF EXISTS vector;
