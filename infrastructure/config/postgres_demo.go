package config

import (
	"errors"
	"os"
	"regexp"
)

// PostgresDemo is explicit configuration for an owned PostgreSQL demo schema.
type PostgresDemo struct{ DSN, Schema, Namespace, SearchEngine string }

func LoadPostgresDemo() (PostgresDemo, error) {
	c := PostgresDemo{DSN: os.Getenv("SERVEX_DEMO_POSTGRES_DSN"), Schema: os.Getenv("SERVEX_DEMO_POSTGRES_SCHEMA"), Namespace: "demo", SearchEngine: os.Getenv("SERVEX_DEMO_SEARCH_ENGINE")}
	if c.SearchEngine == "" {
		c.SearchEngine = "native"
	}
	return c, c.Validate()
}

// Validate confines the demonstration to its explicitly owned schema.
func (c PostgresDemo) Validate() error {
	if c.DSN == "" || !regexp.MustCompile(`^sx_demo_[a-z0-9_]{1,20}$`).MatchString(c.Schema) {
		return errors.New("set SERVEX_DEMO_POSTGRES_DSN and a dedicated SERVEX_DEMO_POSTGRES_SCHEMA (sx_demo_ prefix)")
	}
	if c.Namespace == "" {
		return errors.New("demo namespace is required")
	}
	return nil
}
