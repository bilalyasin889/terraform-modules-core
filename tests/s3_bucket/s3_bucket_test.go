package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/stretchr/testify/assert"

	"fmt"
)

const REGION = "eu-west-2"

func TestS3BucketModule(t *testing.T) {
	scenarios := []struct {
		Name     string
		Vars     map[string]interface{}
		Expected map[string]interface{}
	}{
		{
			Name: "public_read_enabled_and_cors_enabled",
			Vars: map[string]interface{}{
				"bucket_name": "terraform-terratest-bucket-1",
				"tags":        map[string]string{"CreatedBy": "terraform"},
				"public_read": map[string]interface{}{
					"enabled":         true,
					"allowed_domains": []string{"https://example.com"},
				},
				"put_cors": map[string]interface{}{
					"enabled":         true,
					"allowed_domains": []string{"https://example.com"},
				},
			},
			Expected: map[string]interface{}{
				"bucket_name":        "terraform-terratest-bucket-1",
				"arn":                "arn:aws:s3:::terraform-terratest-bucket-1",
				"bucket_domain_name": "terraform-terratest-bucket-1.s3.eu-west-2.amazonaws.com",
				"versioning_enabled": false,
				"public_read":        true,
				"allowed_get_domain": "https://example.com",
				"tags":               map[string]string{"CreatedBy": "terraform"},
			},
		},
		{
			Name: "versioning_enabled_and_force_destroy_enabled",
			Vars: map[string]interface{}{
				"bucket_name":         "terraform-terratest-bucket-2",
				"tags":                map[string]string{"CreatedBy": "terraform"},
				"versioning_enabled":  true,
				"allow_force_destroy": true,
			},
			Expected: map[string]interface{}{
				"bucket_name":        "terraform-terratest-bucket-2",
				"arn":                "arn:aws:s3:::terraform-terratest-bucket-2",
				"bucket_domain_name": "terraform-terratest-bucket-2.s3.eu-west-2.amazonaws.com",
				"versioning_enabled": true,
				"public_read":        false,
				"tags":               map[string]string{"CreatedBy": "terraform"},
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.Name, func(t *testing.T) {
			terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
				TerraformDir: "../../modules/s3_bucket",
				Vars:         scenario.Vars,
				Logger:       logger.Default,
			})

			terraform.WorkspaceSelectOrNew(t, terraformOptions, scenario.Name)

			expected := scenario.Expected

			defer terraform.Destroy(t, terraformOptions)
			terraform.InitAndApply(t, terraformOptions)

			expectBucketName := expected["bucket_name"].(string)

			//Verify bucket exists
			aws.AssertS3BucketExists(t, REGION, expectBucketName)

			// Verify outputs
			actualBucketName := terraform.Output(t, terraformOptions, "bucket_name")
			assert.Equal(t, expectBucketName, actualBucketName)

			actualBucketId := terraform.Output(t, terraformOptions, "bucket_id")
			assert.Equal(t, expectBucketName, actualBucketId)

			expectBucketArn := expected["arn"].(string)
			actualBucketArn := terraform.Output(t, terraformOptions, "bucket_arn")
			assert.Equal(t, expectBucketArn, actualBucketArn)

			expectDomainName := expected["bucket_domain_name"].(string)
			actualDomainName := terraform.Output(t, terraformOptions, "bucket_domain_name")
			assert.Equal(t, expectDomainName, actualDomainName)

			// Validate internal properties
			expectVersioning := expected["versioning_enabled"].(bool)
			actualVersioning := aws.GetS3BucketVersioning(t, REGION, expectBucketName)
			if expectVersioning {
				assert.Equal(t, "Enabled", actualVersioning)
			} else {
				assert.Empty(t, actualVersioning, "Versioning should not be enabled")
			}

			// Check public read access via bucket policy
			expectPublicRead := expected["public_read"].(bool)
			if expectPublicRead {
				actualBucketPolicy := aws.GetS3BucketPolicy(t, REGION, expectBucketName)

				expectedAllowedDomain := expected["allowed_get_domain"].(string)
				expectedBucketPolicy := expectedGetBucketPolicy(expectBucketName, expectedAllowedDomain)

				assert.Contains(t, actualBucketPolicy, expectedBucketPolicy, "Bucket policy should contain expected value", []string{actualBucketPolicy})
			} else {
				policyExists, _ := aws.GetS3BucketPolicyE(t, REGION, expectBucketName)

				assert.Empty(t, policyExists, "Bucket policy should not exist")
			}

			// Check tags
			expectedTags := expected["tags"].(map[string]string)
			actualTags := aws.GetS3BucketTags(t, REGION, expectBucketName)
			assert.Equal(t, expectedTags, actualTags)
		})
	}
}

func expectedGetBucketPolicy(bucketName, domain string) string {
	return fmt.Sprintf(
		`{"Sid":"AllowGetFromAllowedDomains","Effect":"Allow","Principal":"*","Action":"s3:GetObject","Resource":"arn:aws:s3:::%s/*","Condition":{"StringLike":{"aws:Referer":"%s"}}}`,
		bucketName,
		domain,
	)
}
