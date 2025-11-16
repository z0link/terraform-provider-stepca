package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccProvisionerResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}
	name := fmt.Sprintf("acc-prov-%s", acctest.RandString(6))
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckProvisionerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionerConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stepca_provisioner.test", "name", name),
					resource.TestCheckResourceAttr("stepca_provisioner.test", "type", "JWK"),
					resource.TestCheckResourceAttr("stepca_provisioner.test", "admin", "true"),
				),
			},
		},
	})
}

func testAccProvisionerConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "stepca_provisioner" "test" {
  name  = "%s"
  type  = "JWK"
  admin = true
}
`, testAccProviderConfig(), name)
}

func testAccCheckProvisionerDestroy(state *terraform.State) error {
	c := testAccClientFromEnv()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "stepca_provisioner" {
			continue
		}
		name := rs.Primary.Attributes["name"]
		if name == "" {
			continue
		}
		p, err := c.GetProvisioner(context.Background(), name)
		if err != nil {
			return err
		}
		if p != nil {
			return fmt.Errorf("provisioner %s still exists", name)
		}
	}
	return nil
}
