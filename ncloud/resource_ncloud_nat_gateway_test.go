package ncloud

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vpc"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccResourceNcloudNatGateway_basic(t *testing.T) {
	var natGateway vpc.NatGatewayInstance
	resourceName := "ncloud_nat_gateway.nat_gateway"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceNcloudNatGatewayConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNatGatewayExists(resourceName, &natGateway),
					resource.TestMatchResourceAttr(resourceName, "vpc_no", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(resourceName, "instance_no", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", "tf-testacc-nat-gateway"),
					resource.TestCheckResourceAttr(resourceName, "status", "RUN"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceNcloudNatGatewayConfig() string {
	return `
resource "ncloud_vpc" "vpc" {
	name            = "tf-testacc-nat-gateway"
	ipv4_cidr_block = "10.3.0.0/16"
}

resource "ncloud_nat_gateway" "nat_gateway" {
  vpc_no      = ncloud_vpc.vpc.vpc_no
  zone        = "KR-1"
  name        = "tf-testacc-nat-gateway"
  description = "description"
}
`
}

func testAccCheckNatGatewayExists(n string, natGateway *vpc.NatGatewayInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No nat gateway id is set")
		}

		client := testAccProvider.Meta().(*NcloudAPIClient)
		instance, err := getNatGatewayInstance(client, rs.Primary.ID)
		if err != nil {
			return err
		}

		*natGateway = *instance

		return nil
	}
}