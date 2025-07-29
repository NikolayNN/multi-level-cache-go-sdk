package cache

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestGetAll(t *testing.T) {
	expectedHits := []CacheEntryHit[User]{
		{CacheName: "users", Key: "1", Value: User{ID: 1, Name: "Alice"}, Found: true},
		{CacheName: "users", Key: "2", Value: User{}, Found: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != endpointGetAll {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("content type: %s", ct)
		}
		body, _ := io.ReadAll(r.Body)
		expectedBody := `[{"c":"users","k":"1"},{"c":"users","k":"2"}]`
		if strings.TrimSpace(string(body)) != expectedBody {
			t.Errorf("body = %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedHits)
	}))
	defer server.Close()

	client := New(server.URL)
	hits, err := GetAll[User](client, context.Background(), []CacheId{{"users", "1"}, {"users", "2"}})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if !reflect.DeepEqual(hits, expectedHits) {
		t.Errorf("expected %#v, got %#v", expectedHits, hits)
	}
}

func TestPutAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != endpointPutAll {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		expectedBody := `[{"c":"users","k":"1","v":{"id":1,"name":"Alice"}}]`
		if strings.TrimSpace(string(body)) != expectedBody {
			t.Errorf("body = %s", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(server.URL)
	entries := []CacheEntry[any]{{CacheName: "users", Key: "1", Value: User{1, "Alice"}}}
	if err := client.PutAll(context.Background(), entries); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestEvictAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != endpointEvictAll {
			t.Errorf("path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method: %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		expectedBody := `[{"c":"users","k":"1"}]`
		if strings.TrimSpace(string(body)) != expectedBody {
			t.Errorf("body = %s", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(server.URL)
	if err := client.EvictAll(context.Background(), []CacheId{{"users", "1"}}); err != nil {
		t.Fatalf("error: %v", err)
	}
}
