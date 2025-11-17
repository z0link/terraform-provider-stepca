package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
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

func TestClientCertificate(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/certificates/abc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Write([]byte("PEM"))
	})
	mux.HandleFunc("/certificates/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/certificates/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New(srv.URL, "token")
	c.httpClient = srv.Client()

	cert, found, err := c.Certificate(context.Background(), "abc")
	if err != nil {
		t.Fatalf("Certificate returned error: %v", err)
	}
	if !found {
		t.Fatalf("expected certificate to be found")
	}
	if string(cert) != "PEM" {
		t.Fatalf("unexpected certificate: %s", cert)
	}

	_, found, err = c.Certificate(context.Background(), "missing")
	if err != nil {
		t.Fatalf("missing certificate returned error: %v", err)
	}
	if found {
		t.Fatalf("expected missing certificate to report not found")
	}

	if _, _, err := c.Certificate(context.Background(), "error"); err == nil {
		t.Fatalf("expected error from Certificate")
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
			if !p.Admin {
				w.WriteHeader(http.StatusOK)
				return
			}
			t.Fatalf("unexpected payload during update: %#v", p)
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

	p, err := c.GetProvisioner(context.Background(), "test")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if p == nil || p.Name != "test" || p.Type != "JWK" || !p.Admin {
		t.Fatalf("unexpected provisioner %#v", p)
	}

	if err := c.ReplaceProvisioner(context.Background(), "test", Provisioner{Name: "test", Type: "JWK", Admin: false}); err != nil {
		t.Fatalf("replace failed: %v", err)
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
			if a.Provisioner != "operators" {
				t.Fatalf("unexpected admin payload: %#v", a)
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

	a, err := c.GetAdmin(context.Background(), "alice", "admin")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if a == nil || a.Name != "alice" || a.Provisioner != "admin" {
		t.Fatalf("unexpected admin %#v", a)
	}

	if err := c.ReplaceAdmin(context.Background(), "alice", "admin", Admin{Name: "alice", Provisioner: "operators"}); err != nil {
		t.Fatalf("replace failed: %v", err)
	}

	if err := c.DeleteAdmin(context.Background(), "alice", "admin"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}

func TestClientListProvisioners(t *testing.T) {
	fixture, err := os.ReadFile("testdata/provisioners.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}
	var expected []Provisioner
	if err := json.Unmarshal(fixture, &expected); err != nil {
		t.Fatalf("failed to unmarshal fixture: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/admin/provisioners" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(fixture)
	}))
	defer srv.Close()

	c := New(srv.URL, "token").WithAdminToken("admin")
	c.httpClient = srv.Client()

	got, err := c.ListProvisioners(context.Background())
	if err != nil {
		t.Fatalf("ListProvisioners returned error: %v", err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("unexpected provisioners: %#v", got)
	}

	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errSrv.Close()

	c = New(errSrv.URL, "token").WithAdminToken("admin")
	c.httpClient = errSrv.Client()
	if _, err := c.ListProvisioners(context.Background()); err == nil {
		t.Fatal("expected error from ListProvisioners")
	}
}

func TestClientTemplate(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/admin/templates", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var tmpl Template
		if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
			t.Fatalf("decode error: %v", err)
		}
		if tmpl.Name != "example" || tmpl.Body != "BODY" {
			t.Fatalf("unexpected template: %#v", tmpl)
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("/admin/templates/example", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(Template{Name: "example", Body: "BODY", Metadata: map[string]string{"type": "x509"}})
		case http.MethodPut:
			var tmpl Template
			if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if tmpl.Metadata["version"] != "2" {
				t.Fatalf("unexpected metadata: %#v", tmpl.Metadata)
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method: %s", r.Method)
		}
	})
	mux.HandleFunc("/admin/templates/missing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New(srv.URL, "ott").WithAdminToken("adm")
	c.httpClient = srv.Client()

	if err := c.CreateTemplate(context.Background(), Template{Name: "example", Body: "BODY"}); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	tmpl, err := c.GetTemplate(context.Background(), "example")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if tmpl == nil || tmpl.Name != "example" || tmpl.Metadata["type"] != "x509" {
		t.Fatalf("unexpected template: %#v", tmpl)
	}

	if err := c.UpdateTemplate(context.Background(), Template{Name: "example", Body: "BODY", Metadata: map[string]string{"version": "2"}}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if err := c.DeleteTemplate(context.Background(), "example"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	missing, err := c.GetTemplate(context.Background(), "missing")
	if err != nil {
		t.Fatalf("missing get failed: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil template, got %#v", missing)
	}

	if err := c.DeleteTemplate(context.Background(), "missing"); err != nil {
		t.Fatalf("delete missing failed: %v", err)
	}
}
