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
