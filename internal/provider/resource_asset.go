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

var _ resource.Resource = &AssetResource{}

type AssetResource struct{ client *Client }

func NewAssetResource() resource.Resource { return &AssetResource{} }

func (r *AssetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_asset"
}

func (r *AssetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an inventory-monitor asset.",
		Attributes: map[string]schema.Attribute{
			"id":                   schemaID(),
			"serial":               schemaRequiredString("Hardware serial number."),
			"type_id":              schema.Int64Attribute{Required: true, MarkdownDescription: "ID of the asset type."},
			"order_contract_id":    schema.Int64Attribute{Required: true, MarkdownDescription: "ID of the order/supply contract."},
			"partnumber":           schemaOptionalString("Part number."),
			"description":          schemaOptionalString("Description."),
			"assignment_status":    schemaOptionalString("Assignment status: `reserved`, `deployed`, `loaned`, or `stocked`."),
			"lifecycle_status":     schemaOptionalString("Lifecycle status: `new`, `in_stock`, `in_use`, `in_maintenance`, `retired`, or `disposed`."),
			"assigned_object_type": schemaOptionalString("Content type of the assigned object, e.g. `dcim.device`."),
			"assigned_object_id":   schemaOptionalInt64("ID of the assigned object."),
			"project":              schemaOptionalString("Project name."),
			"vendor":               schemaOptionalString("Vendor name."),
			"quantity":             schemaOptionalInt64("Quantity."),
			"price":                schemaOptionalString("Unit price (decimal string)."),
			"currency":             schemaOptionalString("Currency code."),
			"warranty_start":       schemaOptionalString("Warranty start date (YYYY-MM-DD)."),
			"warranty_end":         schemaOptionalString("Warranty end date (YYYY-MM-DD)."),
			"comments":             schemaOptionalString("Comments."),
		},
	}
}

func (r *AssetResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type assetModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	Serial             types.String `tfsdk:"serial"`
	TypeID             types.Int64  `tfsdk:"type_id"`
	OrderContractID    types.Int64  `tfsdk:"order_contract_id"`
	Partnumber         types.String `tfsdk:"partnumber"`
	Description        types.String `tfsdk:"description"`
	AssignmentStatus   types.String `tfsdk:"assignment_status"`
	LifecycleStatus    types.String `tfsdk:"lifecycle_status"`
	AssignedObjectType types.String `tfsdk:"assigned_object_type"`
	AssignedObjectID   types.Int64  `tfsdk:"assigned_object_id"`
	Project            types.String `tfsdk:"project"`
	Vendor             types.String `tfsdk:"vendor"`
	Quantity           types.Int64  `tfsdk:"quantity"`
	Price              types.String `tfsdk:"price"`
	Currency           types.String `tfsdk:"currency"`
	WarrantyStart      types.String `tfsdk:"warranty_start"`
	WarrantyEnd        types.String `tfsdk:"warranty_end"`
	Comments           types.String `tfsdk:"comments"`
}

type assetAPIWrite struct {
	Serial             string   `json:"serial"`
	Type               nestedID `json:"type"`
	OrderContract      nestedID `json:"order_contract"`
	Partnumber         string   `json:"partnumber,omitempty"`
	Description        string   `json:"description,omitempty"`
	AssignmentStatus   string   `json:"assignment_status,omitempty"`
	LifecycleStatus    string   `json:"lifecycle_status,omitempty"`
	AssignedObjectType string   `json:"assigned_object_type,omitempty"`
	AssignedObjectID   *int64   `json:"assigned_object_id,omitempty"`
	Project            string   `json:"project,omitempty"`
	Vendor             string   `json:"vendor,omitempty"`
	Quantity           *int64   `json:"quantity,omitempty"`
	Price              string   `json:"price,omitempty"`
	Currency           string   `json:"currency,omitempty"`
	WarrantyStart      string   `json:"warranty_start,omitempty"`
	WarrantyEnd        string   `json:"warranty_end,omitempty"`
	Comments           string   `json:"comments,omitempty"`
}

type assetAPIRead struct {
	ID                 int64    `json:"id"`
	Serial             string   `json:"serial"`
	Type               nestedID `json:"type"`
	OrderContract      nestedID `json:"order_contract"`
	Partnumber         string   `json:"partnumber"`
	Description        string   `json:"description"`
	AssignmentStatus   string   `json:"assignment_status"`
	LifecycleStatus    string   `json:"lifecycle_status"`
	AssignedObjectType string   `json:"assigned_object_type"`
	AssignedObjectID   *int64   `json:"assigned_object_id"`
	Project            string   `json:"project"`
	Vendor             string   `json:"vendor"`
	Quantity           *int64   `json:"quantity"`
	Price              string   `json:"price"`
	Currency           string   `json:"currency"`
	WarrantyStart      string   `json:"warranty_start"`
	WarrantyEnd        string   `json:"warranty_end"`
	Comments           string   `json:"comments"`
}

func assetModelToAPI(data assetModel) assetAPIWrite {
	body := assetAPIWrite{
		Serial:           data.Serial.ValueString(),
		Type:             nestedID{ID: data.TypeID.ValueInt64()},
		OrderContract:    nestedID{ID: data.OrderContractID.ValueInt64()},
		Partnumber:       data.Partnumber.ValueString(),
		Description:      data.Description.ValueString(),
		AssignmentStatus: data.AssignmentStatus.ValueString(),
		LifecycleStatus:  data.LifecycleStatus.ValueString(),
		Project:          data.Project.ValueString(),
		Vendor:           data.Vendor.ValueString(),
		Price:            data.Price.ValueString(),
		Currency:         data.Currency.ValueString(),
		WarrantyStart:    data.WarrantyStart.ValueString(),
		WarrantyEnd:      data.WarrantyEnd.ValueString(),
		Comments:         data.Comments.ValueString(),
	}
	if !data.AssignedObjectType.IsNull() && !data.AssignedObjectType.IsUnknown() {
		body.AssignedObjectType = data.AssignedObjectType.ValueString()
	}
	if !data.AssignedObjectID.IsNull() && !data.AssignedObjectID.IsUnknown() {
		v := data.AssignedObjectID.ValueInt64()
		body.AssignedObjectID = &v
	}
	if !data.Quantity.IsNull() && !data.Quantity.IsUnknown() {
		v := data.Quantity.ValueInt64()
		body.Quantity = &v
	}
	return body
}

func assetAPIToModel(result assetAPIRead, data *assetModel) {
	data.Serial = types.StringValue(result.Serial)
	data.TypeID = types.Int64Value(result.Type.ID)
	data.OrderContractID = types.Int64Value(result.OrderContract.ID)
	data.Partnumber = types.StringValue(result.Partnumber)
	data.Description = types.StringValue(result.Description)
	data.AssignmentStatus = types.StringValue(result.AssignmentStatus)
	data.LifecycleStatus = types.StringValue(result.LifecycleStatus)
	data.AssignedObjectType = types.StringValue(result.AssignedObjectType)
	data.Project = types.StringValue(result.Project)
	data.Vendor = types.StringValue(result.Vendor)
	data.Price = types.StringValue(result.Price)
	data.Currency = types.StringValue(result.Currency)
	data.WarrantyStart = types.StringValue(result.WarrantyStart)
	data.WarrantyEnd = types.StringValue(result.WarrantyEnd)
	data.Comments = types.StringValue(result.Comments)
	if result.AssignedObjectID != nil {
		data.AssignedObjectID = types.Int64Value(*result.AssignedObjectID)
	} else {
		data.AssignedObjectID = types.Int64Null()
	}
	if result.Quantity != nil {
		data.Quantity = types.Int64Value(*result.Quantity)
	} else {
		data.Quantity = types.Int64Null()
	}
}

func (r *AssetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data assetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := assetModelToAPI(data)
	var result assetAPIRead
	if err := r.client.Create("/assets/", body, &result); err != nil {
		resp.Diagnostics.AddError("Error creating asset", err.Error())
		return
	}
	data.ID = types.Int64Value(result.ID)
	assetAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AssetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data assetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result assetAPIRead
	if err := r.client.Read("/assets/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", &result); err != nil {
		resp.Diagnostics.AddError("Error reading asset", err.Error())
		return
	}
	assetAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AssetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data assetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := assetModelToAPI(data)
	var result assetAPIRead
	if err := r.client.Update("/assets/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", body, &result); err != nil {
		resp.Diagnostics.AddError("Error updating asset", err.Error())
		return
	}
	assetAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AssetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data assetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// The plugin raises ProtectedError when order_contract is set on the asset.
	// Null it out first so the DELETE succeeds, then the contract can also be deleted.
	if !data.OrderContractID.IsNull() && data.OrderContractID.ValueInt64() != 0 {
		var result assetAPIRead
		if err := r.client.Update("/assets/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", map[string]any{"order_contract": nil}, &result); err != nil {
			resp.Diagnostics.AddError("Error clearing asset order_contract before delete", err.Error())
			return
		}
	}
	if err := r.client.Delete("/assets/" + strconv.FormatInt(data.ID.ValueInt64(), 10) + "/"); err != nil {
		resp.Diagnostics.AddError("Error deleting asset", err.Error())
	}
}

func (r *AssetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric ID, got: "+req.ID)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
