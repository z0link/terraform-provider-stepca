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

func TestClientProvisioner(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/admin/provisioners", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var p Provisioner
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if p.Name != "test" || p.Type != "JWK" || !p.Admin {
			t.Fatalf("unexpected provisioner: %#v", p)
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("/admin/provisioners/test", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(Provisioner{Name: "test", Type: "JWK", Admin: true})
		case http.MethodPut:
			var p Provisioner
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if p.Name != "test-renamed" || p.Type != "JWK" || p.Admin {
				t.Fatalf("unexpected payload %#v", p)
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New(srv.URL, "tkn").WithAdminToken("adm")
	c.httpClient = srv.Client()

	if err := c.CreateProvisioner(context.Background(), Provisioner{Name: "test", Type: "JWK", Admin: true}); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if err := c.UpdateProvisioner(context.Background(), "test", Provisioner{Name: "test-renamed", Type: "JWK"}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	p, err := c.GetProvisioner(context.Background(), "test")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if p == nil || p.Name != "test" || p.Type != "JWK" || !p.Admin {
		t.Fatalf("unexpected provisioner %#v", p)
	}

	if err := c.DeleteProvisioner(context.Background(), "test"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}

func TestClientAdmin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/admin/admins", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var a Admin
		if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if a.Name != "alice" || a.Provisioner != "admin" {
			t.Fatalf("unexpected admin: %#v", a)
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("/admin/admins/alice", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("provisioner") != "admin" {
			t.Fatalf("missing provisioner")
		}
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(Admin{Name: "alice", Provisioner: "admin"})
		case http.MethodPut:
			var a Admin
			if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if a.Name != "alice" || a.Provisioner != "root" {
				t.Fatalf("unexpected payload %#v", a)
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New(srv.URL, "tkn").WithAdminToken("adm")
	c.httpClient = srv.Client()

	if err := c.CreateAdmin(context.Background(), Admin{Name: "alice", Provisioner: "admin"}); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if err := c.UpdateAdmin(context.Background(), "alice", "admin", Admin{Name: "alice", Provisioner: "root"}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	a, err := c.GetAdmin(context.Background(), "alice", "admin")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if a == nil || a.Name != "alice" || a.Provisioner != "admin" {
		t.Fatalf("unexpected admin %#v", a)
	}

	if err := c.DeleteAdmin(context.Background(), "alice", "admin"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}
