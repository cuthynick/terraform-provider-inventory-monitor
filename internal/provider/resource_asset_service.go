package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AssetServiceResource{}

type AssetServiceResource struct{ client *Client }

func NewAssetServiceResource() resource.Resource { return &AssetServiceResource{} }

func (r *AssetServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset_service"
}

func (r *AssetServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an inventory-monitor asset service record.",
		Attributes: map[string]schema.Attribute{
			"id":                       schemaID(),
			"asset_id":                 schema.Int64Attribute{Required: true, MarkdownDescription: "ID of the asset being serviced."},
			"contract_id":              schema.Int64Attribute{Required: true, MarkdownDescription: "ID of the service contract."},
			"service_start":            schemaOptionalString("Service start date (YYYY-MM-DD)."),
			"service_end":              schemaOptionalString("Service end date (YYYY-MM-DD)."),
			"service_price":            schemaOptionalString("Service price (decimal string)."),
			"service_currency":         schemaOptionalString("Currency code."),
			"service_category":         schemaOptionalString("Service category."),
			"service_category_vendor":  schemaOptionalString("Vendor service category name."),
			"description":              schemaOptionalString("Description."),
			"comments":                 schemaOptionalString("Comments."),
		},
	}
}

func (r *AssetServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type assetServiceModel struct {
	ID                    types.Int64  `tfsdk:"id"`
	AssetID               types.Int64  `tfsdk:"asset_id"`
	ContractID            types.Int64  `tfsdk:"contract_id"`
	ServiceStart          types.String `tfsdk:"service_start"`
	ServiceEnd            types.String `tfsdk:"service_end"`
	ServicePrice          types.String `tfsdk:"service_price"`
	ServiceCurrency       types.String `tfsdk:"service_currency"`
	ServiceCategory       types.String `tfsdk:"service_category"`
	ServiceCategoryVendor types.String `tfsdk:"service_category_vendor"`
	Description           types.String `tfsdk:"description"`
	Comments              types.String `tfsdk:"comments"`
}

type assetServiceAPIWrite struct {
	Asset                 nestedID `json:"asset"`
	Contract              nestedID `json:"contract"`
	ServiceStart          string   `json:"service_start,omitempty"`
	ServiceEnd            string   `json:"service_end,omitempty"`
	ServicePrice          string   `json:"service_price,omitempty"`
	ServiceCurrency       string   `json:"service_currency,omitempty"`
	ServiceCategory       string   `json:"service_category,omitempty"`
	ServiceCategoryVendor string   `json:"service_category_vendor,omitempty"`
	Description           string   `json:"description,omitempty"`
	Comments              string   `json:"comments,omitempty"`
}

type assetServiceAPIRead struct {
	ID                    int64    `json:"id"`
	Asset                 nestedID `json:"asset"`
	Contract              nestedID `json:"contract"`
	ServiceStart          string   `json:"service_start"`
	ServiceEnd            string   `json:"service_end"`
	ServicePrice          string   `json:"service_price"`
	ServiceCurrency       string   `json:"service_currency"`
	ServiceCategory       string   `json:"service_category"`
	ServiceCategoryVendor string   `json:"service_category_vendor"`
	Description           string   `json:"description"`
	Comments              string   `json:"comments"`
}

func assetServiceModelToAPI(data assetServiceModel) assetServiceAPIWrite {
	return assetServiceAPIWrite{
		Asset:                 nestedID{ID: data.AssetID.ValueInt64()},
		Contract:              nestedID{ID: data.ContractID.ValueInt64()},
		ServiceStart:          data.ServiceStart.ValueString(),
		ServiceEnd:            data.ServiceEnd.ValueString(),
		ServicePrice:          data.ServicePrice.ValueString(),
		ServiceCurrency:       data.ServiceCurrency.ValueString(),
		ServiceCategory:       data.ServiceCategory.ValueString(),
		ServiceCategoryVendor: data.ServiceCategoryVendor.ValueString(),
		Description:           data.Description.ValueString(),
		Comments:              data.Comments.ValueString(),
	}
}

func assetServiceAPIToModel(result assetServiceAPIRead, data *assetServiceModel) {
	data.AssetID = types.Int64Value(result.Asset.ID)
	data.ContractID = types.Int64Value(result.Contract.ID)
	data.ServiceStart = types.StringValue(result.ServiceStart)
	data.ServiceEnd = types.StringValue(result.ServiceEnd)
	data.ServicePrice = types.StringValue(result.ServicePrice)
	data.ServiceCurrency = types.StringValue(result.ServiceCurrency)
	data.ServiceCategory = types.StringValue(result.ServiceCategory)
	data.ServiceCategoryVendor = types.StringValue(result.ServiceCategoryVendor)
	data.Description = types.StringValue(result.Description)
	data.Comments = types.StringValue(result.Comments)
}

func (r *AssetServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data assetServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result assetServiceAPIRead
	if err := r.client.Create("/asset-services/", assetServiceModelToAPI(data), &result); err != nil {
		resp.Diagnostics.AddError("Error creating asset service", err.Error())
		return
	}
	data.ID = types.Int64Value(result.ID)
	assetServiceAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AssetServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data assetServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result assetServiceAPIRead
	if err := r.client.Read("/asset-services/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", &result); err != nil {
		resp.Diagnostics.AddError("Error reading asset service", err.Error())
		return
	}
	assetServiceAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AssetServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data assetServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result assetServiceAPIRead
	if err := r.client.Update("/asset-services/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", assetServiceModelToAPI(data), &result); err != nil {
		resp.Diagnostics.AddError("Error updating asset service", err.Error())
		return
	}
	assetServiceAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AssetServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data assetServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete("/asset-services/" + strconv.FormatInt(data.ID.ValueInt64(), 10) + "/"); err != nil {
		resp.Diagnostics.AddError("Error deleting asset service", err.Error())
	}
}

func (r *AssetServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric ID, got: "+req.ID)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
