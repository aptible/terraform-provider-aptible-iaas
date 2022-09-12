package models

import "github.com/hashicorp/terraform-plugin-framework/tfsdk"

type EntityInterfaceImpl interface {
	ToTFSchema() map[string]tfsdk.Attribute
}
