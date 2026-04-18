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

var _ resource.Resource = &ContractResource{}

type ContractResource struct{ client *Client }

func NewContractResource() resource.Resource { return &ContractResource{} }

func (r *ContractResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contract"
}

func (r *ContractResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an inventory-monitor contract.\n\n> **Note:** The upstream plugin API has a bug where `parent_id` is incorrectly marked required by the serializer. This provider works around it. See https://github.com/CESNET/inventory-monitor-plugin/issues for details.",
		Attributes: map[string]schema.Attribute{
			"id":               schemaID(),
			"name":             schemaRequiredString("Contract display name."),
			"name_internal":    schemaRequiredString("Internal reference number."),
			"contractor_id":    schema.Int64Attribute{Required: true, MarkdownDescription: "ID of the contractor."},
			"type":             schemaRequiredString("Contract type: `supply`, `order`, `service`, or `other`."),
			"parent_id":        schema.Int64Attribute{Optional: true, MarkdownDescription: "ID of the parent contract (for subcontracts)."},
			"description":      schemaOptionalString("Description."),
			"comments":         schemaOptionalString("Comments."),
			"price":            schemaOptionalString("Contract price (decimal string, e.g. `1234.56`)."),
			"currency":         schemaOptionalString("Currency code (required when price is set)."),
			"signed":           schemaOptionalString("Date signed (YYYY-MM-DD)."),
			"accepted":         schemaOptionalString("Date accepted (YYYY-MM-DD)."),
			"invoicing_start":  schemaOptionalString("Invoicing start date (YYYY-MM-DD)."),
			"invoicing_end":    schemaOptionalString("Invoicing end date (YYYY-MM-DD)."),
		},
	}
}

func (r *ContractResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type contractModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	NameInternal   types.String `tfsdk:"name_internal"`
	ContractorID   types.Int64  `tfsdk:"contractor_id"`
	Type           types.String `tfsdk:"type"`
	ParentID       types.Int64  `tfsdk:"parent_id"`
	Description    types.String `tfsdk:"description"`
	Comments       types.String `tfsdk:"comments"`
	Price          types.String `tfsdk:"price"`
	Currency       types.String `tfsdk:"currency"`
	Signed         types.String `tfsdk:"signed"`
	Accepted       types.String `tfsdk:"accepted"`
	InvoicingStart types.String `tfsdk:"invoicing_start"`
	InvoicingEnd   types.String `tfsdk:"invoicing_end"`
}

// contractAPIWrite omits parent when zero so the null bug is avoided.
type contractAPIWrite struct {
	Name           string    `json:"name"`
	NameInternal   string    `json:"name_internal"`
	Contractor     nestedID  `json:"contractor"`
	Type           string    `json:"type"`
	Parent         *nestedID `json:"parent,omitempty"`
	Description    string    `json:"description,omitempty"`
	Comments       string    `json:"comments,omitempty"`
	Price          string    `json:"price,omitempty"`
	Currency       string    `json:"currency,omitempty"`
	Signed         string    `json:"signed,omitempty"`
	Accepted       string    `json:"accepted,omitempty"`
	InvoicingStart string    `json:"invoicing_start,omitempty"`
	InvoicingEnd   string    `json:"invoicing_end,omitempty"`
}

type contractAPIRead struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	NameInternal   string    `json:"name_internal"`
	Contractor     nestedID  `json:"contractor"`
	Type           string    `json:"type"`
	Parent         *nestedID `json:"parent"`
	Description    string    `json:"description"`
	Comments       string    `json:"comments"`
	Price          string    `json:"price"`
	Currency       string    `json:"currency"`
	Signed         string    `json:"signed"`
	Accepted       string    `json:"accepted"`
	InvoicingStart string    `json:"invoicing_start"`
	InvoicingEnd   string    `json:"invoicing_end"`
}

func contractModelToAPI(data contractModel) contractAPIWrite {
	body := contractAPIWrite{
		Name:           data.Name.ValueString(),
		NameInternal:   data.NameInternal.ValueString(),
		Contractor:     nestedID{ID: data.ContractorID.ValueInt64()},
		Type:           data.Type.ValueString(),
		Description:    data.Description.ValueString(),
		Comments:       data.Comments.ValueString(),
		Price:          data.Price.ValueString(),
		Currency:       data.Currency.ValueString(),
		Signed:         data.Signed.ValueString(),
		Accepted:       data.Accepted.ValueString(),
		InvoicingStart: data.InvoicingStart.ValueString(),
		InvoicingEnd:   data.InvoicingEnd.ValueString(),
	}
	if !data.ParentID.IsNull() && !data.ParentID.IsUnknown() && data.ParentID.ValueInt64() != 0 {
		body.Parent = &nestedID{ID: data.ParentID.ValueInt64()}
	}
	return body
}

func contractAPIToModel(result contractAPIRead, data *contractModel) {
	data.Name = types.StringValue(result.Name)
	data.NameInternal = types.StringValue(result.NameInternal)
	data.ContractorID = types.Int64Value(result.Contractor.ID)
	data.Type = types.StringValue(result.Type)
	data.Description = types.StringValue(result.Description)
	data.Comments = types.StringValue(result.Comments)
	data.Price = types.StringValue(result.Price)
	data.Currency = types.StringValue(result.Currency)
	data.Signed = types.StringValue(result.Signed)
	data.Accepted = types.StringValue(result.Accepted)
	data.InvoicingStart = types.StringValue(result.InvoicingStart)
	data.InvoicingEnd = types.StringValue(result.InvoicingEnd)
	if result.Parent != nil {
		data.ParentID = types.Int64Value(result.Parent.ID)
	} else {
		data.ParentID = types.Int64Null()
	}
}

func (r *ContractResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data contractModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := contractModelToAPI(data)
	var result contractAPIRead
	if err := r.client.Create("/contracts/", body, &result); err != nil {
		resp.Diagnostics.AddError("Error creating contract", err.Error())
		return
	}
	data.ID = types.Int64Value(result.ID)
	contractAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContractResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data contractModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result contractAPIRead
	if err := r.client.Read("/contracts/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", &result); err != nil {
		resp.Diagnostics.AddError("Error reading contract", err.Error())
		return
	}
	contractAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContractResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data contractModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := contractModelToAPI(data)
	var result contractAPIRead
	if err := r.client.Update("/contracts/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", body, &result); err != nil {
		resp.Diagnostics.AddError("Error updating contract", err.Error())
		return
	}
	contractAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContractResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data contractModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete("/contracts/" + strconv.FormatInt(data.ID.ValueInt64(), 10) + "/"); err != nil {
		resp.Diagnostics.AddError("Error deleting contract", err.Error())
	}
}

func (r *ContractResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric ID, got: "+req.ID)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
