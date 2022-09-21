package organization

import "github.com/hashicorp/terraform-plugin-framework/types"

type Org struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
