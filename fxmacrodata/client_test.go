package fxmacrodata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildURL(t *testing.T) {
	client := NewClient("test-key")
	got := client.buildURL("/forex/aud/usd", nil)
	want := "https://api.fxmacrodata.com/v1/forex/aud/usd"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestForexPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/forex/eur/usd" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.URL.RawQuery != "limit=100&offset=200" {
			t.Errorf("unexpected query %q", r.URL.RawQuery)
		}
		w.Write([]byte("{}"))
	}))
	defer srv.Close()
	client := NewClient("")
	client.BaseURL = srv.URL
	if _, err := client.ForexPage(context.Background(), "EUR", "USD", Page{Limit: 100, Offset: 200}); err != nil {
		t.Fatal(err)
	}
}

func TestAPIKeyHeader(t *testing.T) {
	for _, key := range []string{"test-key", ""} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.RawQuery != "" {
				t.Errorf("unexpected query %q", r.URL.RawQuery)
			}
			got, ok := r.Header["X-Api-Key"]
			if key == "" && ok {
				t.Errorf("unexpected X-API-Key header %q", got)
			}
			if key != "" && r.Header.Get("X-API-Key") != key {
				t.Errorf("got X-API-Key %q want %q", r.Header.Get("X-API-Key"), key)
			}
			w.Write([]byte("{}"))
		}))
		client := NewClient(key)
		client.BaseURL = srv.URL
		if _, err := client.MarketSessions(context.Background()); err != nil {
			t.Fatal(err)
		}
		srv.Close()
	}
}
