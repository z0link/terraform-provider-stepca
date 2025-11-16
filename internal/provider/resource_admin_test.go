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

func TestAccAdminResource_basic(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}
	provisionerName := fmt.Sprintf("acc-admin-prov-%s", acctest.RandString(6))
	adminName := fmt.Sprintf("acc-admin-%s", acctest.RandString(6))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckAdminDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAdminConfig(provisionerName, adminName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("stepca_provisioner.test", "name", provisionerName),
					resource.TestCheckResourceAttr("stepca_admin.test", "name", adminName),
					resource.TestCheckResourceAttr("stepca_admin.test", "provisioner_name", provisionerName),
				),
			},
		},
	})
}

func testAccAdminConfig(provisionerName, adminName string) string {
	return fmt.Sprintf(`
%s

resource "stepca_provisioner" "test" {
  name  = "%s"
  type  = "JWK"
  admin = true
}

resource "stepca_admin" "test" {
  name             = "%s"
  provisioner_name = stepca_provisioner.test.name
}
`, testAccProviderConfig(), provisionerName, adminName)
}

func testAccCheckAdminDestroy(state *terraform.State) error {
	c := testAccClientFromEnv()
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "stepca_admin" {
			continue
		}
		name := rs.Primary.Attributes["name"]
		provisioner := rs.Primary.Attributes["provisioner_name"]
		if name == "" || provisioner == "" {
			continue
		}
		admin, err := c.GetAdmin(context.Background(), name, provisioner)
		if err != nil {
			return err
		}
		if admin != nil {
			return fmt.Errorf("admin %s/%s still exists", provisioner, name)
		}
	}
	return nil
}
