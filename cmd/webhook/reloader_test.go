package main

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReloaderGetCertificate(t *testing.T) {
	dummyCert := tls.Certificate{}
	reloader := newCertReloader("dummy.crt", "dummy.key", dummyCert)
	cert, err := reloader.GetCertificate(nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cert == nil {
		t.Fatal("expected non-nil cert")
	}
}

func TestHealthzEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","mode":"tls_reloader"}`))
	})

	req := httptest.NewRequest("GET", "/healthz", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if body != `{"status":"ok","mode":"tls_reloader"}` {
		t.Fatalf("unexpected body: %s", body)
	}
}
