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
	t      *testing.T
	serial string
	pem    string
	found  bool
	err    error
}

func (f *fakeCertificateClient) Sign(ctx context.Context, csr string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
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
