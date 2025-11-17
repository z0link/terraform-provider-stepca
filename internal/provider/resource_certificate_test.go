package provider

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type fakeCertificateClient struct {
	t         *testing.T
	serial    string
	pem       string
	found     bool
	err       error
	signPEM   string
	signErr   error
	signCSR   string
	signCalls int
}

func (f *fakeCertificateClient) Sign(ctx context.Context, csr string) ([]byte, error) {
	f.signCalls++
	f.signCSR = csr
	if f.signErr != nil {
		return nil, f.signErr
	}
	if f.signPEM == "" {
		return nil, fmt.Errorf("sign response not configured")
	}
	return []byte(f.signPEM), nil
}

func (f *fakeCertificateClient) Certificate(ctx context.Context, serial string) ([]byte, bool, error) {
	if f.t != nil && f.serial != "" && serial != f.serial {
		f.t.Fatalf("unexpected serial: %s", serial)
	}
	if f.err != nil {
		return nil, false, f.err
	}
	if !f.found {
		return nil, false, nil
	}
	return []byte(f.pem), true, nil
}

func TestCertificateResourceShouldKeepCertificate(t *testing.T) {
	t.Parallel()

	certOne := testCertificate(t, 1, "one.test")
	certTwo := testCertificate(t, 2, "two.test")

	ctx := context.Background()

	tests := []struct {
		name          string
		resource      *certificateResource
		certValue     types.String
		wantKeep      bool
		wantWarnCount int
	}{
		{
			name:          "missing certificate",
			resource:      &certificateResource{},
			certValue:     types.StringNull(),
			wantKeep:      false,
			wantWarnCount: 1,
		},
		{
			name:          "invalid certificate",
			resource:      &certificateResource{client: &fakeCertificateClient{}},
			certValue:     types.StringValue("not-pem"),
			wantKeep:      false,
			wantWarnCount: 1,
		},
		{
			name:          "no client keeps state",
			resource:      &certificateResource{},
			certValue:     types.StringValue(certOne),
			wantKeep:      true,
			wantWarnCount: 0,
		},
		{
			name: "revoked certificate",
			resource: &certificateResource{
				client: &fakeCertificateClient{
					t:      t,
					serial: serialFromCert(t, certOne),
					found:  false,
				},
			},
			certValue:     types.StringValue(certOne),
			wantKeep:      false,
			wantWarnCount: 1,
		},
		{
			name: "drift detected",
			resource: &certificateResource{
				client: &fakeCertificateClient{
					t:      t,
					serial: serialFromCert(t, certOne),
					found:  true,
					pem:    certTwo,
				},
			},
			certValue:     types.StringValue(certOne),
			wantKeep:      false,
			wantWarnCount: 1,
		},
		{
			name: "certificate matches",
			resource: &certificateResource{
				client: &fakeCertificateClient{
					t:      t,
					serial: serialFromCert(t, certOne),
					found:  true,
					pem:    certOne,
				},
			},
			certValue:     types.StringValue(certOne),
			wantKeep:      true,
			wantWarnCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := certificateResourceModel{Cert: tc.certValue}
			keep, diags := tc.resource.shouldKeepCertificate(ctx, &data)
			if keep != tc.wantKeep {
				t.Fatalf("expected keep=%t got %t", tc.wantKeep, keep)
			}
			warnCount := 0
			for _, d := range diags {
				if d.Severity() == diag.SeverityWarning {
					warnCount++
				}
			}
			if warnCount != tc.wantWarnCount {
				t.Fatalf("expected %d warnings got %d", tc.wantWarnCount, warnCount)
			}
		})
	}
}

func TestCertificateResourceApplyCertificateUpdate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	existingCert := types.StringValue("state-cert")

	tests := []struct {
		name          string
		plan          certificateResourceModel
		state         certificateResourceModel
		client        *fakeCertificateClient
		wantCert      string
		wantNullCert  bool
		wantSignCalls int
		wantErr       bool
	}{
		{
			name: "no change reuses state certificate",
			plan: certificateResourceModel{
				CSR: types.StringValue("csr"),
			},
			state: certificateResourceModel{
				CSR:  types.StringValue("csr"),
				Cert: existingCert,
			},
			client:        &fakeCertificateClient{},
			wantCert:      existingCert.ValueString(),
			wantSignCalls: 0,
		},
		{
			name: "csr change reissues certificate",
			plan: certificateResourceModel{CSR: types.StringValue("new-csr")},
			state: certificateResourceModel{
				CSR:  types.StringValue("old-csr"),
				Cert: existingCert,
			},
			client:        &fakeCertificateClient{signPEM: "new-cert"},
			wantCert:      "new-cert",
			wantSignCalls: 1,
		},
		{
			name: "force rotate enabled reissues",
			plan: certificateResourceModel{
				CSR:         types.StringValue("csr"),
				ForceRotate: types.BoolValue(true),
			},
			state: certificateResourceModel{
				CSR:         types.StringValue("csr"),
				ForceRotate: types.BoolValue(false),
				Cert:        existingCert,
			},
			client:        &fakeCertificateClient{signPEM: "force-cert"},
			wantCert:      "force-cert",
			wantSignCalls: 1,
		},
		{
			name: "force rotate disabled reissues",
			plan: certificateResourceModel{
				CSR:         types.StringValue("csr"),
				ForceRotate: types.BoolValue(false),
			},
			state: certificateResourceModel{
				CSR:         types.StringValue("csr"),
				ForceRotate: types.BoolValue(true),
				Cert:        existingCert,
			},
			client:        &fakeCertificateClient{signPEM: "force-disabled"},
			wantCert:      "force-disabled",
			wantSignCalls: 1,
		},
		{
			name: "sign failure reports diagnostics",
			plan: certificateResourceModel{
				CSR:         types.StringValue("csr"),
				ForceRotate: types.BoolValue(true),
			},
			state: certificateResourceModel{
				CSR:  types.StringValue("csr"),
				Cert: existingCert,
			},
			client:        &fakeCertificateClient{signErr: fmt.Errorf("boom")},
			wantNullCert:  true,
			wantSignCalls: 1,
			wantErr:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resource := &certificateResource{client: tc.client}
			updated, diags := resource.applyCertificateUpdate(ctx, tc.plan, tc.state)
			if tc.wantErr {
				if !diags.HasError() {
					t.Fatalf("expected diagnostics error")
				}
			} else if diags.HasError() {
				t.Fatalf("unexpected diagnostics error: %v", diags)
			}
			if tc.client != nil && tc.client.signCalls != tc.wantSignCalls {
				t.Fatalf("expected %d sign calls got %d", tc.wantSignCalls, tc.client.signCalls)
			}
			if tc.wantNullCert {
				if !updated.Cert.IsNull() {
					t.Fatalf("expected certificate to be null")
				}
				return
			}
			if updated.Cert.ValueString() != tc.wantCert {
				t.Fatalf("expected certificate %q got %q", tc.wantCert, updated.Cert.ValueString())
			}
		})
	}
}

func testCertificate(t *testing.T, serial int64, cn string) string {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		Subject:               pkix.Name{CommonName: cn},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, priv.Public(), priv)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func serialFromCert(t *testing.T, pemData string) string {
	t.Helper()
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		t.Fatalf("failed to decode pem data")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse cert: %v", err)
	}
	return strings.ToLower(cert.SerialNumber.Text(16))
}
