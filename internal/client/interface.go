package client

import (
	"context"

	cac "github.com/aptible/cloud-api-clients/clients/go"
)

/*
CloudClient
The goal of this interface is to be an abstraction layer above the cloud-api.
Whenever we want to interface with the API, we should use this interface.
*/
type CloudClient interface {
	ListEnvironments(ctx context.Context, orgId string) ([]cac.EnvironmentOutput, error)
	DescribeEnvironment(ctx context.Context, orgId, envId string) (*cac.EnvironmentOutput, error)
	CreateEnvironment(ctx context.Context, orgId string, params cac.EnvironmentInput) (*cac.EnvironmentOutput, error)
	DestroyEnvironment(ctx context.Context, orgId, envId string) error

	ListOrgs(ctx context.Context) ([]cac.OrganizationOutput, error)
	CreateOrg(ctx context.Context, orgId string, params cac.OrganizationInput) (*cac.OrganizationOutput, error)
	FindOrg(ctx context.Context, orgId string) (*cac.OrganizationOutput, error)

	ListAssetBundles(ctx context.Context, orgId, envId string) ([]cac.AssetBundle, error)
	CreateAsset(ctx context.Context, orgId, envId string, params cac.AssetInput) (*cac.AssetOutput, error)
	ListAssets(ctx context.Context, orgId, envId string) ([]cac.AssetOutput, error)
	DescribeAsset(ctx context.Context, orgId, envId, assetId string) (*cac.AssetOutput, error)
	DestroyAsset(ctx context.Context, orgId, envId, assetID string) error
	UpdateAsset(ctx context.Context, assetId string, envId string, orgId string, params cac.AssetInput) (*cac.AssetOutput, error)

	ListOperationsByAsset(ctx context.Context, orgId, assetId string) ([]cac.OperationOutput, error)

	CreateConnection(ctx context.Context, orgId, envId, assetId string, params cac.ConnectionInput) (*cac.ConnectionOutput, error)
	DestroyConnection(ctx context.Context, orgId, envId, assetId, connectionId string) error
	GetConnection(ctx context.Context, orgId, envId, assetId, connectionId string) (*cac.ConnectionOutput, error)
}
