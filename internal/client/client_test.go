package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSign(t *testing.T) {
	// Successful response
	signServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/sign" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if body["csr"] != "testcsr" || body["ott"] != "token" {
			t.Fatalf("unexpected body: %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"crt": "CERTPEM"})
	}))
	defer signServer.Close()

	c := New(signServer.URL, "token")
	c.httpClient = signServer.Client()
	cert, err := c.Sign(context.Background(), "testcsr")
	if err != nil {
		t.Fatalf("Sign returned error: %v", err)
	}
	if string(cert) != "CERTPEM" {
		t.Fatalf("unexpected cert: %s", cert)
	}

	// Error response handling
	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errServer.Close()

	c = New(errServer.URL, "token")
	c.httpClient = errServer.Client()
	if _, err := c.Sign(context.Background(), "badcsr"); err == nil {
		t.Fatal("expected error from Sign")
	}
}

func TestClientVersion(t *testing.T) {
	versionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/version" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("1.2.3\n"))
	}))
	defer versionServer.Close()

	c := New(versionServer.URL, "token")
	c.httpClient = versionServer.Client()
	v, err := c.Version(context.Background())
	if err != nil {
		t.Fatalf("Version returned error: %v", err)
	}
	if v != "1.2.3" {
		t.Fatalf("unexpected version: %s", v)
	}

	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer errServer.Close()

	c = New(errServer.URL, "token")
	c.httpClient = errServer.Client()
	if _, err := c.Version(context.Background()); err == nil {
		t.Fatal("expected error from Version")
	}
}

func TestClientRootCertificate(t *testing.T) {
	rootServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/root" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ROOTPEM"))
	}))
	defer rootServer.Close()

	c := New(rootServer.URL, "token")
	c.httpClient = rootServer.Client()
	cert, err := c.RootCertificate(context.Background())
	if err != nil {
		t.Fatalf("RootCertificate returned error: %v", err)
	}
	if string(cert) != "ROOTPEM" {
		t.Fatalf("unexpected cert: %s", cert)
	}

	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errServer.Close()

	c = New(errServer.URL, "token")
	c.httpClient = errServer.Client()
	if _, err := c.RootCertificate(context.Background()); err == nil {
		t.Fatal("expected error from RootCertificate")
	}
}
