package config

import "testing"

func TestPostgresDemoConfiguration(t *testing.T) {
	t.Setenv("SERVEX_DEMO_POSTGRES_DSN", "postgres://fixture")
	for _, schema := range []string{"", "public", "sx_demo_bad-name", "sx_demo_abcdefghijklmnopqrstuvwxyz"} {
		t.Setenv("SERVEX_DEMO_POSTGRES_SCHEMA", schema)
		if _, err := LoadPostgresDemo(); err == nil {
			t.Fatalf("accepted schema %q", schema)
		}
	}
	t.Setenv("SERVEX_DEMO_POSTGRES_SCHEMA", "sx_demo_test")
	c, err := LoadPostgresDemo()
	if err != nil || c.SearchEngine != "native" || c.Namespace != "demo" {
		t.Fatalf("%+v: %v", c, err)
	}
}
