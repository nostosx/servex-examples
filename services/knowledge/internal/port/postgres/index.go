// Package postgres implements the sample's transactional persistence boundary.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	pg "github.com/nostosx/servex-contrib/database/postgres"
	"github.com/nostosx/servex-contrib/jobqueue/pgmq"
	searchpg "github.com/nostosx/servex-contrib/search/postgres"
	vectorpg "github.com/nostosx/servex-contrib/vector/pgvector"
	"github.com/nostosx/servex-examples/services/knowledge/application"
	"github.com/nostosx/servex/messaging/jobqueue"
	"github.com/nostosx/servex/storage/search"
	"github.com/nostosx/servex/storage/vector"
)

type Index struct {
	DB      *sql.DB
	Vectors *vectorpg.Store
	Text    *searchpg.Index
	Queue   *pgmq.Queue
	Sources pg.Table
}

// Submit commits the business document and its indexing job together.
func (s *Index) Submit(ctx context.Context, d application.Document) error {
	payload, err := json.Marshal(d)
	if err != nil {
		return err
	}
	table, err := s.Sources.SQL()
	if err != nil {
		return err
	}
	return pg.Transaction(ctx, s.DB, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "INSERT INTO "+table+" (id,body) VALUES($1,$2) ON CONFLICT(id) DO UPDATE SET body=EXCLUDED.body", d.ID, d.Text); err != nil {
			return err
		}
		_, err := s.Queue.EnqueueTx(ctx, tx, &jobqueue.Job{Type: "index", Payload: payload}, jobqueue.WithUnique(d.ID, 0))
		return err
	})
}

// Write makes text/vector index updates atomic; repeated deliveries upsert the
// same ID. A real application also checks document revisions before replacement.
func (s *Index) Write(ctx context.Context, d application.Document) error {
	return pg.Transaction(ctx, s.DB, func(tx *sql.Tx) error {
		if err := s.Vectors.UpsertTx(ctx, tx, vector.Record{ID: d.ID, Vector: d.Embedding, Payload: []byte(d.Text)}); err != nil {
			return err
		}
		return s.Text.UpsertTx(ctx, tx, search.Document{ID: d.ID, Text: d.Text})
	})
}
