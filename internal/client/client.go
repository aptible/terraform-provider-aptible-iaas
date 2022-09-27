package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"

	cac "github.com/aptible/cloud-api-clients/clients/go"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// client - internal cac struct used only for this service with some common configuration
type Client struct {
	ctx context.Context

	apiClient *cac.APIClient
	debug     bool
	token     string
}

// NewClient - generate a new cloud api cloud_api_client
func NewClient(debug bool, host string, token string) CloudClient {
	config := cac.NewConfiguration()
	config.Host = host
	config.Scheme = "https"

	apiClient := cac.NewAPIClient(config)

	ctx := context.Background()
	ctx = context.WithValue(ctx, cac.ContextAccessToken, token)

	return &Client{
		ctx:       ctx,
		apiClient: apiClient,

		debug: debug,
		token: token,
	}
}

func (c *Client) PrintRequestParams(tfctx context.Context, params interface{}) {
	if !c.debug {
		return
	}

	out, err := json.Marshal(params)
	if err != nil {
		tflog.Error(tfctx, err.Error())
		return
	}

	tflog.Debug(tfctx, "REQUEST PARAMS", map[string]interface{}{"params": string(out)})
}

func (c *Client) HandleResponse(tfctx context.Context, r *http.Response) {
	if r == nil {
		fmt.Printf("The HTTP response is nil which means the request was never made.  Are you sure your API domain is set properly? (%s)\n", c.apiClient.GetConfig().Host)
		return
	}
	c.PrintResponse(tfctx, r)
}

func (c *Client) PrintResponse(tfctx context.Context, r *http.Response) {
	if !c.debug {
		return
	}

	reqDump, err := httputil.DumpRequestOut(r.Request, false)
	if err != nil {
		tflog.Error(tfctx, err.Error())
	}

	tflog.Debug(tfctx, "REQUEST", map[string]interface{}{"out": string(reqDump)})

	respDump, err := httputil.DumpResponse(r, true)
	if err != nil {
		tflog.Error(tfctx, err.Error())
	}

	tflog.Debug(tfctx, "RESPONSE", map[string]interface{}{"out": string(respDump)})
}

func (c *Client) ListEnvironments(tfctx context.Context, orgId string) ([]cac.EnvironmentOutput, error) {
	request := c.
		apiClient.
		OrganizationsApi.
		OrganizationGetEnvironments(c.ctx, orgId)
	env, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return env, err
}

func (c *Client) CreateEnvironment(tfctx context.Context, orgId string, params cac.EnvironmentInput) (*cac.EnvironmentOutput, error) {
	request := c.
		apiClient.
		EnvironmentsApi.
		EnvironmentCreate(c.ctx, orgId).
		EnvironmentInput(params)
	env, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return env, err
}

func (c *Client) DescribeEnvironment(tfctx context.Context, orgId string, envId string) (*cac.EnvironmentOutput, error) {
	env, r, err := c.
		apiClient.
		EnvironmentsApi.
		EnvironmentGet(
			c.ctx,
			envId,
			orgId,
		).
		Execute()
	c.HandleResponse(tfctx, r)
	return env, err
}

func (c *Client) DestroyEnvironment(tfctx context.Context, orgId string, envId string) error {
	_, r, err := c.
		apiClient.
		EnvironmentsApi.
		EnvironmentDelete(
			c.ctx,
			envId,
			orgId,
		).
		Execute()
	c.HandleResponse(tfctx, r)
	return err
}

func (c *Client) CreateOrg(tfctx context.Context, orgId string, params cac.OrganizationInput) (*cac.OrganizationOutput, error) {
	request := c.
		apiClient.
		OrganizationsApi.
		OrganizationUpdate(c.ctx, orgId).
		OrganizationInput(params)
	org, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return org, err
}

func (c *Client) FindOrg(tfctx context.Context, orgId string) (*cac.OrganizationOutput, error) {
	org, r, err := c.
		apiClient.
		OrganizationsApi.
		OrganizationGet(c.ctx, orgId).
		Execute()
	c.HandleResponse(tfctx, r)
	return org, err
}

func (c *Client) CreateAsset(tfctx context.Context, orgId string, envId string, params cac.AssetInput) (*cac.AssetOutput, error) {
	c.PrintRequestParams(tfctx, params)

	request := c.apiClient.AssetsApi.
		AssetCreate(
			c.ctx,
			envId,
			orgId,
		).AssetInput(params)
	asset, r, err := request.Execute()

	c.HandleResponse(tfctx, r)

	return asset, err
}

func (c *Client) DestroyAsset(tfctx context.Context, orgId string, envId string, assetId string) error {
	request := c.
		apiClient.
		AssetsApi.
		AssetDelete(
			c.ctx,
			assetId,
			envId,
			orgId,
		)
	_, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return err
}

func (c *Client) UpdateAsset(tfctx context.Context, assetId string, envId string, orgId string, params cac.AssetInput) (*cac.AssetOutput, error) {
	request := c.apiClient.AssetsApi.
		AssetUpdate(c.ctx, assetId, envId, orgId).
		AssetInput(params)
	asset, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return asset, err
}

func (c *Client) ListAssets(tfctx context.Context, orgId string, envId string) ([]cac.AssetOutput, error) {
	request := c.apiClient.EnvironmentsApi.EnvironmentGetAssets(
		c.ctx,
		envId,
		orgId,
	)
	assets, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return assets, err
}

func (c *Client) DescribeAsset(tfctx context.Context, orgId string, envId string, assetId string) (*cac.AssetOutput, error) {
	request := c.apiClient.AssetsApi.
		AssetGet(
			c.ctx,
			assetId,
			envId,
			orgId,
		)
	asset, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return asset, err
}

func (c *Client) ListOrgs(tfctx context.Context) ([]cac.OrganizationOutput, error) {
	request := c.apiClient.OrganizationsApi.OrganizationList(c.ctx)
	orgs, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return orgs, err
}

func (c *Client) ListOperationsByAsset(tfctx context.Context, orgId string, assetId string) ([]cac.OperationOutput, error) {
	request := c.
		apiClient.
		OrganizationsApi.
		OrganizationGetOperations(c.ctx, orgId).
		AssetId(assetId)
	ops, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return ops, err
}

func (c *Client) ListAssetBundles(tfctx context.Context, orgId string, envId string) ([]cac.AssetBundle, error) {
	request := c.
		apiClient.
		EnvironmentsApi.
		EnvironmentGetAllowedAssetBundles(
			c.ctx,
			envId,
			orgId,
		)
	bundles, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return bundles, err
}

func (c *Client) GetConnection(tfctx context.Context, orgId, envId, assetId, connectionId string) (*cac.ConnectionOutput, error) {
	request := c.
		apiClient.
		ConnectionsApi.
		ConnectionGet(
			c.ctx,
			assetId,
			envId,
			connectionId,
			orgId,
		)
	conn, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return conn, err
}

func (c *Client) CreateConnection(tfctx context.Context, orgId, envId, assetId string, params cac.ConnectionInput) (*cac.ConnectionOutput, error) {
	request := c.
		apiClient.
		ConnectionsApi.
		ConnectionCreate(
			c.ctx,
			assetId,
			envId,
			orgId,
		).
		ConnectionInput(params)
	conn, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return conn, err
}

func (c *Client) DestroyConnection(tfctx context.Context, orgId, envId, assetId, connectionId string) error {
	request := c.
		apiClient.
		ConnectionsApi.
		ConnectionDelete(
			c.ctx,
			assetId,
			connectionId,
			envId,
			orgId,
		)
	_, r, err := request.Execute()
	c.HandleResponse(tfctx, r)
	return err
}
