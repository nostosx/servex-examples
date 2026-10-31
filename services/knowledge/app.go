// Package knowledge assembles an owned PostgreSQL indexing demonstration.
package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	pg "github.com/nostosx/servex-contrib/database/postgres"
	"github.com/nostosx/servex-contrib/jobqueue/pgmq"
	searchpg "github.com/nostosx/servex-contrib/search/postgres"
	vectorpg "github.com/nostosx/servex-contrib/vector/pgvector"
	"github.com/nostosx/servex-examples/infrastructure/config"
	"github.com/nostosx/servex-examples/services/knowledge/application"
	port "github.com/nostosx/servex-examples/services/knowledge/internal/port/postgres"
	"github.com/nostosx/servex/endpoint"
	"github.com/nostosx/servex/errorx"
	"github.com/nostosx/servex/messaging/jobqueue"
	"github.com/nostosx/servex/storage/search"
	"github.com/nostosx/servex/storage/vector"
)

type Result struct {
	VectorHits int `json:"vector_hits"`
	TextHits   int `json:"text_hits"`
}

// RunPostgresDemo explicitly creates only the configured demo resources and
// leaves them in place for inspection. Database extensions must already exist.
func RunPostgresDemo(ctx context.Context, db *sql.DB, cfg config.PostgresDemo) (result Result, err error) {
	if err := cfg.Validate(); err != nil {
		return result, err
	}
	schema, err := pg.Ident(cfg.Schema)
	if err != nil {
		return result, err
	}
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+schema); err != nil {
		return result, err
	}
	sources := pg.Table{Schema: cfg.Schema, Name: "sources"}
	table, err := sources.SQL()
	if err != nil {
		return result, err
	}
	if _, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS "+table+" (id text PRIMARY KEY,body text NOT NULL)"); err != nil {
		return result, err
	}
	vectors, err := vectorpg.New(db, vectorpg.Config{Namespace: cfg.Namespace, Dimensions: 3, Table: pg.Table{Schema: cfg.Schema, Name: "vectors"}})
	if err != nil {
		return result, err
	}
	text, err := searchpg.New(db, searchpg.Config{Namespace: cfg.Namespace, Engine: searchpg.Engine(cfg.SearchEngine), Table: pg.Table{Schema: cfg.Schema, Name: "documents"}})
	if err != nil {
		return result, err
	}
	queue, err := pgmq.New(db, pgmq.Config{Queue: cfg.Schema, Reservations: pg.Table{Schema: cfg.Schema, Name: "reservations"}})
	if err != nil {
		return result, err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = errors.Join(err, queue.Close(cleanup))
	}()
	if err := vectors.Migrate(ctx); err != nil {
		return result, err
	}
	if err := text.Migrate(ctx); err != nil {
		return result, err
	}
	if err := queue.Migrate(ctx); err != nil {
		return result, err
	}
	// Index creation is explicit. Native search also works without this optional index.
	var indexed bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass($1) IS NOT NULL`, schema+`."documents_search"`).Scan(&indexed); err != nil {
		return result, err
	}
	if !indexed {
		if err := text.CreateIndex(ctx, "documents_search"); err != nil {
			return result, err
		}
	}
	index := &port.Index{DB: db, Vectors: vectors, Text: text, Queue: queue, Sources: sources}
	worker := jobqueue.NewWorker(queue, map[string]endpoint.Endpoint{"index": func(ctx context.Context, request any) (any, error) {
		job := request.(*jobqueue.Job)
		var d application.Document
		if err := json.Unmarshal(job.Payload, &d); err != nil {
			return nil, errorx.New(errorx.InvalidArgument, "DOCUMENT", "invalid document").WithCause(err)
		}
		return nil, index.Write(ctx, d)
	}}, jobqueue.WithName("postgres-index"))
	if err := worker.Start(ctx); err != nil {
		return result, err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = errors.Join(err, worker.Stop(cleanup))
	}()
	if err := index.Submit(ctx, application.Document{ID: "postgres", Text: "PostgreSQL document search", Embedding: []float32{1, 0, 0}}); err != nil {
		return result, err
	}
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	var failure <-chan error
	if source, ok := worker.(interface{ Err() <-chan error }); ok {
		failure = source.Err()
	}
	for {
		length, err := queue.Length(ctx)
		if err != nil {
			return result, err
		}
		if length == 0 {
			break
		}
		select {
		case err := <-failure:
			if err == nil {
				err = errors.New("indexing worker stopped before settlement")
			}
			return result, err
		case <-ctx.Done():
			return result, ctx.Err()
		case <-ticker.C:
		}
	}
	v, err := vectors.Search(ctx, vector.Query{Vector: []float32{1, 0, 0}, Limit: 10})
	if err != nil {
		return result, err
	}
	hits, err := text.Search(ctx, search.Query{Text: "PostgreSQL", Limit: 10})
	if err != nil {
		return result, err
	}
	result = Result{VectorHits: len(v), TextHits: len(hits)}
	if result.VectorHits != 1 || result.TextHits != 1 {
		return result, errors.New("indexed document was not retrieved")
	}
	return result, nil
}
