//go:build integration

// Integration tests that hit a live Kion instance.
//
// This file is behind the `integration` build tag because:
//  1. It requires real credentials (KION_URL and KION_API_KEY).
//  2. It targets a specific generated sub-package (v3_16) whose types
//     and methods are pinned to that API version. If you want to run
//     integration tests against a different version, bump the import
//     and update the type references accordingly.
//
// Run with: make test-integration
// Or:       KION_URL=... KION_API_KEY=... go test -tags integration -v ./...
package kion_test

import (
	"context"
	"os"
	"testing"

	kion "github.com/kionsoftware/kion-sdk-go"
	v3_16 "github.com/kionsoftware/kion-sdk-go/generated/v3_16"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// integrationClient is shared across all integration subtests.
var integrationClient *v3_16.Client

// skipIfNoCredentials skips the test if KION_URL and KION_API_KEY are not set.
func skipIfNoCredentials(t *testing.T) {
	t.Helper()
	if os.Getenv("KION_URL") == "" || os.Getenv("KION_API_KEY") == "" {
		t.Skip("KION_URL and KION_API_KEY must be set for integration tests")
	}
}

// setupClient creates a shared client for integration tests. Call once per test run.
func setupClient(t *testing.T) *v3_16.Client {
	t.Helper()
	if integrationClient != nil {
		return integrationClient
	}

	url := os.Getenv("KION_URL")
	key := os.Getenv("KION_API_KEY")

	client, err := v3_16.New(url,
		kion.WithAPIKey(key),
		kion.WithSkipVerify(true),
	)
	require.NoError(t, err, "failed to create Kion client")
	integrationClient = client
	return client
}

func TestIntegrationListEndpoints(t *testing.T) {
	skipIfNoCredentials(t)
	c := setupClient(t)
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func(t *testing.T, c *v3_16.Client)
	}{
		// ---- account ----
		{"account", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetAccountIndex(ctx, v3_16.GetAccountIndexParams{})
			require.NoError(t, err)
			resp, ok := res.(*v3_16.AccountListResponse)
			require.True(t, ok, "expected *AccountListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- account_cache ----
		{"account_cache", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetAccountCacheIndex(ctx)
			require.NoError(t, err)
			resp, ok := res.(*v3_16.AccountCacheListResponse)
			require.True(t, ok, "expected *AccountCacheListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- ami ----
		{"ami", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetAMIIndex(ctx)
			require.NoError(t, err)
			resp, ok := res.(*v3_16.AMIListResponse)
			require.True(t, ok, "expected *AMIListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- app_api_key ----
		{"app_api_key", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetAppAPIKeyIndex(ctx)
			require.NoError(t, err)
			resp, ok := res.(*v3_16.AppAPIKeyListResponse)
			require.True(t, ok, "expected *AppAPIKeyListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- azure_arm_template ----
		{"azure_arm_template", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetAzureARMTemplateIndex(ctx)
			require.NoError(t, err)
			resp, ok := res.(*v3_16.AzureARMTemplateListResponse)
			require.True(t, ok, "expected *AzureARMTemplateListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- billing_rule (paginated) ----
		{"billing_rule", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetBillingRuleIndex(ctx, v3_16.GetBillingRuleIndexParams{})
			require.NoError(t, err)
			resp, ok := res.(*v3_16.PaginatedBillingRuleListResponse)
			require.True(t, ok, "expected *PaginatedBillingRuleListResponse, got %T", res)
			assert.True(t, resp.Data.Set, "expected Data to be set")
		}},

		// ---- cft ----
		{"cft", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetCFTIndex(ctx)
			require.NoError(t, err)
			resp, ok := res.(*v3_16.CFTListResponseWithOwnersAndTags)
			require.True(t, ok, "expected *CFTListResponseWithOwnersAndTags, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- cloud_rule ----
		{"cloud_rule", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetCloudRuleIndex(ctx, v3_16.GetCloudRuleIndexParams{})
			require.NoError(t, err)
			resp, ok := res.(*v3_16.CloudRuleListResponse)
			require.True(t, ok, "expected *CloudRuleListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- compliance_program (paginated) ----
		{"compliance_program", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetComplianceProgramPaginatedIndex(ctx, v3_16.GetComplianceProgramPaginatedIndexParams{})
			require.NoError(t, err)
			resp, ok := res.(*v3_16.PaginatedComplianceProgramListResponse)
			require.True(t, ok, "expected *PaginatedComplianceProgramListResponse, got %T", res)
			assert.True(t, resp.Data.Set, "expected Data to be set")
		}},

		// ---- compliance_standard ----
		{"compliance_standard", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetComplianceStandardIndex(ctx, v3_16.GetComplianceStandardIndexParams{})
			require.NoError(t, err)
			resp, ok := res.(*v3_16.ComplianceStandardListResponse)
			require.True(t, ok, "expected *ComplianceStandardListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- custom_variable (global list, paginated) ----
		{"custom_variable", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetAllGlobalCustomVariables(ctx, v3_16.GetAllGlobalCustomVariablesParams{})
			require.NoError(t, err)
			resp, ok := res.(*v3_16.PaginatedCustomVariableListResponse)
			require.True(t, ok, "expected *PaginatedCustomVariableListResponse, got %T", res)
			assert.True(t, resp.Data.Set, "expected Data to be set")
		}},

		// ---- iam_policy ----
		{"iam_policy", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetIAMPolicyIndex(ctx)
			require.NoError(t, err)
			resp, ok := res.(*v3_16.IAMPolicyListResponse)
			require.True(t, ok, "expected *IAMPolicyListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- idms ----
		{"idms", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetIDMSIndex(ctx)
			require.NoError(t, err)
			resp, ok := res.(*v3_16.IDMSListResponse)
			require.True(t, ok, "expected *IDMSListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- label ----
		{"label", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetLabelIndex(ctx, v3_16.GetLabelIndexParams{})
			require.NoError(t, err)
			resp, ok := res.(*v3_16.LabelListPaginatedResponse)
			require.True(t, ok, "expected *LabelListPaginatedResponse, got %T", res)
			assert.True(t, resp.Data.Set, "expected Data to be set")
		}},

		// ---- permission_scheme ----
		{"permission_scheme", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetPermissionSchemeIndex(ctx)
			require.NoError(t, err)
			resp, ok := res.(*v3_16.PermissionSchemeListResponse)
			require.True(t, ok, "expected *PermissionSchemeListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- project ----
		{"project", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetProjectIndex(ctx, v3_16.GetProjectIndexParams{})
			require.NoError(t, err)
			resp, ok := res.(*v3_16.ProjectListResponse)
			require.True(t, ok, "expected *ProjectListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- service_control_policy ----
		{"service_control_policy", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetServiceControlPolicyIndex(ctx)
			require.NoError(t, err)
			resp, ok := res.(*v3_16.ServiceControlPolicyListResponse)
			require.True(t, ok, "expected *ServiceControlPolicyListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},

		// ---- user_group ----
		{"user_group", func(t *testing.T, c *v3_16.Client) {
			res, err := c.GetUGroupIndex(ctx, v3_16.GetUGroupIndexParams{})
			require.NoError(t, err)
			resp, ok := res.(*v3_16.UGroupListResponse)
			require.True(t, ok, "expected *UGroupListResponse, got %T", res)
			assert.NotNil(t, resp.Data)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fn(t, c)
		})
	}
}
