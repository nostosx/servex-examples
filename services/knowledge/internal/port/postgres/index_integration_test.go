//go:build integration

package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	pg "github.com/nostosx/servex-contrib/database/postgres"
	"github.com/nostosx/servex-contrib/jobqueue/pgmq"
	searchpg "github.com/nostosx/servex-contrib/search/postgres"
	vectorpg "github.com/nostosx/servex-contrib/vector/pgvector"
	"github.com/nostosx/servex-examples/services/knowledge/application"
	"github.com/nostosx/servex/messaging/jobqueue"
	"github.com/nostosx/servex/storage/search"
	"github.com/nostosx/servex/storage/vector"
)

func TestIntegrationTransactionalIndex(t *testing.T) {
	dsn := os.Getenv("SERVEX_EXAMPLES_IT_INDEX_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set SERVEX_EXAMPLES_IT_INDEX_POSTGRES_DSN to an owned extension fixture")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(8)
	schema := "sx_demo_" + strings.ToLower(rand.Text())[:16]
	ident, err := pg.Ident(schema)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), "CREATE SCHEMA "+ident); err != nil {
		t.Fatal(err)
	}
	sources := pg.Table{Schema: schema, Name: "sources"}
	table, err := sources.SQL()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), "CREATE TABLE "+table+"(id text PRIMARY KEY,body text NOT NULL)"); err != nil {
		t.Fatal(err)
	}
	v, err := vectorpg.New(db, vectorpg.Config{Namespace: "demo", Dimensions: 3, Table: pg.Table{Schema: schema, Name: "vectors"}})
	if err != nil {
		t.Fatal(err)
	}
	text, err := searchpg.New(db, searchpg.Config{Namespace: "demo", Table: pg.Table{Schema: schema, Name: "documents"}})
	if err != nil {
		t.Fatal(err)
	}
	queue, err := pgmq.New(db, pgmq.Config{Queue: schema, Reservations: pg.Table{Schema: schema, Name: "reservations"}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := queue.Close(ctx); err != nil {
			t.Error(err)
		}
		if _, err := db.ExecContext(ctx, "SELECT pgmq.drop_queue($1)", schema); err != nil {
			t.Error(err)
		}
		if _, err := db.ExecContext(ctx, "DROP SCHEMA "+ident+" CASCADE"); err != nil {
			t.Error(err)
		}
	}()
	for _, migrate := range []func(context.Context) error{v.Migrate, text.Migrate, queue.Migrate} {
		if err := migrate(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
	index := &Index{DB: db, Vectors: v, Text: text, Queue: queue, Sources: sources}
	d := application.Document{ID: "doc", Text: "PostgreSQL original", Embedding: []float32{1, 0, 0}}
	if err := index.Submit(t.Context(), d); err != nil {
		t.Fatal(err)
	}
	d.Text = "PostgreSQL must rollback"
	if err := index.Submit(t.Context(), d); !errors.Is(err, jobqueue.ErrDuplicate) {
		t.Fatalf("duplicate: %v", err)
	}
	var body string
	if err := db.QueryRowContext(t.Context(), "SELECT body FROM "+table+" WHERE id=$1", d.ID).Scan(&body); err != nil || body != "PostgreSQL original" {
		t.Fatalf("source rollback: %s: %v", body, err)
	}
	if length, err := queue.Length(t.Context()); err != nil || length != 1 {
		t.Fatalf("queue rollback: %d: %v", length, err)
	}
	d.Text = body
	for range 2 {
		if err := index.Write(t.Context(), d); err != nil {
			t.Fatal(err)
		}
	}
	hits, err := text.Search(t.Context(), search.Query{Text: "PostgreSQL", Limit: 5})
	if err != nil || len(hits) != 1 {
		t.Fatalf("idempotent: %+v: %v", hits, err)
	}
	// Force the second write to fail after the vector upsert, verifying rollback.
	if _, err := db.ExecContext(t.Context(), "ALTER TABLE "+ident+".documents ADD CONSTRAINT no_rejected CHECK(body <> 'rejected')"); err != nil {
		t.Fatal(err)
	}
	d.Text = "rejected"
	d.Embedding = []float32{0, 1, 0}
	if err := index.Write(t.Context(), d); err == nil {
		t.Fatal("expected text constraint failure")
	}
	vectors, err := v.Search(t.Context(), vector.Query{Vector: []float32{1, 0, 0}, Limit: 1})
	if err != nil || len(vectors) != 1 || vectors[0].Distance != 0 || string(vectors[0].Record.Payload) != body {
		t.Fatalf("vector rollback: %+v: %v", vectors, err)
	}
}
