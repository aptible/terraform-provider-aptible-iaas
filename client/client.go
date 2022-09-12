package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	cac "github.com/aptible/cloud-api-clients/clients/go"
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

func (c *Client) HandleResponse(r *http.Response) {
	if r == nil {
		fmt.Printf("The HTTP response is nil which means the request was never made.  Are you sure your API domain is set properly? (%s)\n", c.apiClient.GetConfig().Host)
		return
	}
	c.PrintResponse(r)
}

func (c *Client) PrintResponse(r *http.Response) {
	if !c.debug {
		return
	}

	log.Println("--- DEBUG ---")
	reqDump, err := httputil.DumpRequestOut(r.Request, false)
	if err != nil {
		fmt.Println(err)
	}

	log.Printf("REQUEST:\n%s", string(reqDump))

	respDump, err := httputil.DumpResponse(r, true)
	if err != nil {
		log.Println(err)
	}

	log.Printf("RESPONSE:\n%s\n", string(respDump))
}

func (c *Client) ListEnvironments(orgId string) ([]cac.EnvironmentOutput, error) {
	request := c.
		apiClient.
		OrganizationsApi.
		OrganizationGetEnvironments(c.ctx, orgId)
	env, r, err := request.Execute()
	c.HandleResponse(r)
	return env, err
}

func (c *Client) CreateEnvironment(orgId string, params cac.EnvironmentInput) (*cac.EnvironmentOutput, error) {
	request := c.
		apiClient.
		EnvironmentsApi.
		EnvironmentCreate(c.ctx, orgId).
		EnvironmentInput(params)
	env, r, err := request.Execute()
	c.HandleResponse(r)
	return env, err
}

func (c *Client) DestroyEnvironment(orgId string, envId string) error {
	_, r, err := c.
		apiClient.
		EnvironmentsApi.
		EnvironmentDelete(
			c.ctx,
			envId,
			orgId,
		).
		Execute()
	c.HandleResponse(r)
	return err
}

func (c *Client) CreateOrg(orgId string, params cac.OrganizationInput) (*cac.OrganizationOutput, error) {
	request := c.
		apiClient.
		OrganizationsApi.
		OrganizationUpdate(c.ctx, orgId).
		OrganizationInput(params)
	org, r, err := request.Execute()
	c.HandleResponse(r)
	return org, err
}

func (c *Client) FindOrg(orgId string) (*cac.OrganizationOutput, error) {
	org, r, err := c.
		apiClient.
		OrganizationsApi.
		OrganizationGet(c.ctx, orgId).
		Execute()
	c.HandleResponse(r)
	return org, err
}

func (c *Client) CreateAsset(orgId string, envId string, params cac.AssetInput) (*cac.AssetOutput, error) {
	request := c.
		apiClient.
		AssetsApi.
		AssetCreate(
			c.ctx,
			envId,
			orgId,
		).
		AssetInput(params)
	asset, r, err := request.Execute()
	c.HandleResponse(r)
	return asset, err
}

func (c *Client) DestroyAsset(orgId string, envId string, assetId string) error {
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
	c.HandleResponse(r)
	return err
}

func (c *Client) AssetUpdate(assetId string, envId string, orgId string, params cac.AssetInput) (*cac.AssetOutput, error) {
	request := c.
		apiClient.
		AssetsApi.
		AssetUpdate(c.ctx, assetId, envId, orgId).
		AssetInput(params)
	asset, r, err := request.Execute()
	c.HandleResponse(r)
	return asset, err
}

func (c *Client) ListAssets(orgId string, envId string) ([]cac.AssetOutput, error) {
	request := c.apiClient.EnvironmentsApi.EnvironmentGetAssets(
		c.ctx,
		envId,
		orgId,
	)
	assets, r, err := request.Execute()
	c.HandleResponse(r)
	return assets, err
}

func (c *Client) DescribeAsset(orgId string, envId string, assetId string) (*cac.AssetOutput, error) {
	request := c.
		apiClient.
		AssetsApi.
		AssetGet(
			c.ctx,
			assetId,
			envId,
			orgId,
		)
	asset, r, err := request.Execute()
	c.HandleResponse(r)
	return asset, err
}

func (c *Client) ListOrgs() ([]cac.OrganizationOutput, error) {
	request := c.apiClient.OrganizationsApi.OrganizationList(c.ctx)
	orgs, r, err := request.Execute()
	c.HandleResponse(r)
	return orgs, err
}

func (c *Client) ListOperationsByAsset(orgId string, assetId string) ([]cac.OperationOutput, error) {
	request := c.
		apiClient.
		OrganizationsApi.
		OrganizationGetOperations(c.ctx, orgId).
		AssetId(assetId)
	ops, r, err := request.Execute()
	c.HandleResponse(r)
	return ops, err
}

func (c *Client) ListAssetBundles(orgId string, envId string) ([]cac.AssetBundle, error) {
	request := c.
		apiClient.
		EnvironmentsApi.
		EnvironmentGetAllowedAssetBundles(
			c.ctx,
			envId,
			orgId,
		)
	bundles, r, err := request.Execute()
	c.HandleResponse(r)
	return bundles, err
}

func (c *Client) CreateConnection(orgId, envId, assetId string, params cac.ConnectionInput) (*cac.ConnectionOutput, error) {
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
	c.HandleResponse(r)
	return conn, err
}

func (c *Client) DestroyConnection(orgId, envId, assetId, connectionId string) error {
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
	c.HandleResponse(r)
	return err
}
