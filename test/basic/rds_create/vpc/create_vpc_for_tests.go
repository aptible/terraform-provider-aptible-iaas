package vpc

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

var terraformOptions = &terraform.Options{
	TerraformDir: "../vpc/",
	Vars: map[string]interface{}{
		"organization_id": os.Getenv("ORGANIZATION_ID"),
		"environment_id":  os.Getenv("ENVIRONMENT_ID"),
		"aptible_host":    os.Getenv("APTIBLE_HOST"),
		"vpc_name":        "rds-create-vpc",
	},
}

// AcquireVPCOrCreate - will create a VPC with a lock of its existence
func AcquireVPCOrCreate(t *testing.T, lockName string) error {
	files, err := ioutil.ReadDir("../vpc/locks")
	if err != nil {
		return err
	}

	alreadyRequested := false
	for _, file := range files {
		if file.Name() == ".gitkeep" {
			continue
		}
		alreadyRequested = true
		break
	}

	// write the file anyway in case we'll need it
	if err = ioutil.WriteFile(fmt.Sprintf("../vpc/locks/%s", lockName), []byte{}, 0644); err != nil {
		return err
	}

	if !alreadyRequested {
		// wait for VPC to be online (should already exist)
		terraform.InitAndApply(t, terraformOptions)
	}

	return nil
}

func DeleteVPCIfUnused(t *testing.T, lockName string) error {
	files, err := ioutil.ReadDir("../vpc/locks")
	if err != nil {
		return err
	}

	mapOfLocks := map[string]bool{}
	for _, file := range files {
		if file.Name() == ".gitkeep" {
			continue
		}
		mapOfLocks[file.Name()] = true
	}

	found := mapOfLocks[lockName]
	if found {
		delete(mapOfLocks, lockName)
		if err = os.Remove(fmt.Sprintf("../vpc/locks/%s", lockName)); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unable to find file to clear locks: %s", lockName)
	}

	if len(mapOfLocks) == 0 {
		// safe to destroy, destroy it!
		terraform.Destroy(t, terraformOptions)
	}

	return nil
}
