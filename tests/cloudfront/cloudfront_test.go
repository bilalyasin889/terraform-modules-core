package test

import (
	"fmt"
	"os"
	cloudfrontUtils "terraform-modules-core/tests/cloudfront/utils"
	coreUtils "terraform-modules-core/tests/utils"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/stretchr/testify/assert"
)

var tld string
var zoneID string
var certArn string

func TestMain(m *testing.M) {
	zoneID, tld = coreUtils.GetDefaultHostedZone()
	certArn = coreUtils.GetWildCardCertificateARN()

	os.Exit(m.Run())
}

// ----------------------------
// BASIC TEST
// ----------------------------
func TestBasicCloudFront(t *testing.T) {
	opts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../../modules/cloudfront",
		Vars: map[string]interface{}{
			"name":    "basic-custom-origin",
			"enabled": true,

			"domains": []string{
				fmt.Sprintf("cdn.%s", tld),
			},
			"certificate_arn": certArn,

			"origins": map[string]interface{}{
				"origin1": map[string]interface{}{
					"type":        "custom",
					"domain_name": "example.com",
				},
			},

			"default_cache_behavior": map[string]interface{}{
				"origin_id": "origin1",
			},
		},
		Logger: logger.Default,
	})

	terraform.WorkspaceSelectOrNew(t, opts, "basic")
	defer terraform.Destroy(t, opts)
	terraform.InitAndApply(t, opts)

	cloudfrontId := terraform.Output(t, opts, "cloudfront_distribution_id")
	assert.NotEmpty(t, cloudfrontId)

	distribution := cloudfrontUtils.GetCloudFrontDistribution(t, cloudfrontId)
	assert.NotNil(t, distribution)

	cloudfrontArn := terraform.Output(t, opts, "cloudfront_distribution_arn")
	assert.Equal(t, *distribution.ARN, cloudfrontArn)

	cloudfrontDomainName := terraform.Output(t, opts, "cloudfront_distribution_domain_name")
	assert.Equal(t, *distribution.DomainName, cloudfrontDomainName)

	cloudfrontHostedZoneId := terraform.Output(t, opts, "cloudfront_distribution_hosted_zone_id")
	assert.NotEmpty(t, cloudfrontHostedZoneId)

	cloudfrontStatus := terraform.Output(t, opts, "cloudfront_distribution_status")
	assert.Equal(t, "Deployed", cloudfrontStatus)
	assert.Equal(t, "Deployed", *distribution.Status)

	distributionConfig := distribution.DistributionConfig
	assert.Equal(t, true, *distributionConfig.Enabled)

	assert.Equal(t, fmt.Sprintf("cdn.%s", tld), distributionConfig.Aliases.Items[0])

	// Validate origin
	cloudfrontUtils.AssertOrigin(
		t,
		distributionConfig.Origins.Items,
		"origin1",
		"example.com",
		3,
		10,
		"",
		false,
	)

	// Validate default cache behavior
	defaultBehavior := distributionConfig.DefaultCacheBehavior
	cloudfrontUtils.AssertDefaultCacheBehaviour(
		t,
		defaultBehavior,
		"origin1",
		cloudfrontUtils.DefaultAllowedMethods,
		cloudfrontUtils.DefaultCachedMethods,
		cloudfrontUtils.DefaultViewerProtocolPolicy,
		cloudfrontUtils.DefaultCachePolicyId,
		false,
	)

	// Validate cache behaviors
	assert.Empty(t, distributionConfig.CacheBehaviors.Items)

	// Validate custom error responses
	assert.Empty(t, distributionConfig.CustomErrorResponses.Items)

	// Validate price class
	assert.Equal(t, types.PriceClassPriceClassAll, distributionConfig.PriceClass)

	// Validate restrictions
	assert.Equal(t, types.GeoRestrictionTypeNone, distributionConfig.Restrictions.GeoRestriction.RestrictionType)

	// Validate viewer certificate
	assert.Equal(t, certArn, *distributionConfig.ViewerCertificate.ACMCertificateArn)
}

func TestFullFeatureS3CloudFront(t *testing.T) {
	// ----------------------------
	// S3 setup
	// ----------------------------
	bucketName := "terraform-terratest-full"
	bucketRegion := "eu-west-2"
	bucketDomainName := bucketName + ".s3." + bucketRegion + ".amazonaws.com"
	bucketARN := coreUtils.CreateS3Bucket(t, bucketName, bucketRegion)

	// ----------------------------
	// Terraform options
	// ----------------------------
	opts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../../modules/cloudfront",
		Vars: map[string]interface{}{
			"name":            "full-feature-s3-cf",
			"enabled":         true,
			"domains":         []string{fmt.Sprintf("cdn-full.%s", tld)},
			"certificate_arn": certArn,
			"hosted_zone_id":  zoneID,

			"origins": map[string]interface{}{
				"origin1": map[string]interface{}{ // S3 default origin
					"type":        "s3",
					"domain_name": bucketDomainName,
					"bucket_arn":  bucketARN,
					"bucket_id":   bucketName,
				},
				"origin2": map[string]interface{}{
					"type":        "custom",
					"domain_name": "example.com",
				},
			},

			"default_cache_behavior": map[string]interface{}{
				"origin_id":              "origin1",
				"allowed_methods":        []string{"GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"},
				"cached_methods":         []string{"GET", "HEAD"},
				"viewer_protocol_policy": "redirect-to-https",
				"functions": []map[string]interface{}{
					{
						"event_type":   "viewer-request",
						"function_key": "my-default-function",
					},
				},
			},

			"ordered_cache_behaviors": []map[string]interface{}{
				{
					"path_pattern":           "/images/*",
					"target_origin_id":       "origin2",
					"allowed_methods":        []string{"GET", "HEAD"},
					"cached_methods":         []string{"GET", "HEAD"},
					"viewer_protocol_policy": "redirect-to-https",
				},
			},

			"custom_error_responses": []map[string]interface{}{
				{
					"error_code":            404,
					"response_code":         200,
					"response_page_path":    "/index.html",
					"error_caching_min_ttl": 0,
				},
				{
					"error_code":            500,
					"response_code":         502,
					"response_page_path":    "/error.html",
					"error_caching_min_ttl": 0,
				},
			},

			"restrictions": map[string]interface{}{
				"type":      "none",
				"locations": []string{},
			},

			"price_class": "PriceClass_All",

			"cloudfront_functions": map[string]interface{}{
				"my-default-function": map[string]interface{}{
					"name":    "my-default-function",
					"runtime": "cloudfront-js-2.0",
					"code":    "function handler(event) { return event.request; }",
					"publish": true,
				},
			},
		},
		Logger: logger.Default,
	})

	// ----------------------------
	// Apply Terraform
	// ----------------------------
	terraform.InitAndApply(t, opts)

	// ----------------------------
	// Retrieve CloudFront distribution
	// ----------------------------
	cloudfrontId := terraform.Output(t, opts, "cloudfront_distribution_id")
	assert.NotEmpty(t, cloudfrontId)

	distribution := cloudfrontUtils.GetCloudFrontDistribution(t, cloudfrontId)
	assert.NotNil(t, distribution)

	cloudfrontArn := terraform.Output(t, opts, "cloudfront_distribution_arn")
	assert.Equal(t, *distribution.ARN, cloudfrontArn)

	cloudfrontDomainName := terraform.Output(t, opts, "cloudfront_distribution_domain_name")
	assert.Equal(t, *distribution.DomainName, cloudfrontDomainName)

	cloudfrontHostedZoneId := terraform.Output(t, opts, "cloudfront_distribution_hosted_zone_id")
	assert.NotEmpty(t, cloudfrontHostedZoneId)

	cloudfrontStatus := terraform.Output(t, opts, "cloudfront_distribution_status")
	assert.Equal(t, "Deployed", cloudfrontStatus)
	assert.Equal(t, "Deployed", *distribution.Status)

	distributionConfig := distribution.DistributionConfig
	assert.True(t, *distributionConfig.Enabled)
	assert.Len(t, distributionConfig.Origins.Items, 2)
	assert.Len(t, distributionConfig.CacheBehaviors.Items, 1)
	assert.Len(t, distributionConfig.CustomErrorResponses.Items, 2)
	assert.Equal(t, types.PriceClassPriceClassAll, distributionConfig.PriceClass)
	assert.Equal(t, types.GeoRestrictionTypeNone, distributionConfig.Restrictions.GeoRestriction.RestrictionType)
	assert.Equal(t, certArn, *distributionConfig.ViewerCertificate.ACMCertificateArn)

	// ----------------------------
	// Validate origins
	// ----------------------------
	cloudfrontUtils.AssertOrigin(
		t,
		distributionConfig.Origins.Items,
		"origin1",
		bucketDomainName,
		3,
		10,
		"",
		true,
	)
	cloudfrontUtils.AssertOrigin(
		t,
		distributionConfig.Origins.Items,
		"origin2",
		"example.com",
		3,
		10,
		"",
		false,
	)

	// ----------------------------
	// Validate default cache behavior
	// ----------------------------
	defaultBehavior := distributionConfig.DefaultCacheBehavior
	cloudfrontUtils.AssertDefaultCacheBehaviour(
		t,
		defaultBehavior,
		"origin1",
		[]string{"GET", "HEAD", "OPTIONS", "PUT", "POST", "PATCH", "DELETE"},
		cloudfrontUtils.DefaultCachedMethods,
		cloudfrontUtils.DefaultViewerProtocolPolicy,
		cloudfrontUtils.DefaultCachePolicyId,
		true,
	)

	// ----------------------------
	// Validate ordered cache behavior
	// ----------------------------
	orderedBehavior := distributionConfig.CacheBehaviors.Items[0]
	cloudfrontUtils.AssertOrderedCacheBehaviour(
		t,
		orderedBehavior,
		"origin2",
		"/images/*",
		[]string{"GET", "HEAD"},
		[]string{"GET", "HEAD"},
		cloudfrontUtils.DefaultViewerProtocolPolicy,
		cloudfrontUtils.DefaultCachePolicyId,
	)

	// ----------------------------
	// Validate custom error responses
	// ----------------------------
	errorResponse1 := distributionConfig.CustomErrorResponses.Items[0]
	cloudfrontUtils.AssertCustomErrorResponse(
		t,
		errorResponse1,
		404,
		"200",
		"/index.html",
		0,
	)

	// ----------------------------
	// Validate DNS records
	// ----------------------------
	cloudfrontUtils.AssertDNSRecordsExist(t, zoneID, []string{fmt.Sprintf("cdn-full.%s", tld)})

	// ----------------------------
	// Cleanup: destroy Terraform first, then delete S3 bucket
	// ----------------------------
	terraform.Destroy(t, opts)
	coreUtils.DeleteS3Bucket(t, bucketName, bucketRegion)
}
