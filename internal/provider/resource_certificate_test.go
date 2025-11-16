package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCertificateResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}
	csr := testAccGenerateCSR(t, "acc-cert")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig(csr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("stepca_certificate.test", "certificate"),
					resource.TestMatchResourceAttr("stepca_certificate.test", "certificate", regexp.MustCompile("BEGIN CERTIFICATE")),
				),
			},
		},
	})
}

func testAccCertificateConfig(csr string) string {
	return fmt.Sprintf(`
%s

resource "stepca_certificate" "test" {
  csr = <<-EOT
%s
EOT
}
`, testAccProviderConfig(), csr)
}
