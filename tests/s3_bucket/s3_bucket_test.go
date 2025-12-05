package test

import (
	"fmt"
	"os"
	s3Utils "terraform-modules-core/tests/s3_bucket/utils"
	coreUtils "terraform-modules-core/tests/utils"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/stretchr/testify/assert"
)

const REGION = "eu-west-2"

var s3Client *s3.Client
var accountId string

func TestMain(m *testing.M) {
	s3Client = aws.NewS3Client(nil, REGION)

	id, err := coreUtils.GetAWSAccountID(REGION)
	if err != nil {
		fmt.Printf("Error retrieving AWS account ID: %v\n", err)
		os.Exit(1)
	} else {
		accountId = id
	}

	os.Exit(m.Run())
}

// ----------------------------
// BASIC TEST
// ----------------------------
func TestS3BucketBasic(t *testing.T) {
	bucketName := "terraform-terratest-basic"
	tags := map[string]string{"CreatedBy": "terratest"}

	opts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../../modules/s3_bucket",
		Vars: map[string]interface{}{
			"bucket_name": bucketName,
			"tags":        tags,
		},
		Logger: logger.Default,
	})

	terraform.WorkspaceSelectOrNew(t, opts, "basic")
	defer terraform.Destroy(t, opts)
	terraform.InitAndApply(t, opts)

	aws.AssertS3BucketExists(t, REGION, bucketName)

	assert.Equal(t, bucketName, terraform.Output(t, opts, "bucket_name"))
	assert.Equal(t, bucketName, terraform.Output(t, opts, "bucket_id"))

	actualTags := aws.GetS3BucketTags(t, REGION, bucketName)
	assert.Equal(t, tags, actualTags)
}

// ----------------------------
// PUBLIC READ ONLY
// ----------------------------
func TestS3BucketPublicRead(t *testing.T) {
	bucketName := "terraform-terratest-public-read"
	tags := map[string]string{"CreatedBy": "terratest"}
	allowedDomains := []string{
		"https://example.com",
		"https://www.example.com",
	}

	opts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../../modules/s3_bucket",
		Vars: map[string]interface{}{
			"bucket_name": bucketName,
			"tags":        tags,
			"public_read": map[string]interface{}{
				"enabled":         true,
				"allowed_domains": allowedDomains,
			},
		},
		Logger: logger.Default,
	})

	terraform.WorkspaceSelectOrNew(t, opts, "public-read")
	defer terraform.Destroy(t, opts)
	terraform.InitAndApply(t, opts)

	aws.AssertS3BucketExists(t, REGION, bucketName)

	// Validate public access block is set as expected
	s3Utils.AssertS3PublicAccessBlock(t, s3Client, accountId, bucketName, true, true, false, false)

	// Validate bucket policy contains expected values
	s3Utils.AssertBucketPolicyAllowsGet(t, s3Client, accountId, bucketName, allowedDomains)
}

// ----------------------------
// PUT CORS ONLY
// ----------------------------
func TestS3BucketPutCors(t *testing.T) {
	bucketName := "terraform-terratest-cors"
	allowedDomains := []string{
		"https://example.com",
		"https://www.example.com",
	}
	opts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../../modules/s3_bucket",
		Vars: map[string]interface{}{
			"bucket_name": bucketName,
			"tags":        map[string]string{"CreatedBy": "terratest"},
			"put_cors": map[string]interface{}{
				"enabled":         true,
				"allowed_domains": allowedDomains,
			},
		},
		Logger: logger.Default,
	})

	terraform.WorkspaceSelectOrNew(t, opts, "cors")
	defer terraform.Destroy(t, opts)
	terraform.InitAndApply(t, opts)

	aws.AssertS3BucketExists(t, REGION, bucketName)

	expectedMethods := []string{"PUT"}
	expectedHeaders := []string{"*"}
	expectedMaxAge := int32(300)
	s3Utils.AssertBucketCors(t, s3Client, accountId, bucketName, expectedMethods, allowedDomains, expectedHeaders, expectedMaxAge)
}

// ----------------------------
// VERSIONING + FORCE DESTROY
// ----------------------------
func TestS3BucketVersioningForceDestroy(t *testing.T) {
	bucketName := "terraform-terratest-versioning"

	opts := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../../modules/s3_bucket",
		Vars: map[string]interface{}{
			"bucket_name":         bucketName,
			"tags":                map[string]string{"CreatedBy": "terratest"},
			"versioning_enabled":  true,
			"allow_force_destroy": true,
		},
		Logger: logger.Default,
	})

	terraform.WorkspaceSelectOrNew(t, opts, "versioning")
	defer terraform.Destroy(t, opts)
	terraform.InitAndApply(t, opts)

	aws.AssertS3BucketExists(t, REGION, bucketName)

	versioning := aws.GetS3BucketVersioning(t, REGION, bucketName)
	assert.Equal(t, "Enabled", versioning)
}
