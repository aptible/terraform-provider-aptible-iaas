package rds_create

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
)

const (
	// VPCCountRequested - maximum vpc count requested to allow for multiple tests
	VPCCountRequested = 10
	// VPCQuotaCode - Corresponds to the AWS quota code that allows to change the maximum number of VPCs per Region.
	// This quota is directly tied to the maximum number of internet gateways per Region.
	VPCQuotaCode = "L-F678F1CE"
	// IGWQuotaCode - Corresponds to the AWS quota code that allows to change the maximum number of internet gateways per Region.
	// This quota is directly tied to the maximum number of VPCs per Region. To increase this quota, increase the number of VPCs per Region.
	IGWQuotaCode = "L-A4707A72"
	// MinutesToWaitForQuota - total minutes to wait before erroring out
	MinutesToWaitForQuota = 25
	// TotalTimeToWaitForQuota - total time to wait for quota before erroring out
	TotalTimeToWaitForQuota = time.Minute * MinutesToWaitForQuota
)

// waitForQuotaValueToBeActive - wait for quota to come online by repeatedly waiting on the aws sdk
func waitForQuotaValueToBeActive(ctx context.Context, c *servicequotas.Client, quotaCode, serviceCode string, expectedValue float64) error {
	timeToError := time.Now().Add(TotalTimeToWaitForQuota)
	for {
		if time.Now().After(timeToError) {
			return fmt.Errorf("Too much time has passed (%d minutes) waiting for quota to become active\n", MinutesToWaitForQuota)
		}

		quota, err := c.GetServiceQuota(ctx, &servicequotas.GetServiceQuotaInput{
			QuotaCode:   aws.String(quotaCode),
			ServiceCode: aws.String(serviceCode),
		})
		if err != nil {
			return err
		}

		if quota.Quota.ErrorReason != nil {
			return fmt.Errorf("service quote errored with reason: (%v) - %s", quota.Quota.ErrorReason.ErrorCode.Values(), *quota.Quota.ErrorReason.ErrorMessage)
		}

		if *quota.Quota.Value == expectedValue {
			return nil
		}

		log.Println(fmt.Sprintf("Still waiting for %s: %s with value of %v. "+
			"Will wait till %s. Sleeping 30s", quotaCode, serviceCode, expectedValue, timeToError.String()))
		time.Sleep(30 * time.Second)
	}
}

// requestStatusIncreaseIfApplicable - request a quota increase if needed
func requestStatusIncreaseIfApplicable(ctx context.Context, c *servicequotas.Client, quotaCode, serviceCode string, expectedValue float64) error {
	existingRequests, err := c.ListRequestedServiceQuotaChangeHistoryByQuota(ctx, &servicequotas.ListRequestedServiceQuotaChangeHistoryByQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	})
	if err != nil {
		return err
	}

	alreadyRequested := false
	var requestStatus types.RequestStatus
	for _, request := range existingRequests.RequestedQuotas {
		if *request.DesiredValue == expectedValue &&
			(request.Status == types.RequestStatusApproved || request.Status == types.RequestStatusPending || request.Status == types.RequestStatusDenied || request.Status == types.RequestStatusCaseOpened) {
			requestStatus = request.Status
			alreadyRequested = true
		}
	}

	// if previously denied, try to request it again just in case
	if !alreadyRequested || requestStatus == types.RequestStatusDenied {
		result, err := c.RequestServiceQuotaIncrease(ctx, &servicequotas.RequestServiceQuotaIncreaseInput{
			DesiredValue: &expectedValue,
			QuotaCode:    aws.String(quotaCode),
			ServiceCode:  aws.String(serviceCode),
		})
		if err != nil {
			return err
		}
		requestStatus = result.RequestedQuota.Status
	}

	// if requested no matter the outcome, break, something is sad :(
	if requestStatus == types.RequestStatusDenied {
		return fmt.Errorf("service quota change denied with status: %s", requestStatus)
	}

	log.Println(fmt.Sprintf("Found request for status, no error needed right now. %s", requestStatus))

	return nil
}

// CheckOrRequestVPCLimit - this should request a limit increase on vpcs in an account if there are an insufficient
// amount already present
func CheckOrRequestVPCLimit() error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	c := servicequotas.NewFromConfig(cfg)
	vpcQuotaOutput, err := c.GetServiceQuota(ctx, &servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String(VPCQuotaCode),
		ServiceCode: aws.String("vpc"),
	})
	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf("Successfully retrieved quota from servicequotas for VPC: %v", *vpcQuotaOutput.Quota.Value))

	if *vpcQuotaOutput.Quota.Value != VPCCountRequested {
		log.Println(fmt.Sprintf("VPC Quota output does not match requested value, requesting changes from "+
			"%v to %v", *vpcQuotaOutput.Quota.Value, VPCCountRequested))

		valueToUse := float64(VPCCountRequested)
		if quotaRequestErr := requestStatusIncreaseIfApplicable(ctx, c, VPCQuotaCode, "vpc", valueToUse); quotaRequestErr != nil {
			return quotaRequestErr
		}
	}

	if quotaAchievedErr := waitForQuotaValueToBeActive(ctx, c, VPCQuotaCode, "vpc", VPCCountRequested); quotaAchievedErr != nil {
		return quotaAchievedErr
	}

	igwQuotaOutput, err := c.GetServiceQuota(ctx, &servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String(IGWQuotaCode),
		ServiceCode: aws.String("vpc"),
	})
	if err != nil {
		return err
	}

	log.Println(fmt.Sprintf("Successfully retrieved quota from servicequotas for IGW: %v", *igwQuotaOutput.Quota.Value))

	if *igwQuotaOutput.Quota.Value != VPCCountRequested {
		log.Println(fmt.Sprintf("Internet Gateway Quota output does not match requested value, requesting changes from "+
			"%v to %v", *vpcQuotaOutput.Quota.Value, VPCCountRequested))

		valueToUse := float64(VPCCountRequested)
		if quotaRequestErr := requestStatusIncreaseIfApplicable(ctx, c, IGWQuotaCode, "vpc", valueToUse); quotaRequestErr != nil {
			return quotaRequestErr
		}
	}

	if quotaAchievedErr := waitForQuotaValueToBeActive(ctx, c, IGWQuotaCode, "vpc", VPCCountRequested); quotaAchievedErr != nil {
		return quotaAchievedErr
	}

	log.Println("Checked and/or updated all service quotas value. Proceeding to test.")

	return nil
}
