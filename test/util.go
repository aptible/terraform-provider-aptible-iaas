package test

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

func copy(map1 map[string]interface{}) map[string]interface{} {
	map2 := make(map[string]interface{})
	for key, value := range map1 {
		map2[key] = value
	}
	return map2
}

func checkEnvVars(t *testing.T) {

	if os.Getenv("ENVIRONMENT_ID") == "" {
		t.Errorf("Missing ENVIRONMENT_ID environment variable, set it!")
		return
	}
	if os.Getenv("ORGANIZATION_ID") == "" {
		t.Errorf("Missing ORGANIZATION_ID environment variable, set it!")
		return
	}
	if os.Getenv("APTIBLE_TOKEN") == "" {
		t.Errorf("Missing APTIBLE_TOKEN environment variable, set it!")
		return
	}
	if os.Getenv("APTIBLE_HOST") == "" {
		t.Errorf("Missing APTIBLE_HOST environment variable, set it!")
		return
	}
}

func stripBraces(s string) string {
	return s[1 : len(s)-1]
}

func runTerratestLoop(t *testing.T, terraformOptions *terraform.Options, assertionsFunc func()) {
	terraform.Init(t, terraformOptions)

	// plan
	terraform.Plan(t, terraformOptions)

	// apply
	terraform.Apply(t, terraformOptions)
	assertionsFunc()
}
