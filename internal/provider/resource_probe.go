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

var _ resource.Resource = &ProbeResource{}

type ProbeResource struct{ client *Client }

func NewProbeResource() resource.Resource { return &ProbeResource{} }

func (r *ProbeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_probe"
}

func (r *ProbeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an inventory-monitor probe (a discovery snapshot linking a serial number to a device/site/location at a point in time).",
		Attributes: map[string]schema.Attribute{
			"id":                  schemaID(),
			"name":                schemaRequiredString("Probe name."),
			"time":                schemaRequiredString("Discovery timestamp (RFC3339, e.g. `2024-01-15T10:30:00Z`)."),
			"serial":              schemaRequiredString("Discovered serial number."),
			"device_id":           schema.Int64Attribute{Required: true, MarkdownDescription: "NetBox device ID."},
			"site_id":             schema.Int64Attribute{Required: true, MarkdownDescription: "NetBox site ID."},
			"location_id":         schema.Int64Attribute{Required: true, MarkdownDescription: "NetBox location ID."},
			"part":                schemaOptionalString("Part number discovered."),
			"device_descriptor":   schemaOptionalString("Raw device descriptor string from discovery."),
			"site_descriptor":     schemaOptionalString("Raw site descriptor string from discovery."),
			"location_descriptor": schemaOptionalString("Raw location descriptor string from discovery."),
			"description":         schemaOptionalString("Description."),
			"category":            schemaOptionalString("Probe category."),
			"comments":            schemaOptionalString("Comments."),
		},
	}
}

func (r *ProbeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type probeModel struct {
	ID                 types.Int64  `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Time               types.String `tfsdk:"time"`
	Serial             types.String `tfsdk:"serial"`
	DeviceID           types.Int64  `tfsdk:"device_id"`
	SiteID             types.Int64  `tfsdk:"site_id"`
	LocationID         types.Int64  `tfsdk:"location_id"`
	Part               types.String `tfsdk:"part"`
	DeviceDescriptor   types.String `tfsdk:"device_descriptor"`
	SiteDescriptor     types.String `tfsdk:"site_descriptor"`
	LocationDescriptor types.String `tfsdk:"location_descriptor"`
	Description        types.String `tfsdk:"description"`
	Category           types.String `tfsdk:"category"`
	Comments           types.String `tfsdk:"comments"`
}

type probeAPIWrite struct {
	Name               string   `json:"name"`
	Time               string   `json:"time"`
	Serial             string   `json:"serial"`
	Device             nestedID `json:"device"`
	Site               nestedID `json:"site"`
	Location           nestedID `json:"location"`
	Part               string   `json:"part,omitempty"`
	DeviceDescriptor   string   `json:"device_descriptor,omitempty"`
	SiteDescriptor     string   `json:"site_descriptor,omitempty"`
	LocationDescriptor string   `json:"location_descriptor,omitempty"`
	Description        string   `json:"description,omitempty"`
	Category           string   `json:"category,omitempty"`
	Comments           string   `json:"comments,omitempty"`
}

type probeAPIRead struct {
	ID                 int64    `json:"id"`
	Name               string   `json:"name"`
	Time               string   `json:"time"`
	Serial             string   `json:"serial"`
	Device             nestedID `json:"device"`
	Site               nestedID `json:"site"`
	Location           nestedID `json:"location"`
	Part               string   `json:"part"`
	DeviceDescriptor   string   `json:"device_descriptor"`
	SiteDescriptor     string   `json:"site_descriptor"`
	LocationDescriptor string   `json:"location_descriptor"`
	Description        string   `json:"description"`
	Category           string   `json:"category"`
	Comments           string   `json:"comments"`
}

func probeModelToAPI(data probeModel) probeAPIWrite {
	return probeAPIWrite{
		Name:               data.Name.ValueString(),
		Time:               data.Time.ValueString(),
		Serial:             data.Serial.ValueString(),
		Device:             nestedID{ID: data.DeviceID.ValueInt64()},
		Site:               nestedID{ID: data.SiteID.ValueInt64()},
		Location:           nestedID{ID: data.LocationID.ValueInt64()},
		Part:               data.Part.ValueString(),
		DeviceDescriptor:   data.DeviceDescriptor.ValueString(),
		SiteDescriptor:     data.SiteDescriptor.ValueString(),
		LocationDescriptor: data.LocationDescriptor.ValueString(),
		Description:        data.Description.ValueString(),
		Category:           data.Category.ValueString(),
		Comments:           data.Comments.ValueString(),
	}
}

func probeAPIToModel(result probeAPIRead, data *probeModel) {
	data.Name = types.StringValue(result.Name)
	data.Time = types.StringValue(result.Time)
	data.Serial = types.StringValue(result.Serial)
	data.DeviceID = types.Int64Value(result.Device.ID)
	data.SiteID = types.Int64Value(result.Site.ID)
	data.LocationID = types.Int64Value(result.Location.ID)
	data.Part = types.StringValue(result.Part)
	data.DeviceDescriptor = types.StringValue(result.DeviceDescriptor)
	data.SiteDescriptor = types.StringValue(result.SiteDescriptor)
	data.LocationDescriptor = types.StringValue(result.LocationDescriptor)
	data.Description = types.StringValue(result.Description)
	data.Category = types.StringValue(result.Category)
	data.Comments = types.StringValue(result.Comments)
}

func (r *ProbeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data probeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result probeAPIRead
	if err := r.client.Create("/probes/", probeModelToAPI(data), &result); err != nil {
		resp.Diagnostics.AddError("Error creating probe", err.Error())
		return
	}
	data.ID = types.Int64Value(result.ID)
	probeAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProbeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data probeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result probeAPIRead
	if err := r.client.Read("/probes/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", &result); err != nil {
		resp.Diagnostics.AddError("Error reading probe", err.Error())
		return
	}
	probeAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProbeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data probeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var result probeAPIRead
	if err := r.client.Update("/probes/"+strconv.FormatInt(data.ID.ValueInt64(), 10)+"/", probeModelToAPI(data), &result); err != nil {
		resp.Diagnostics.AddError("Error updating probe", err.Error())
		return
	}
	probeAPIToModel(result, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProbeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data probeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Delete("/probes/" + strconv.FormatInt(data.ID.ValueInt64(), 10) + "/"); err != nil {
		resp.Diagnostics.AddError("Error deleting probe", err.Error())
	}
}

func (r *ProbeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", "Expected a numeric ID, got: "+req.ID)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
