package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &InventoryMonitorProvider{}
var _ provider.ProviderWithFunctions = &InventoryMonitorProvider{}

type InventoryMonitorProvider struct {
	version string
}

type InventoryMonitorProviderModel struct {
	ServerURL types.String `tfsdk:"server_url"`
	APIToken  types.String `tfsdk:"api_token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &InventoryMonitorProvider{version: version}
	}
}

func (p *InventoryMonitorProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "inventorymonitor"
	resp.Version = p.version
}

func (p *InventoryMonitorProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provider for managing resources in the NetBox inventory-monitor plugin (https://github.com/CESNET/inventory-monitor-plugin).",
		Attributes: map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				MarkdownDescription: "NetBox server URL (e.g. `http://localhost:8000`). Can also be set via `NETBOX_SERVER_URL` env var.",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "NetBox API token. Can also be set via `NETBOX_API_TOKEN` env var.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *InventoryMonitorProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data InventoryMonitorProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverURL := os.Getenv("NETBOX_SERVER_URL")
	if !data.ServerURL.IsNull() {
		serverURL = data.ServerURL.ValueString()
	}
	if serverURL == "" {
		resp.Diagnostics.AddError("Missing server_url", "Set server_url in provider config or NETBOX_SERVER_URL env var.")
		return
	}

	apiToken := os.Getenv("NETBOX_API_TOKEN")
	if !data.APIToken.IsNull() {
		apiToken = data.APIToken.ValueString()
	}
	if apiToken == "" {
		resp.Diagnostics.AddError("Missing api_token", "Set api_token in provider config or NETBOX_API_TOKEN env var.")
		return
	}

	client := NewClient(serverURL, apiToken)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *InventoryMonitorProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAssetTypeResource,
		NewContractorResource,
		NewContractResource,
		NewAssetResource,
		NewAssetServiceResource,
		NewInvoiceResource,
		NewRMAResource,
		NewProbeResource,
	}
}

func (p *InventoryMonitorProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *InventoryMonitorProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}
