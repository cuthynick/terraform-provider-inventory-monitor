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

var _ resource.Resource = &RMAResource{}

type RMAResource struct{ client *Client }

func NewRMAResource() resource.Resource { return &RMAResource{} }

func (r *RMAResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rma"
}

func (r *RMAResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an inventory-monitor RMA (Return Merchandise Authorization).",
		Attributes: map[string]schema.Attribute{
			"id":                  schemaID(),
			"asset_id":            schema.Int64Attribute{Required: true, MarkdownDescription: "ID of the asset being returned."},
			"issue_description":   schemaRequiredString("Description of the issue requiring an RMA."),
			"rma_number":          schemaOptionalString("RMA reference number from the vendor."),
			"original_serial":     schemaOptionalString("Original serial number (if replaced)."),
			"replacement_serial":  schemaOptionalString("Replacement unit serial number."),
			"status":              schemaOptionalString("RMA status: `pending`, `shipped`, `received`, `investigating`, `approved`, `rejected`, or `completed`."),
			"date_issued":         schemaOptionalString("Date the RMA was issued (YYYY-MM-DD)."),
			"date_replaced":       schemaOptionalString("Date the replacement was received (YYYY-MM-DD)."),
			"vendor_response":     schemaOptionalString("Vendor response notes."),
		},
	}
}

func (r *RMAResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type rmaModel struct {
	ID                types.Int64  `tfsdk:"id"`
	AssetID           types.Int64  `tfsdk:"asset_id"`
	IssueDescription  types.String `tfsdk:"issue_description"`
	RMANumber         types.String `tfsdk:"rma_number"`
	OriginalSerial    types.String `tfsdk:"original_serial"`
	ReplacementSerial types.String `tfsdk:"replacement_serial"`
	Status            types.String `tfsdk:"status"`
	DateIssued        types.String `tfsdk:"date_issued"`
	DateReplaced      types.String `tfsdk:"date_replaced"`
	VendorResponse    types.String `tfsdk:"vendor_response"`
}

type rmaAPIWrite struct {
	Asset             nestedID `json:"asset"`
	IssueDescription  string   `json:"issue_description"`
	RMANumber         string   `json:"rma_number,omitempty"`
	OriginalSerial    string   `json:"original_serial,omitempty"`
	ReplacementSerial string   `json:"replacement_serial,omitempty"`
	Status            string   `json:"status,omitempty"`
	DateIssued        string   `json:"date_issued,omitempty"`
	DateReplaced      string   `json:"date_replaced,omitempty"`
	VendorResponse    string   `json:"vendor_response,omitempty"`
}

type rmaAPIRead struct {
	ID                int64    `json:"id"`
	Asset             nestedID `json:"asset"`
	IssueDescription  string   `json:"issue_description"`
	RMANumber         string   `json:"rma_number"`
	OriginalSerial    string   `json:"original_serial"`
	ReplacementSerial string   `json:"replacement_serial"`
	Status            string   `json:"status"`
	DateIssued        string   `json:"date_issued"`
	DateReplaced      string   `json:"date_replaced"`
	VendorResponse    string   `json:"vendor_response"`
}

func rmaModelToAPI(data rmaModel) rmaAPIWrite {
	return rmaAPIWrite{
		Asset:             nestedID{ID: data.AssetID.ValueInt64()},
		IssueDescription:  data.IssueDescription.ValueString(),
		RMANumber:         data.RMANumber.ValueString(),
		OriginalSerial:    data.OriginalSerial.ValueString(),
		ReplacementSerial: data.ReplacementSerial.ValueString(),
		Status:            data.Status.ValueString(),
		DateIssued:        data.DateIssued.ValueString(),
		DateReplaced:      data.DateReplaced.ValueString(),
		VendorResponse:    data.VendorResponse.ValueString(),
	}
}

func rmaAPIToModel(result rmaAPIRead, data *rmaModel) {
	data.AssetID = types.Int64Value(result.Asset.ID)
	data.IssueDescription = types.StringValue(result.IssueDescription)
	data.RMANumber = types.StringValue(result.RMANumber)
	data.OriginalSerial = types.StringValue(result.OriginalSerial)
	data.ReplacementSerial = types.StringValue(result.ReplacementSerial)
	data.Status = types.StringValue(result.Status)
	data.DateIssued = types.StringValue(result.DateIssued)
	data.DateReplaced = types.StringValue(result.DateReplaced)
	data.VendorResponse = types.StringValue(result.VendorResponse)
}

func (r *RMAResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data rmaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result rmaAPIRead
	if err := r.client.Create("/rmas/", rmaModelToAPI(data), &result); err != nil {
		resp.Diagnostics.AddError("Error creating RMA", err.Error())
		return
	}
	data.ID = types.Int64Value(result.ID)
	rmaAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RMAResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data rmaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result rmaAPIRead
	if err := r.client.Read("/rmas/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", &result); err != nil {
		resp.Diagnostics.AddError("Error reading RMA", err.Error())
		return
	}
	rmaAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RMAResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data rmaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result rmaAPIRead
	if err := r.client.Update("/rmas/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", rmaModelToAPI(data), &result); err != nil {
		resp.Diagnostics.AddError("Error updating RMA", err.Error())
		return
	}
	rmaAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RMAResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data rmaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete("/rmas/" + strconv.FormatInt(data.ID.ValueInt64(), 10) + "/"); err != nil {
		resp.Diagnostics.AddError("Error deleting RMA", err.Error())
	}
}

func (r *RMAResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric ID, got: "+req.ID)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
