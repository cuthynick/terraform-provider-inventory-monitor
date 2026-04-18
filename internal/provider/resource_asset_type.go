package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AssetTypeResource{}

type AssetTypeResource struct{ client *Client }

func NewAssetTypeResource() resource.Resource { return &AssetTypeResource{} }

func (r *AssetTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset_type"
}

func (r *AssetTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an inventory-monitor asset type.",
		Attributes: map[string]schema.Attribute{
			"id":          schemaID(),
			"name":        schemaRequiredString("Asset type name."),
			"slug":        schemaRequiredString("URL-friendly slug (unique)."),
			"description": schemaOptionalString("Description."),
			"color":       schemaOptionalString("Hex color code (without #), e.g. `aa1409`."),
		},
	}
}

func (r *AssetTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("expected *Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

type assetTypeModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
	Color       types.String `tfsdk:"color"`
}

type assetTypeAPI struct {
	ID          int64  `json:"id,omitempty"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
}

func (r *AssetTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data assetTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := assetTypeAPI{
		Name:        data.Name.ValueString(),
		Slug:        data.Slug.ValueString(),
		Description: data.Description.ValueString(),
		Color:       data.Color.ValueString(),
	}
	var result assetTypeAPI
	if err := r.client.Create("/asset-types/", body, &result); err != nil {
		resp.Diagnostics.AddError("Error creating asset type", err.Error())
		return
	}
	data.ID = types.Int64Value(result.ID)
	data.Name = types.StringValue(result.Name)
	data.Slug = types.StringValue(result.Slug)
	data.Description = types.StringValue(result.Description)
	data.Color = types.StringValue(result.Color)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AssetTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data assetTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result assetTypeAPI
	if err := r.client.Read("/asset-types/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", &result); err != nil {
		resp.Diagnostics.AddError("Error reading asset type", err.Error())
		return
	}
	data.Name = types.StringValue(result.Name)
	data.Slug = types.StringValue(result.Slug)
	data.Description = types.StringValue(result.Description)
	data.Color = types.StringValue(result.Color)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AssetTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data assetTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := assetTypeAPI{
		Name:        data.Name.ValueString(),
		Slug:        data.Slug.ValueString(),
		Description: data.Description.ValueString(),
		Color:       data.Color.ValueString(),
	}
	var result assetTypeAPI
	if err := r.client.Update("/asset-types/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", body, &result); err != nil {
		resp.Diagnostics.AddError("Error updating asset type", err.Error())
		return
	}
	data.Name = types.StringValue(result.Name)
	data.Slug = types.StringValue(result.Slug)
	data.Description = types.StringValue(result.Description)
	data.Color = types.StringValue(result.Color)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AssetTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data assetTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete("/asset-types/" + strconv.FormatInt(data.ID.ValueInt64(), 10) + "/"); err != nil {
		resp.Diagnostics.AddError("Error deleting asset type", err.Error())
	}
}

// schemaID is shared by all resources.
func schemaID() schema.Int64Attribute {
	return schema.Int64Attribute{
		Computed:            true,
		MarkdownDescription: "NetBox object ID.",
		PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
	}
}

func schemaRequiredString(desc string) schema.StringAttribute {
	return schema.StringAttribute{Required: true, MarkdownDescription: desc}
}

func schemaOptionalString(desc string) schema.StringAttribute {
	return schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: desc}
}

func schemaOptionalInt64(desc string) schema.Int64Attribute {
	return schema.Int64Attribute{Optional: true, Computed: true, MarkdownDescription: desc}
}

func (r *AssetTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric ID, got: "+req.ID)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
