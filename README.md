# terraform-provider-inventory-monitor

A Terraform provider for the [CESNET inventory-monitor NetBox plugin](https://github.com/CESNET/inventory-monitor-plugin).

Manage assets, contracts, invoices, RMAs, and discovery probes as Terraform resources, fully integrated with your existing NetBox infrastructure.

## Resources

| Resource | Description |
|---|---|
| `inventorymonitor_asset_type` | Asset type / category (e.g. Network Switch) |
| `inventorymonitor_contractor` | Vendor or contractor linked to a NetBox tenant |
| `inventorymonitor_contract` | Supply, service, or order contract |
| `inventorymonitor_asset` | Physical asset with serial number, lifecycle status, and NetBox object assignment |
| `inventorymonitor_asset_service` | Service contract line item for an asset |
| `inventorymonitor_invoice` | Invoice linked to a contract |
| `inventorymonitor_rma` | Return Merchandise Authorization |
| `inventorymonitor_probe` | Discovery snapshot linking a serial to a device/site/location at a point in time |

All resources support import via `terraform import <resource>.<name> <id>`.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build from source)
- NetBox instance with the [inventory-monitor plugin](https://github.com/CESNET/inventory-monitor-plugin) installed

## Usage

```hcl
terraform {
  required_providers {
    inventorymonitor = {
      source  = "your-namespace/inventory-monitor"
      version = "~> 0.1"
    }
  }
}

provider "inventorymonitor" {
  server_url = "http://your-netbox-host:8000"
  api_token  = var.netbox_api_token
}
```

Provider configuration can also be supplied via environment variables:

```bash
export NETBOX_SERVER_URL="http://your-netbox-host:8000"
export NETBOX_API_TOKEN="your-api-token"
```

### Example

```hcl
resource "inventorymonitor_asset_type" "network_switch" {
  name        = "Network Switch"
  slug        = "network-switch"
  description = "Managed network switch"
  color       = "2196f3"
}

resource "inventorymonitor_contractor" "acme" {
  name      = "Acme Corp"
  tenant_id = 1   # NetBox tenant ID
  company   = "Acme Corporation"
}

resource "inventorymonitor_contract" "hardware_supply" {
  name          = "Acme Hardware Supply"
  name_internal = "ACME-HW-001"
  contractor_id = inventorymonitor_contractor.acme.id
  type          = "supply"
}

resource "inventorymonitor_asset" "switch_01" {
  serial            = "S00001"
  type_id           = inventorymonitor_asset_type.network_switch.id
  order_contract_id = inventorymonitor_contract.hardware_supply.id
  assignment_status = "deployed"
  lifecycle_status  = "in_use"
  assigned_object_type = "dcim.device"
  assigned_object_id   = 42   # NetBox device ID
}
```

## Building from Source

```bash
git clone https://github.com/cuthynick/terraform-provider-inventory-monitor.git
cd terraform-provider-inventory-monitor
go build ./...
```

To install locally using a [filesystem mirror](https://developer.hashicorp.com/terraform/cli/config/config-file#filesystem_mirror):

```bash
PLUGIN_DIR=~/.terraform.d/plugins/registry.terraform.io/your-namespace/inventory-monitor/0.1.0/$(go env GOOS)_$(go env GOARCH)
mkdir -p "$PLUGIN_DIR"
go build -o "$PLUGIN_DIR/terraform-provider-inventory-monitor_v0.1.0" .
```

Add to `~/.terraformrc`:

```hcl
provider_installation {
  filesystem_mirror {
    path    = "/home/youruser/.terraform.d/plugins"
    include = ["your-namespace/inventory-monitor"]
  }
  direct {
    exclude = ["your-namespace/inventory-monitor"]
  }
}
```

## Running Acceptance Tests

Acceptance tests require a running NetBox instance with the inventory-monitor plugin installed and at least 6 pre-existing tenants.

```bash
export TF_ACC=1
export NETBOX_SERVER_URL="http://localhost:8000"
export NETBOX_API_TOKEN="your-api-token"

# Six unique NetBox tenant IDs (one per contractor test, to satisfy the
# plugin's one-contractor-per-tenant constraint)
export NETBOX_TEST_TENANT_IDS="1,2,3,4,5,6"

# A NetBox device, site, and location that exist in your instance
# (used by the asset and probe acceptance tests)
export NETBOX_TEST_DEVICE_ID="42"
export NETBOX_TEST_SITE_ID="1"
export NETBOX_TEST_LOCATION_ID="1"

go test ./internal/provider/ -v -run TestAcc
```

## Known Plugin Quirks

Two serializer bugs exist in the upstream inventory-monitor plugin that affect this provider:

1. **`contract.parent` required error** — The serializer marks `parent` as required despite the model allowing null. If you hit `{"parent": ["This field is required."]}`, apply this patch to your NetBox instance's `inventory_monitor/api/serializers.py`:
   ```python
   # line ~131, inside ContractSerializer.build_relational_field or __init__
   fields["parent"] = ContractSerializer(nested=True, required=False, allow_null=True)
   ```

2. **`asset.order_contract` deletion error** — The plugin raises a `ProtectedError` when deleting an asset that has `order_contract` set. This provider works around this automatically by nulling the field via PATCH before issuing DELETE.

## Contributing

Contributions are welcome. Please open an issue or pull request on GitHub.

When adding a new resource, follow the existing pattern in `internal/provider/resource_*.go`: separate `*APIWrite` / `*APIRead` structs, `modelToAPI` / `APIToModel` helpers, and implement `ImportState` with integer ID parsing.

## License

MIT
