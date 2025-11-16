package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSign(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
		want    []byte
		wantErr bool
	}{
		{
			name: "success",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("unexpected method: %s", r.Method)
				}
				if r.URL.Path != "/sign" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				var body map[string]string
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode error: %v", err)
				}
				if body["csr"] != "testcsr" || body["ott"] != "token" {
					t.Fatalf("unexpected body: %#v", body)
				}
				_ = json.NewEncoder(w).Encode(map[string]string{"crt": "CERTPEM"})
			},
			want: []byte("CERTPEM"),
		},
		{
			name: "http error",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
		{
			name: "decode error",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("not-json"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tc.handler(t, w, r)
			}))
			defer server.Close()

			c := New(server.URL, "token")
			c.httpClient = server.Client()

			got, err := c.Sign(context.Background(), "testcsr")
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != string(tc.want) {
				t.Fatalf("unexpected cert: %s", got)
			}
		})
	}
}

func TestClientVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
		want    string
		wantErr bool
	}{
		{
			name: "success",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/version" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("1.2.3\n"))
			},
			want: "1.2.3",
		},
		{
			name: "http error",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tc.handler(t, w, r)
			}))
			defer server.Close()

			c := New(server.URL, "token")
			c.httpClient = server.Client()

			got, err := c.Version(context.Background())
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected version: %s", got)
			}
		})
	}
}

func TestClientRootCertificate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
		want    []byte
		wantErr bool
	}{
		{
			name: "success",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/root" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				_, _ = w.Write([]byte("ROOT"))
			},
			want: []byte("ROOT"),
		},
		{
			name: "http error",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tc.handler(t, w, r)
			}))
			defer server.Close()

			c := New(server.URL, "token")
			c.httpClient = server.Client()

			got, err := c.RootCertificate(context.Background())
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != string(tc.want) {
				t.Fatalf("unexpected root: %s", got)
			}
		})
	}
}

func TestClientCreateProvisioner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
		wantErr bool
	}{
		{
			name: "success",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("unexpected method: %s", r.Method)
				}
				var p Provisioner
				if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
					t.Fatalf("decode error: %v", err)
				}
				if p.Name != "name" || p.Type != "JWK" || !p.Admin {
					t.Fatalf("unexpected payload: %#v", p)
				}
				w.WriteHeader(http.StatusCreated)
			},
		},
		{
			name: "http error",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tc.handler(t, w, r)
			}))
			defer server.Close()

			c := New(server.URL, "token").WithAdminToken("admin")
			c.httpClient = server.Client()

			err := c.CreateProvisioner(context.Background(), Provisioner{Name: "name", Type: "JWK", Admin: true})
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClientDeleteProvisioner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{name: "success", statusCode: http.StatusNoContent},
		{name: "not found", statusCode: http.StatusNotFound},
		{name: "error", statusCode: http.StatusBadGateway, wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			c := New(server.URL, "token").WithAdminToken("admin")
			c.httpClient = server.Client()

			err := c.DeleteProvisioner(context.Background(), "name")
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClientGetProvisioner(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
		want    *Provisioner
		wantNil bool
		wantErr bool
	}{
		{
			name: "success",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(Provisioner{Name: "name", Type: "JWK", Admin: true})
			},
			want: &Provisioner{Name: "name", Type: "JWK", Admin: true},
		},
		{
			name: "not found",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantNil: true,
		},
		{
			name: "error",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			wantErr: true,
		},
		{
			name: "decode error",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("not-json"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tc.handler(t, w, r)
			}))
			defer server.Close()

			c := New(server.URL, "token").WithAdminToken("admin")
			c.httpClient = server.Client()

			got, err := c.GetProvisioner(context.Background(), "name")
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %#v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected provisioner")
			}
			if *got != *tc.want {
				t.Fatalf("unexpected provisioner: %#v", got)
			}
		})
	}
}

func TestClientCreateAdmin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
		wantErr bool
	}{
		{
			name: "success",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fatalf("unexpected method: %s", r.Method)
				}
				var a Admin
				if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
					t.Fatalf("decode error: %v", err)
				}
				if a.Name != "alice" || a.Provisioner != "admin" {
					t.Fatalf("unexpected payload: %#v", a)
				}
				w.WriteHeader(http.StatusCreated)
			},
		},
		{
			name: "http error",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tc.handler(t, w, r)
			}))
			defer server.Close()

			c := New(server.URL, "token").WithAdminToken("admin")
			c.httpClient = server.Client()

			err := c.CreateAdmin(context.Background(), Admin{Name: "alice", Provisioner: "admin"})
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClientDeleteAdmin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{name: "success", statusCode: http.StatusNoContent},
		{name: "not found", statusCode: http.StatusNotFound},
		{name: "error", statusCode: http.StatusBadRequest, wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
			}))
			defer server.Close()

			c := New(server.URL, "token").WithAdminToken("admin")
			c.httpClient = server.Client()

			err := c.DeleteAdmin(context.Background(), "alice", "admin")
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClientGetAdmin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		handler func(t *testing.T, w http.ResponseWriter, r *http.Request)
		want    *Admin
		wantNil bool
		wantErr bool
	}{
		{
			name: "success",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(Admin{Name: "alice", Provisioner: "admin"})
			},
			want: &Admin{Name: "alice", Provisioner: "admin"},
		},
		{
			name: "not found",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantNil: true,
		},
		{
			name: "error",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			},
			wantErr: true,
		},
		{
			name: "decode error",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("oops"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tc.handler(t, w, r)
			}))
			defer server.Close()

			c := New(server.URL, "token").WithAdminToken("admin")
			c.httpClient = server.Client()

			got, err := c.GetAdmin(context.Background(), "alice", "admin")
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %#v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected admin")
			}
			if *got != *tc.want {
				t.Fatalf("unexpected admin: %#v", got)
			}
		})
	}
}
