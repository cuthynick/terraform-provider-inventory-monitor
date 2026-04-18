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

var _ resource.Resource = &InvoiceResource{}

type InvoiceResource struct{ client *Client }

func NewInvoiceResource() resource.Resource { return &InvoiceResource{} }

func (r *InvoiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_invoice"
}

func (r *InvoiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an inventory-monitor invoice.",
		Attributes: map[string]schema.Attribute{
			"id":              schemaID(),
			"name":            schemaRequiredString("Invoice display name."),
			"name_internal":   schemaRequiredString("Internal invoice reference."),
			"contract_id":     schema.Int64Attribute{Required: true, MarkdownDescription: "ID of the related contract."},
			"project":         schemaOptionalString("Project name."),
			"description":     schemaOptionalString("Description."),
			"price":           schemaOptionalString("Invoice price (decimal string)."),
			"invoicing_start": schemaOptionalString("Invoicing start date (YYYY-MM-DD)."),
			"invoicing_end":   schemaOptionalString("Invoicing end date (YYYY-MM-DD)."),
			"comments":        schemaOptionalString("Comments."),
		},
	}
}

func (r *InvoiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type invoiceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	NameInternal   types.String `tfsdk:"name_internal"`
	ContractID     types.Int64  `tfsdk:"contract_id"`
	Project        types.String `tfsdk:"project"`
	Description    types.String `tfsdk:"description"`
	Price          types.String `tfsdk:"price"`
	InvoicingStart types.String `tfsdk:"invoicing_start"`
	InvoicingEnd   types.String `tfsdk:"invoicing_end"`
	Comments       types.String `tfsdk:"comments"`
}

type invoiceAPIWrite struct {
	Name           string   `json:"name"`
	NameInternal   string   `json:"name_internal"`
	Contract       nestedID `json:"contract"`
	Project        string   `json:"project,omitempty"`
	Description    string   `json:"description,omitempty"`
	Price          string   `json:"price,omitempty"`
	InvoicingStart string   `json:"invoicing_start,omitempty"`
	InvoicingEnd   string   `json:"invoicing_end,omitempty"`
	Comments       string   `json:"comments,omitempty"`
}

type invoiceAPIRead struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	NameInternal   string   `json:"name_internal"`
	Contract       nestedID `json:"contract"`
	Project        string   `json:"project"`
	Description    string   `json:"description"`
	Price          string   `json:"price"`
	InvoicingStart string   `json:"invoicing_start"`
	InvoicingEnd   string   `json:"invoicing_end"`
	Comments       string   `json:"comments"`
}

func invoiceModelToAPI(data invoiceModel) invoiceAPIWrite {
	return invoiceAPIWrite{
		Name:           data.Name.ValueString(),
		NameInternal:   data.NameInternal.ValueString(),
		Contract:       nestedID{ID: data.ContractID.ValueInt64()},
		Project:        data.Project.ValueString(),
		Description:    data.Description.ValueString(),
		Price:          data.Price.ValueString(),
		InvoicingStart: data.InvoicingStart.ValueString(),
		InvoicingEnd:   data.InvoicingEnd.ValueString(),
		Comments:       data.Comments.ValueString(),
	}
}

func invoiceAPIToModel(result invoiceAPIRead, data *invoiceModel) {
	data.Name = types.StringValue(result.Name)
	data.NameInternal = types.StringValue(result.NameInternal)
	data.ContractID = types.Int64Value(result.Contract.ID)
	data.Project = types.StringValue(result.Project)
	data.Description = types.StringValue(result.Description)
	data.Price = types.StringValue(result.Price)
	data.InvoicingStart = types.StringValue(result.InvoicingStart)
	data.InvoicingEnd = types.StringValue(result.InvoicingEnd)
	data.Comments = types.StringValue(result.Comments)
}

func (r *InvoiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data invoiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result invoiceAPIRead
	if err := r.client.Create("/invoices/", invoiceModelToAPI(data), &result); err != nil {
		resp.Diagnostics.AddError("Error creating invoice", err.Error())
		return
	}
	data.ID = types.Int64Value(result.ID)
	invoiceAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InvoiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data invoiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result invoiceAPIRead
	if err := r.client.Read("/invoices/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", &result); err != nil {
		resp.Diagnostics.AddError("Error reading invoice", err.Error())
		return
	}
	invoiceAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InvoiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data invoiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result invoiceAPIRead
	if err := r.client.Update("/invoices/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", invoiceModelToAPI(data), &result); err != nil {
		resp.Diagnostics.AddError("Error updating invoice", err.Error())
		return
	}
	invoiceAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InvoiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data invoiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete("/invoices/" + strconv.FormatInt(data.ID.ValueInt64(), 10) + "/"); err != nil {
		resp.Diagnostics.AddError("Error deleting invoice", err.Error())
	}
}

func (r *InvoiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric ID, got: "+req.ID)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
