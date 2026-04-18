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

var _ resource.Resource = &ContractorResource{}

type ContractorResource struct{ client *Client }

func NewContractorResource() resource.Resource { return &ContractorResource{} }

func (r *ContractorResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_contractor"
}

func (r *ContractorResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an inventory-monitor contractor.",
		Attributes: map[string]schema.Attribute{
			"id":          schemaID(),
			"name":        schemaRequiredString("Contractor contact name."),
			"tenant_id":   schema.Int64Attribute{Required: true, MarkdownDescription: "NetBox tenant ID."},
			"company":     schemaOptionalString("Company name."),
			"address":     schemaOptionalString("Physical address."),
			"description": schemaOptionalString("Description."),
			"comments":    schemaOptionalString("Comments."),
		},
	}
}

func (r *ContractorResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type contractorModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	TenantID    types.Int64  `tfsdk:"tenant_id"`
	Company     types.String `tfsdk:"company"`
	Address     types.String `tfsdk:"address"`
	Description types.String `tfsdk:"description"`
	Comments    types.String `tfsdk:"comments"`
}

type contractorAPI struct {
	ID          int64    `json:"id,omitempty"`
	Name        string   `json:"name"`
	Tenant      nestedID `json:"tenant"`
	Company     string   `json:"company,omitempty"`
	Address     string   `json:"address,omitempty"`
	Description string   `json:"description,omitempty"`
	Comments    string   `json:"comments,omitempty"`
}

type contractorAPIRead struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Tenant      nestedID `json:"tenant"`
	Company     string   `json:"company"`
	Address     string   `json:"address"`
	Description string   `json:"description"`
	Comments    string   `json:"comments"`
}

func (r *ContractorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data contractorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := contractorAPI{
		Name:        data.Name.ValueString(),
		Tenant:      nestedID{ID: data.TenantID.ValueInt64()},
		Company:     data.Company.ValueString(),
		Address:     data.Address.ValueString(),
		Description: data.Description.ValueString(),
		Comments:    data.Comments.ValueString(),
	}
	var result contractorAPIRead
	if err := r.client.Create("/contractors/", body, &result); err != nil {
		resp.Diagnostics.AddError("Error creating contractor", err.Error())
		return
	}
	data.ID = types.Int64Value(result.ID)
	data.Name = types.StringValue(result.Name)
	data.TenantID = types.Int64Value(result.Tenant.ID)
	data.Company = types.StringValue(result.Company)
	data.Address = types.StringValue(result.Address)
	data.Description = types.StringValue(result.Description)
	data.Comments = types.StringValue(result.Comments)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContractorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data contractorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result contractorAPIRead
	if err := r.client.Read("/contractors/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", &result); err != nil {
		resp.Diagnostics.AddError("Error reading contractor", err.Error())
		return
	}
	data.Name = types.StringValue(result.Name)
	data.TenantID = types.Int64Value(result.Tenant.ID)
	data.Company = types.StringValue(result.Company)
	data.Address = types.StringValue(result.Address)
	data.Description = types.StringValue(result.Description)
	data.Comments = types.StringValue(result.Comments)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContractorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data contractorModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := contractorAPI{
		Name:        data.Name.ValueString(),
		Tenant:      nestedID{ID: data.TenantID.ValueInt64()},
		Company:     data.Company.ValueString(),
		Address:     data.Address.ValueString(),
		Description: data.Description.ValueString(),
		Comments:    data.Comments.ValueString(),
	}
	var result contractorAPIRead
	if err := r.client.Update("/contractors/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", body, &result); err != nil {
		resp.Diagnostics.AddError("Error updating contractor", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ContractorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data contractorModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete("/contractors/" + strconv.FormatInt(data.ID.ValueInt64(), 10) + "/"); err != nil {
		resp.Diagnostics.AddError("Error deleting contractor", err.Error())
	}
}

func (r *ContractorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric ID, got: "+req.ID)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
