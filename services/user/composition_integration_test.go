//go:build integration

package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/nostosx/servex-examples/application/notes"
	"github.com/nostosx/servex-examples/infrastructure/config"
)

// TestIntegrationComposition requires disposable, application-owned services.
// It never creates or drops databases or buckets, and cleans up its own records.
func TestIntegrationComposition(t *testing.T) {
	dsn := os.Getenv("SERVEX_EXAMPLES_IT_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("SERVEX_EXAMPLES_IT_POSTGRES_DSN not configured")
	}
	required := func(key string) string {
		t.Helper()
		value := os.Getenv("SERVEX_EXAMPLES_IT_" + key)
		if value == "" {
			t.Fatalf("SERVEX_EXAMPLES_IT_%s required", key)
		}
		return value
	}
	cfg := workspaceConfig(t)
	cfg.Composition.PostgresDSN = dsn
	cfg.Composition.RedisAddress = required("REDIS_ADDR")
	cfg.Composition.Namespace = "it-" + uuid.NewV7().String() + ":"
	cfg.Composition.S3 = config.S3{Endpoint: required("S3_ENDPOINT"), Bucket: required("S3_BUCKET"), Region: "us-east-1", AccessKey: required("S3_ACCESS_KEY"), SecretKey: required("S3_SECRET_KEY")}
	cfg.Composition.EtcdEndpoints = strings.Split(required("ETCD_ENDPOINTS"), ",")
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	objects := awss3.New(awss3.Options{Region: "us-east-1", BaseEndpoint: aws.String(cfg.Composition.S3.Endpoint), UsePathStyle: true, HTTPClient: &http.Client{Timeout: 10 * time.Second}, Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(cfg.Composition.S3.AccessKey, cfg.Composition.S3.SecretKey, ""))})
	if _, err := objects.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(cfg.Composition.S3.Bucket)}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := objects.DeleteBucket(cleanup, &awss3.DeleteBucketInput{Bucket: aws.String(cfg.Composition.S3.Bucket)}); err != nil {
			t.Error(err)
		}
	})
	if err := Migrate(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	first, err := Build(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeRuntime(t, first) })
	if err := first.App.Start(ctx); err != nil {
		t.Fatal(err)
	}
	second, err := Build(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeRuntime(t, second) })
	if err := second.App.Start(ctx); err != nil {
		t.Fatal(err)
	}
	one := first.Servers.HTTP.URL().String()
	two := second.Servers.HTTP.URL().String()
	client := workspaceClient(t)
	signIn(t, client, one)
	originalCookies := client.Jar.Cookies(mustURL(t, one))
	request(t, client, http.MethodGet, two+"/v1/me", "", http.StatusOK)
	created := request(t, client, http.MethodPost, one+"/v1/notes", `{"text":"persisted by GORM"}`, http.StatusCreated)
	var note notes.Note
	if err := json.Unmarshal(created, &note); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Restore a live session to remove this test-owned row, even after logout assertions.
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		requestContext(t, cleanup, client, http.MethodPost, two+"/auth/password", `{"username":"demo","password":"test-only-password"}`, http.StatusOK)
		requestContext(t, cleanup, client, http.MethodDelete, two+"/v1/notes/"+note.ID, "", http.StatusNoContent)
	})
	listing := request(t, client, http.MethodGet, two+"/v1/notes", "", http.StatusOK)
	if !strings.Contains(string(listing), note.ID) {
		t.Fatalf("second instance did not read SQL row: %s", listing)
	}
	request(t, client, http.MethodPut, two+"/v1/notes/"+note.ID, `{"text":"updated from second instance"}`, http.StatusNoContent)
	listing = request(t, client, http.MethodGet, one+"/v1/notes", "", http.StatusOK)
	if !strings.Contains(string(listing), "updated from second instance") {
		t.Fatal(string(listing))
	}

	request(t, client, http.MethodPost, one+"/v1/jobs/refresh", "", http.StatusAccepted)
	deadline := time.NewTimer(20 * time.Second)
	defer deadline.Stop()
	for {
		body := request(t, client, http.MethodGet, two+"/v1/jobs/last-run", "", http.StatusOK)
		var state struct {
			At string `json:"at"`
		}
		if err := json.Unmarshal(body, &state); err != nil {
			t.Fatal(err)
		}
		if state.At != "" {
			if _, err := time.Parse(time.RFC3339Nano, state.At); err != nil {
				t.Fatal(err)
			}
			break
		}
		select {
		case <-deadline.C:
			t.Fatal("background worker did not update shared Redis cache")
		case <-time.After(25 * time.Millisecond):
		}
	}
	uploaded := request(t, client, http.MethodPost, one+"/v1/objects", "stored in S3", http.StatusCreated)
	var object struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(uploaded, &object); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := objects.DeleteObject(cleanup, &awss3.DeleteObjectInput{Bucket: aws.String(cfg.Composition.S3.Bucket), Key: aws.String(cfg.Composition.Namespace + "objects/" + object.Key)}); err != nil {
			t.Error(err)
		}
	})
	page := request(t, client, http.MethodGet, two+"/v1/objects", "", http.StatusOK)
	var listed struct {
		Keys []string `json:"keys"`
	}
	if err := json.Unmarshal(page, &listed); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, key := range listed.Keys {
		if key != object.Key {
			continue
		}
		found = true
		data := request(t, client, http.MethodGet, two+"/v1/objects/content?key="+url.QueryEscape(key), "", http.StatusOK)
		if string(data) != "stored in S3" {
			t.Fatalf("listed key did not roundtrip: %q", data)
		}
	}
	if !found {
		t.Fatalf("S3 listing omitted upload: %s", page)
	}
	probe := request(t, client, http.MethodPost, two+"/v1/coordination/probe", "", http.StatusOK)
	var lock struct {
		Revision int64 `json:"revision"`
	}
	if err := json.Unmarshal(probe, &lock); err != nil {
		t.Fatal(err)
	}
	if lock.Revision <= 0 {
		t.Fatalf("invalid fencing revision: %s", probe)
	}
	request(t, client, http.MethodDelete, two+"/v1/objects/content?key="+url.QueryEscape(object.Key), "", http.StatusNoContent)
	request(t, client, http.MethodPost, two+"/auth/logout", "", http.StatusNoContent)
	client.Jar.SetCookies(mustURL(t, one), originalCookies)
	request(t, client, http.MethodGet, one+"/v1/me", "", http.StatusUnauthorized)
}
