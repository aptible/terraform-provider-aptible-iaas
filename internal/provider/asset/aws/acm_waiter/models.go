package acmwaiter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/aptible/terraform-provider-aptible-iaas/internal/client"
	assetutil "github.com/aptible/terraform-provider-aptible-iaas/internal/provider/asset/util"
)

var resourceTypeName = "_aws_acm_waiter"
var resourceDescription = "ACM certificate waiter resource"

// TODO - autogenerated
type ResourceModel struct {
	Id             types.String `tfsdk:"id" json:"id"`
	AssetVersion   types.String `tfsdk:"asset_version" json:"asset_version"`
	EnvironmentId  types.String `tfsdk:"environment_id" json:"environment_id"`
	OrganizationId types.String `tfsdk:"organization_id" json:"organization_id"`
	Status         types.String `tfsdk:"status" json:"status"`

	CertificateArn  types.String `tfsdk:"certificate_arn"`
	ValidationFqdns types.List   `tfsdk:"validation_fqdns"`
}

var AssetSchema = map[string]tfsdk.Attribute{
	"id": {
		Description: "A valid asset id",
		Type:        types.StringType,
		Computed:    true,
	},
	"status": {
		Type:     types.StringType,
		Computed: true,
	},

	"environment_id": {
		Description: "A valid environment id",
		Type:        types.StringType,
		Required:    true,
	},
	"organization_id": {
		Description: "A valid organization id",
		Type:        types.StringType,
		Required:    true,
	},
	"asset_version": {
		Type:     types.StringType,
		Computed: true,
	},
	"certificate_arn": {
		Required: true,
		Type:     types.StringType,
	},
	"validation_fqdns": {
		Optional:    true,
		Type:        types.ListType{ElemType: types.StringType},
		Description: "The DNS Records created to enable validation. This should include the validation records for both the primary domain and any SANs. Do not set if using EMAIL based validation.",
	},
}

func planToAssetInput(ctx context.Context, plan ResourceModel) (cac.AssetInput, error) {
	params := map[string]interface{}{
		"certificate_arn": plan.CertificateArn.Value,
	}

	if !plan.ValidationFqdns.IsNull() && !plan.ValidationFqdns.IsUnknown() {
		domains := []string{}
		_ = plan.ValidationFqdns.ElementsAs(ctx, &domains, false)
		params["validation_fqdns"] = domains
	}

	input := cac.AssetInput{
		Asset:           client.CompileAsset("aws", "acm_certificate_waiter", assetutil.DefaultAssetVersion),
		AssetVersion:    assetutil.DefaultAssetVersion,
		AssetParameters: params,
	}

	return input, nil
}

func assetOutputToPlan(ctx context.Context, plan ResourceModel, output *cac.AssetOutput) (*ResourceModel, error) {
	domains := []attr.Value{}
	for _, d := range output.CurrentAssetParameters.Data["validation_fqdns"].([]interface{}) {
		domains = append(domains, types.String{Value: d.(string)})
	}
	validationFqdns := types.List{Elems: domains, ElemType: types.StringType}
	if len(domains) == 0 {
		validationFqdns.Null = true
	}

	model := &ResourceModel{
		Id:              types.String{Value: output.Id},
		AssetVersion:    types.String{Value: output.AssetVersion},
		EnvironmentId:   types.String{Value: output.Environment.Id},
		OrganizationId:  types.String{Value: output.Environment.Organization.Id},
		Status:          types.String{Value: string(output.Status)},
		CertificateArn:  types.String{Value: output.CurrentAssetParameters.Data["certificate_arn"].(string)},
		ValidationFqdns: validationFqdns,
	}

	return model, nil
}
