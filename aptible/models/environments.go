package models

type Environment struct {
	ID    string `tfsdk:"id"`
	OrgID string `tfsdk:"org_id"`
	Name  string `tfsdk:"name"`
}
