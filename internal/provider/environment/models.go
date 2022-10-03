package environment

import "github.com/hashicorp/terraform-plugin-framework/types"

type Env struct {
	ID    types.String `tfsdk:"id"`
	OrgID types.String `tfsdk:"org_id"`
	Name  types.String `tfsdk:"name"`
}
