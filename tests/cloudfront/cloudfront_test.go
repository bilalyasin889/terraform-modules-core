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
	origin := distributionConfig.Origins.Items[0]
	cloudfrontUtils.AssertOrigin(t, origin, "origin1", "example.com", 3, 10, "", false)

	// Validate default cache behavior
	defaultBehavior := distributionConfig.DefaultCacheBehavior
	cloudfrontUtils.AssertDefaultCacheBehaviour(t, defaultBehavior, "origin1", cloudfrontUtils.DefaultAllowedMethods, cloudfrontUtils.DefaultCachedMethods, cloudfrontUtils.DefaultViewerProtocolPolicy, cloudfrontUtils.DefaultCachePolicyId)

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

// ----------------------------
// S3 Origin and hosted zone TEST
// ----------------------------
func TestS3OriginCloudFront(t *testing.T) {
	bucketName := "terraform-terratest-basic"
	bucketRegion := "eu-west-2"
	bucketDomainName := bucketName + ".s3." + bucketRegion + ".amazonaws.com"
	bucketARN := coreUtils.CreateS3Bucket(t, bucketName, bucketRegion)

	opts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../../modules/cloudfront",
		Vars: map[string]interface{}{
			"name":    "s3-origin-hosted-zone",
			"enabled": true,

			"domains": []string{
				fmt.Sprintf("cdn-s3.%s", tld),
			},
			"certificate_arn": certArn,
			"hosted_zone_id":  zoneID,

			"origins": map[string]interface{}{
				"origin1": map[string]interface{}{
					"type":        "s3",
					"domain_name": bucketDomainName,
					"bucket_arn":  bucketARN,
					"bucket_id":   bucketName,
				},
			},

			"default_cache_behavior": map[string]interface{}{
				"origin_id": "origin1",
			},
		},
		Logger: logger.Default,
	})

	terraform.WorkspaceSelectOrNew(t, opts, "s3-origin-hosted-zone")
	defer coreUtils.DeleteS3Bucket(t, bucketName, bucketRegion)
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

	assert.Equal(t, fmt.Sprintf("cdn-s3.%s", tld), distributionConfig.Aliases.Items[0])

	// Validate origin
	origin := distributionConfig.Origins.Items[0]
	cloudfrontUtils.AssertOrigin(t, origin, "origin1", bucketDomainName, 3, 10, "", true)

	// Validate default cache behavior
	defaultBehavior := distributionConfig.DefaultCacheBehavior
	cloudfrontUtils.AssertDefaultCacheBehaviour(t, defaultBehavior, "origin1", cloudfrontUtils.DefaultAllowedMethods, cloudfrontUtils.DefaultCachedMethods, cloudfrontUtils.DefaultViewerProtocolPolicy, cloudfrontUtils.DefaultCachePolicyId)

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

	// Validate DNS records
	cloudfrontUtils.AssertDNSRecordsExist(t, zoneID, []string{fmt.Sprintf("cdn-s3.%s", tld)})
}
