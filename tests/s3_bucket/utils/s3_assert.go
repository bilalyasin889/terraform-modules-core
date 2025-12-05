package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
)

func AssertS3PublicAccessBlock(
	t *testing.T,
	client *s3.Client,
	accountId string,
	bucketName string,
	expectedBlockPublicAcls bool,
	expectedIgnorePublicAcls bool,
	expectedBlockPublicPolicy bool,
	expectedRestrictPublicBuckets bool,
) {
	out, err := client.GetPublicAccessBlock(context.Background(), &s3.GetPublicAccessBlockInput{
		Bucket:              &bucketName,
		ExpectedBucketOwner: &accountId,
	})
	if err != nil {
		t.Fatalf("Failed to get Public Access Block: %v", err)
	}

	config := out.PublicAccessBlockConfiguration

	assert.Equal(t, expectedBlockPublicAcls, *config.BlockPublicAcls, "BlockPublicAcls mismatch")
	assert.Equal(t, expectedIgnorePublicAcls, *config.IgnorePublicAcls, "IgnorePublicAcls mismatch")
	assert.Equal(t, expectedBlockPublicPolicy, *config.BlockPublicPolicy, "BlockPublicPolicy mismatch")
	assert.Equal(t, expectedRestrictPublicBuckets, *config.RestrictPublicBuckets, "RestrictPublicBuckets mismatch")
}

func AssertBucketPolicyAllowsGet(
	t *testing.T,
	client *s3.Client,
	accountId string,
	bucketName string,
	allowedDomains []string,
) {
	policy := getBucketPolicyJSON(t, client, accountId, bucketName)
	stmt := policy.Statement[0]

	// Effect
	assert.Equal(t, "Allow", stmt.Effect)

	// Principal
	principal := stmt.Principal
	assert.Equal(t, "*", principal)

	// Action
	actions := toStringSlice(stmt.Action)
	assert.Contains(t, actions, "s3:GetObject")

	// Resource
	expectedRes := fmt.Sprintf("arn:aws:s3:::%s/*", bucketName)
	resources := toStringSlice(stmt.Resource)
	assert.Contains(t, resources, expectedRes)

	// Condition: Referer list
	referers := extractRefererList(stmt.Condition)
	for _, d := range allowedDomains {
		assert.Contains(t, referers, d)
	}
}

func AssertBucketCors(
	t *testing.T,
	client *s3.Client,
	accountId string,
	bucketName string,
	expectedMethods []string,
	expectedOrigins []string,
	expectedHeaders []string,
	expectedMaxAge int32,
) {
	cors := getBucketCorsJSON(t, client, accountId, bucketName)

	assert.Len(t, cors.CORSRules, 1)
	rule := cors.CORSRules[0]

	assert.ElementsMatch(t, expectedMethods, rule.AllowedMethods)
	assert.ElementsMatch(t, expectedOrigins, rule.AllowedOrigins)
	assert.ElementsMatch(t, expectedHeaders, rule.AllowedHeaders)
	assert.Equal(t, expectedMaxAge, *rule.MaxAgeSeconds)
}

func getBucketPolicyJSON(
	t *testing.T,
	client *s3.Client,
	accountId string,
	bucketName string,
) BucketPolicyDocument {
	out, err := client.GetBucketPolicy(context.Background(), &s3.GetBucketPolicyInput{
		Bucket:              &bucketName,
		ExpectedBucketOwner: &accountId,
	})
	if err != nil {
		t.Fatalf("Failed to get policy: %v", err)
	}

	var doc BucketPolicyDocument
	if err := json.Unmarshal([]byte(*out.Policy), &doc); err != nil {
		t.Fatalf("Failed to parse policy JSON: %v", err)
	}

	return doc
}

func getBucketCorsJSON(
	t *testing.T,
	client *s3.Client,
	accountId string,
	bucketName string,
) *s3.GetBucketCorsOutput {
	out, err := client.GetBucketCors(context.Background(), &s3.GetBucketCorsInput{
		Bucket:              &bucketName,
		ExpectedBucketOwner: &accountId,
	})
	if err != nil {
		t.Fatalf("Failed to get bucket CORS: %v", err)
	}

	return out
}

func toStringSlice(v interface{}) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []interface{}:
		out := make([]string, len(t))
		for i, x := range t {
			out[i] = x.(string)
		}
		return out
	case []string:
		return t
	}
	return []string{}
}

func extractRefererList(cond map[string]interface{}) []string {
	if cond == nil {
		return []string{}
	}
	sl := cond["StringLike"].(map[string]interface{})
	return toStringSlice(sl["aws:Referer"])
}
