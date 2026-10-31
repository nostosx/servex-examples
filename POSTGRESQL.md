# Transactional PostgreSQL indexing demo

Snapshot v1.6.0 uses Servex v1.5.0 and database/postgres, vector/pgvector,
search/postgres and jobqueue/pgmq v1.0.0 with public dependencies. The thin
entry is cmd/postgres-demo. Configuration lives in infrastructure/config;
services/knowledge/app.go assembles a servex Worker and its adapters. The
application document is backend-independent; its private PostgreSQL port owns
transactional persistence.

Flow:

1. Explicitly create a dedicated sx_demo_ schema, source table, vector/document
   tables and a PGMQ queue. Extensions must already be installed.
2. Commit the source document and PGMQ job in one *sql.Tx with EnqueueTx.
3. A servex jobqueue.NewWorker consumes the job and atomically upserts both
   indexes in another transaction. Repeated delivery uses the same document ID.
4. Wait for queue settlement and query the vector and text contracts.
5. Print {"vector_hits":1,"text_hits":1}; stop the Worker and close owned resources.

The fixed three-dimensional embedding is deterministic demo data, not evidence
of a live model/embedding provider. No HTTP/gRPC server or cloud account is used.
The program leaves its dedicated schema and queue for inspection; it does not
delete existing application data. Delete only the resources of your owned fixture.

Use the disposable extension fixture documented in servex-contrib/POSTGRESQL.md:

```sh
export SERVEX_DEMO_POSTGRES_DSN='postgres://servex:servex@127.0.0.1:15439/servex_it?sslmode=disable'
export SERVEX_DEMO_POSTGRES_SCHEMA=sx_demo_knowledge
export SERVEX_DEMO_SEARCH_ENGINE=native # Or pg_search/pgroonga/pg_trgm with that extension.
go run ./cmd/postgres-demo
```

## Transaction acceptance

The real suite covers source+enqueue rollback on duplicate reservation, repeated
index delivery and rollback of a vector update when the text write fails.
Set SERVEX_EXAMPLES_IT_INDEX_POSTGRES_DSN to an owned database with vector and
pgmq installed. The composed-application integration also requires its existing
SQL/Redis/S3/etcd variables; see CONTRIBUTING.md.

```sh
export SERVEX_EXAMPLES_IT_INDEX_POSTGRES_DSN="$SERVEX_DEMO_POSTGRES_DSN"
task test INTEGRATION=true
```

The gate rejects skipped, absent, failed or unfinished integration tests.
The command leaves demo resources for inspection; the test cleans its generated
schema and queue. Hosted CI provisions a separate index database.

Index writes are atomic and repeated delivery upserts a stable document ID.
Production applications must enforce revisions and ordering to prevent an older
queued document overwriting a newer one. pg_cron is a persistent SQL scheduler;
recurring enqueue templates are documented in servex-contrib/jobqueue/pgmq
and scheduler/pgcron, not hidden Go callbacks.

中文：业务文档和任务在同一事务中提交，向量与全文索引也在同一事务中更新；重复投递
按稳定 ID 幂等写入。真实验收覆盖重复任务回滚和第二个索引写失败后的跨索引回滚。
生产应用仍需自行实现版本顺序检查。仅使用独立 sx_demo_ schema 和自有测试队列。
