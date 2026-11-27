//go:build integration

package store

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func open(t *testing.T) (*Store, context.Context) {
	t.Helper()
	s, err := Open(env("PG_DSN", "host=localhost user=postgres dbname=s2 sslmode=disable"), env("REDIS_ADDR", "localhost:6379"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := s.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s, ctx
}

func TestRecordRoundTrip(t *testing.T) {
	s, ctx := open(t)
	kind := fmt.Sprintf("rt-%d", time.Now().UnixNano())
	for i := 0; i < 50; i++ {
		if _, err := s.Record(ctx, kind, fmt.Sprintf("payload-%d", i)); err != nil {
			t.Fatal(err)
		}
	}
	db, err := s.CountDB(ctx, kind)
	if err != nil || db != 50 {
		t.Fatalf("db count %d err %v", db, err)
	}
	c, err := s.CountCache(ctx, kind)
	if err != nil || c != 50 {
		t.Fatalf("cache count %d err %v", c, err)
	}
}

func TestConcurrentWriters(t *testing.T) {
	s, ctx := open(t)
	kind := fmt.Sprintf("cc-%d", time.Now().UnixNano())
	var wg sync.WaitGroup
	errs := make(chan error, 16)
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			for i := 0; i < 250; i++ {
				if _, err := s.Record(ctx, kind, fmt.Sprintf("w%d-%d", w, i)); err != nil {
					errs <- err
					return
				}
			}
		}(w)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
	db, _ := s.CountDB(ctx, kind)
	c, _ := s.CountCache(ctx, kind)
	if db != 2000 || c != 2000 {
		t.Fatalf("db=%d cache=%d want 2000", db, c)
	}
}

func TestCacheTTL(t *testing.T) {
	s, ctx := open(t)
	k := fmt.Sprintf("ttl-%d", time.Now().UnixNano())
	if err := s.Cache(ctx, k, "v", 300*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if v, err := s.RDB.Get(ctx, k).Result(); err != nil || v != "v" {
		t.Fatalf("get %q %v", v, err)
	}
	time.Sleep(500 * time.Millisecond)
	if _, err := s.RDB.Get(ctx, k).Result(); err == nil {
		t.Fatal("key should have expired")
	}
}
