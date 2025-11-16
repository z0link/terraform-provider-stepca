package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/z0link/terraform-provider-stepca/internal/client"
)

const (
	envCAURL            = "STEPCA_TEST_CA_URL"
	envToken            = "STEPCA_TEST_TOKEN"
	envAdminName        = "STEPCA_TEST_ADMIN_NAME"
	envAdminKey         = "STEPCA_TEST_ADMIN_KEY"
	envAdminProvisioner = "STEPCA_TEST_ADMIN_PROVISIONER"
	envAdminToken       = "STEPCA_TEST_ADMIN_TOKEN"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"stepca": providerserver.NewProtocol6WithError(New()),
}

type testAccSettings struct {
	CAURL            string
	Token            string
	AdminName        string
	AdminKey         string
	AdminProvisioner string
	AdminToken       string
}

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set to run acceptance tests")
	}
	required := []string{envCAURL, envToken, envAdminName, envAdminKey, envAdminProvisioner, envAdminToken}
	for _, key := range required {
		if os.Getenv(key) == "" {
			t.Fatalf("%s must be set for acceptance tests", key)
		}
	}
}

func testAccSettingsFromEnv() testAccSettings {
	return testAccSettings{
		CAURL:            os.Getenv(envCAURL),
		Token:            os.Getenv(envToken),
		AdminName:        os.Getenv(envAdminName),
		AdminKey:         os.Getenv(envAdminKey),
		AdminProvisioner: os.Getenv(envAdminProvisioner),
		AdminToken:       os.Getenv(envAdminToken),
	}
}

func testAccProviderConfig() string {
	settings := testAccSettingsFromEnv()
	return fmt.Sprintf(`
provider "stepca" {
  ca_url            = "%s"
  token             = "%s"
  admin_name        = "%s"
  admin_key         = "%s"
  admin_provisioner = "%s"
  admin_token       = "%s"
}
`, settings.CAURL, settings.Token, settings.AdminName, settings.AdminKey, settings.AdminProvisioner, settings.AdminToken)
}

func testAccClientFromEnv() *client.Client {
	settings := testAccSettingsFromEnv()
	c := client.New(settings.CAURL, settings.Token).
		WithAdminName(settings.AdminName).
		WithAdminKey(settings.AdminKey)
	if settings.AdminProvisioner != "" {
		c = c.WithAdminProvisioner(settings.AdminProvisioner)
	}
	if settings.AdminToken != "" {
		c = c.WithAdminToken(settings.AdminToken)
	}
	return c
}

func testAccGenerateCSR(t *testing.T, cn string) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}
	template := &x509.CertificateRequest{Subject: pkix.Name{CommonName: cn}}
	der, err := x509.CreateCertificateRequest(rand.Reader, template, key)
	if err != nil {
		t.Fatalf("failed to create csr: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der})
	return strings.TrimSpace(string(pemBytes))
}
