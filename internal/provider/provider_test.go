package provider_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/nickcuthbert/terraform-provider-inventory-monitor/internal/provider"
)

func randSuffix() string {
	return fmt.Sprintf("%06d", rand.Intn(999999))
}

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"inventorymonitor": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testTenantIDs returns the 6 tenant IDs from NETBOX_TEST_TENANT_IDS (comma-separated).
func testTenantIDs(t *testing.T) [6]int64 {
	t.Helper()
	raw := os.Getenv("NETBOX_TEST_TENANT_IDS")
	parts := strings.Split(raw, ",")
	if len(parts) < 6 {
		t.Fatalf("NETBOX_TEST_TENANT_IDS must contain at least 6 comma-separated tenant IDs, got %q", raw)
	}
	var ids [6]int64
	for i := 0; i < 6; i++ {
		id, err := strconv.ParseInt(strings.TrimSpace(parts[i]), 10, 64)
		if err != nil {
			t.Fatalf("NETBOX_TEST_TENANT_IDS[%d] is not a valid integer: %q", i, parts[i])
		}
		ids[i] = id
	}
	return ids
}

func testEnvInt64(t *testing.T, key string) int64 {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Fatalf("%s must be set for acceptance tests", key)
	}
	id, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil {
		t.Fatalf("%s must be a valid integer, got %q", key, v)
	}
	return id
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("NETBOX_SERVER_URL") == "" {
		t.Fatal("NETBOX_SERVER_URL must be set for acceptance tests")
	}
	if os.Getenv("NETBOX_API_TOKEN") == "" {
		t.Fatal("NETBOX_API_TOKEN must be set for acceptance tests")
	}
	if os.Getenv("NETBOX_TEST_TENANT_IDS") == "" {
		t.Fatal("NETBOX_TEST_TENANT_IDS must be set for acceptance tests (6 comma-separated NetBox tenant IDs)")
	}
	if os.Getenv("NETBOX_TEST_DEVICE_ID") == "" {
		t.Fatal("NETBOX_TEST_DEVICE_ID must be set for acceptance tests")
	}
	if os.Getenv("NETBOX_TEST_SITE_ID") == "" {
		t.Fatal("NETBOX_TEST_SITE_ID must be set for acceptance tests")
	}
	if os.Getenv("NETBOX_TEST_LOCATION_ID") == "" {
		t.Fatal("NETBOX_TEST_LOCATION_ID must be set for acceptance tests")
	}
}

func providerConfig() string {
	return `
provider "inventorymonitor" {
  server_url = "` + os.Getenv("NETBOX_SERVER_URL") + `"
  api_token  = "` + os.Getenv("NETBOX_API_TOKEN") + `"
}
`
}

func TestAccAssetType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + `
resource "inventorymonitor_asset_type" "test" {
  name        = "Test Switch Type"
  slug        = "test-switch-type"
  description = "Created by acceptance test"
  color       = "2196f3"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("inventorymonitor_asset_type.test", "name", "Test Switch Type"),
					resource.TestCheckResourceAttr("inventorymonitor_asset_type.test", "slug", "test-switch-type"),
					resource.TestCheckResourceAttrSet("inventorymonitor_asset_type.test", "id"),
				),
			},
			{
				Config: providerConfig() + `
resource "inventorymonitor_asset_type" "test" {
  name        = "Test Switch Type Updated"
  slug        = "test-switch-type"
  description = "Updated by acceptance test"
  color       = "2196f3"
}`,
				Check: resource.TestCheckResourceAttr("inventorymonitor_asset_type.test", "name", "Test Switch Type Updated"),
			},
		},
	})
}

func TestAccContractor(t *testing.T) {
	tenants := testTenantIDs(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "inventorymonitor_contractor" "test" {
  name      = "Test Vendor"
  tenant_id = %d
  company   = "Test Vendor Inc."
  address   = "123 Test St"
}`, tenants[0]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("inventorymonitor_contractor.test", "name", "Test Vendor"),
					resource.TestCheckResourceAttr("inventorymonitor_contractor.test", "company", "Test Vendor Inc."),
					resource.TestCheckResourceAttrSet("inventorymonitor_contractor.test", "id"),
				),
			},
		},
	})
}

func TestAccContract(t *testing.T) {
	tenants := testTenantIDs(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "inventorymonitor_contractor" "test" {
  name      = "Contract Test Vendor"
  tenant_id = %d
}
resource "inventorymonitor_contract" "test" {
  name           = "Test Supply Contract"
  name_internal  = "TSC-001"
  contractor_id  = inventorymonitor_contractor.test.id
  type           = "supply"
  description    = "Created by acceptance test"
}`, tenants[1]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("inventorymonitor_contract.test", "name", "Test Supply Contract"),
					resource.TestCheckResourceAttr("inventorymonitor_contract.test", "type", "supply"),
					resource.TestCheckResourceAttrSet("inventorymonitor_contract.test", "id"),
				),
			},
		},
	})
}

func TestAccAsset(t *testing.T) {
	sfx := randSuffix()
	tenants := testTenantIDs(t)
	deviceID := testEnvInt64(t, "NETBOX_TEST_DEVICE_ID")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "inventorymonitor_asset_type" "test" {
  name = "Asset Test Type %s"
  slug = "asset-test-type-%s"
}
resource "inventorymonitor_contractor" "test" {
  name      = "Asset Test Vendor %s"
  tenant_id = %d
}
resource "inventorymonitor_contract" "test" {
  name          = "Asset Test Contract %s"
  name_internal = "ATC-%s"
  contractor_id = inventorymonitor_contractor.test.id
  type          = "order"
}
resource "inventorymonitor_asset" "test" {
  serial            = "ACC-TEST-%s"
  type_id           = inventorymonitor_asset_type.test.id
  order_contract_id = inventorymonitor_contract.test.id
  assignment_status = "deployed"
  lifecycle_status  = "in_use"
  description       = "Acceptance test asset"
  assigned_object_type = "dcim.device"
  assigned_object_id   = %d
}`, sfx, sfx, sfx, tenants[2], sfx, sfx, sfx, deviceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("inventorymonitor_asset.test", "id"),
					resource.TestCheckResourceAttr("inventorymonitor_asset.test", "assignment_status", "deployed"),
				),
			},
		},
	})
}

func TestAccInvoice(t *testing.T) {
	tenants := testTenantIDs(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "inventorymonitor_contractor" "test" {
  name      = "Invoice Test Vendor"
  tenant_id = %d
}
resource "inventorymonitor_contract" "test" {
  name          = "Invoice Test Contract"
  name_internal = "ITC-001"
  contractor_id = inventorymonitor_contractor.test.id
  type          = "supply"
}
resource "inventorymonitor_invoice" "test" {
  name          = "INV-2024-001"
  name_internal = "INTERNAL-INV-001"
  contract_id   = inventorymonitor_contract.test.id
  description   = "Acceptance test invoice"
}`, tenants[3]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("inventorymonitor_invoice.test", "name", "INV-2024-001"),
					resource.TestCheckResourceAttrSet("inventorymonitor_invoice.test", "id"),
				),
			},
		},
	})
}

func TestAccRMA(t *testing.T) {
	sfx := randSuffix()
	tenants := testTenantIDs(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "inventorymonitor_asset_type" "test" {
  name = "RMA Test Type %s"
  slug = "rma-test-type-%s"
}
resource "inventorymonitor_contractor" "test" {
  name      = "RMA Test Vendor %s"
  tenant_id = %d
}
resource "inventorymonitor_contract" "test" {
  name          = "RMA Test Contract %s"
  name_internal = "RTC-%s"
  contractor_id = inventorymonitor_contractor.test.id
  type          = "order"
}
resource "inventorymonitor_asset" "test" {
  serial            = "RMA-TEST-%s"
  type_id           = inventorymonitor_asset_type.test.id
  order_contract_id = inventorymonitor_contract.test.id
}
resource "inventorymonitor_rma" "test" {
  asset_id          = inventorymonitor_asset.test.id
  issue_description = "Unit not powering on"
  rma_number        = "RMA-%s"
  status            = "pending"
}`, sfx, sfx, sfx, tenants[4], sfx, sfx, sfx, sfx),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("inventorymonitor_rma.test", "issue_description", "Unit not powering on"),
				resource.TestCheckResourceAttr("inventorymonitor_rma.test", "status", "pending"),
				resource.TestCheckResourceAttrSet("inventorymonitor_rma.test", "id"),
			),
		},
		},
	})
}

func TestAccProbe(t *testing.T) {
	deviceID := testEnvInt64(t, "NETBOX_TEST_DEVICE_ID")
	siteID := testEnvInt64(t, "NETBOX_TEST_SITE_ID")
	locationID := testEnvInt64(t, "NETBOX_TEST_LOCATION_ID")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "inventorymonitor_probe" "test" {
  name        = "test-probe-001"
  time        = "2099-12-31T23:59:59Z"
  serial      = "PROBE-TEST-001"
  device_id   = %d
  site_id     = %d
  location_id = %d
  description = "Acceptance test probe"
}`, deviceID, siteID, locationID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("inventorymonitor_probe.test", "serial", "PROBE-TEST-001"),
					resource.TestCheckResourceAttrSet("inventorymonitor_probe.test", "id"),
				),
			},
		},
	})
}

func TestAccAssetService(t *testing.T) {
	sfx := randSuffix()
	tenants := testTenantIDs(t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig() + fmt.Sprintf(`
resource "inventorymonitor_asset_type" "test" {
  name = "Service Test Type %s"
  slug = "service-test-type-%s"
}
resource "inventorymonitor_contractor" "test" {
  name      = "Service Test Vendor %s"
  tenant_id = %d
}
resource "inventorymonitor_contract" "test" {
  name          = "Service Test Contract %s"
  name_internal = "STC-%s"
  contractor_id = inventorymonitor_contractor.test.id
  type          = "service"
}
resource "inventorymonitor_asset" "test" {
  serial            = "SVC-TEST-%s"
  type_id           = inventorymonitor_asset_type.test.id
  order_contract_id = inventorymonitor_contract.test.id
}
resource "inventorymonitor_asset_service" "test" {
  asset_id         = inventorymonitor_asset.test.id
  contract_id      = inventorymonitor_contract.test.id
  service_start    = "2024-01-01"
  service_end      = "2025-01-01"
  service_category = "Hardware Maintenance"
  description      = "Annual hardware support"
}`, sfx, sfx, sfx, tenants[5], sfx, sfx, sfx),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("inventorymonitor_asset_service.test", "service_category", "Hardware Maintenance"),
					resource.TestCheckResourceAttrSet("inventorymonitor_asset_service.test", "id"),
				),
			},
		},
	})
}
