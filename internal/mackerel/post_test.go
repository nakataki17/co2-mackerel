package mackerel

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPostServiceMetrics_success(t *testing.T) {
	var gotMethod, gotCT, gotKey string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotCT = r.Header.Get("Content-Type")
		gotKey = r.Header.Get("X-Api-Key")
		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	metrics := []ServiceMetric{
		{Name: "co2.living.ppm", Time: 1700000000, Value: 100},
	}
	err := PostServiceMetrics(context.Background(), client, srv.URL, "svc", "k", metrics)
	if err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method: %s", gotMethod)
	}
	if gotCT != "application/json" {
		t.Fatalf("Content-Type: %s", gotCT)
	}
	if gotKey != "k" {
		t.Fatalf("X-Api-Key: %q", gotKey)
	}
	if len(gotBody) == 0 {
		t.Fatal("empty body")
	}
}

func TestPostServiceMetrics_retry_503(t *testing.T) {
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		if n < 2 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	err := PostServiceMetrics(context.Background(), client, srv.URL, "svc", "k", []ServiceMetric{
		{Name: "m", Time: 1, Value: 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("attempts: %d", n)
	}
}

func TestTsdbEndpoint(t *testing.T) {
	u, err := tsdbEndpoint("https://api.mackerelio.com", "environmental-sensors")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://api.mackerelio.com/api/v0/services/environmental-sensors/tsdb"
	if u != want {
		t.Fatalf("got %q want %q", u, want)
	}
}
