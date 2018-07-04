package ncloud

import (
	"fmt"
	"github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestAccResourceNcloudServerBasic(t *testing.T) {
	var serverInstance sdk.ServerInstance
	testServerName := getTestServerName()

	testCheck := func() func(*terraform.State) error {
		return func(*terraform.State) error {
			if serverInstance.ServerName != testServerName {
				return fmt.Errorf("not found: %s", testServerName)
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "ncloud_server.server",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig(testServerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(
						"ncloud_server.server", &serverInstance),
					testCheck(),
					resource.TestCheckResourceAttr(
						"ncloud_server.server",
						"server_image_product_code",
						"SPSW0LINUX000032"),
					resource.TestCheckResourceAttr(
						"ncloud_server.server",
						"server_product_code",
						"SPSVRSTAND000004"),
				),
			},
		},
	})
}

func TestAccResourceInstanceChangeServerInstanceSpec(t *testing.T) {
	var before sdk.ServerInstance
	var after sdk.ServerInstance
	testServerName := getTestServerName()

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "ncloud_server.server",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServerConfig(testServerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(
						"ncloud_server.server", &before),
				),
			},
			{
				Config: testAccInstanceChangeSpecConfig(testServerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(
						"ncloud_server.server", &after),
					testAccCheckInstanceNotRecreated(
						t, &before, &after),
				),
			},
		},
	})
}

// ignore test: must need real test data
func testAccResourceRecreateServerInstance(t *testing.T) {
	var before sdk.ServerInstance
	var after sdk.ServerInstance
	testServerName := getTestServerName()

	resource.Test(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "ncloud_server.server",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRecreateServerInstanceBeforeConfig(testServerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(
						"ncloud_server.server", &before),
				),
			},
			{
				Config: testAccRecreateServerInstanceAfterConfig(testServerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerExists(
						"ncloud_server.server", &after),
					testAccCheckInstanceNotRecreated(
						t, &before, &after),
					resource.TestCheckResourceAttr(
						"ncloud_server.server",
						"server_image_product_code",
						"SPSWBMLINUX00002"),
				),
			},
		},
	})
}

func testAccCheckServerExists(n string, i *sdk.ServerInstance) resource.TestCheckFunc {
	return testAccCheckInstanceExistsWithProvider(n, i, func() *schema.Provider { return testAccProvider })
}

func testAccCheckInstanceExistsWithProvider(n string, i *sdk.ServerInstance, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		provider := providerF()
		conn := provider.Meta().(*NcloudSdk).conn
		instance, err := getServerInstance(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if instance != nil {
			*i = *instance
			return nil
		}

		return fmt.Errorf("server instance not found")
	}
}

func testAccCheckInstanceNotRecreated(t *testing.T,
	before, after *sdk.ServerInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.ServerInstanceNo != after.ServerInstanceNo {
			t.Fatalf("Ncloud Instance IDs have changed. Before %s. After %s", before.ServerInstanceNo, after.ServerInstanceNo)
		}
		return nil
	}
}

func testAccCheckServerDestroy(s *terraform.State) error {
	return testAccCheckInstanceDestroyWithProvider(s, testAccProvider)
}

func testAccCheckInstanceDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*NcloudSdk).conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "ncloud_server" {
			continue
		}
		instance, err := getServerInstance(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if instance == nil {
			continue
		}

		if instance.ServerInstanceStatusName != "terminating" {
			return fmt.Errorf("found unterminated instance: %s", instance.ServerInstanceNo)
		}
	}

	return nil
}

func getTestServerName() string {
	rInt := acctest.RandIntRange(1, 9999)
	testServerName := fmt.Sprintf("tf-%d-vm", rInt)
	return testServerName
}

func testAccServerConfig(testServerName string) string {
	return fmt.Sprintf(`
resource "ncloud_server" "server" {
	"server_name" = "%s"
	"server_image_product_code" = "SPSW0LINUX000032"
	"server_product_code" = "SPSVRSTAND000004"
}
`, testServerName)
}

func testAccInstanceChangeSpecConfig(testServerName string) string {
	return fmt.Sprintf(`
resource "ncloud_server" "server" {
	"server_name" = "%s"
	"server_image_product_code" = "SPSW0LINUX000032"
	"server_product_code" = "SPSVRSTAND000056"
}
`, testServerName)
}

func testAccRecreateServerInstanceBeforeConfig(testServerName string) string {
	return fmt.Sprintf(`
resource "ncloud_server" "server" {
	"server_name" = "%s"
	"server_image_product_code" = "SPSWBMLINUX00001"
	"server_product_code" = "SPSVRBM000000001"
}
`, testServerName)
}

func testAccRecreateServerInstanceAfterConfig(testServerName string) string {
	return fmt.Sprintf(`
resource "ncloud_server" "server" {
	"server_name" = "%s"
	"server_image_product_code" = "SPSWBMLINUX00002"
	"server_product_code" = "SPSVRBM000000001"
}
`, testServerName)
}