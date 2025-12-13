package utils

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53Types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetCloudFrontDistribution(
	t *testing.T,
	distributionId string,
) *types.Distribution {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	require.NoError(t, err, "failed to load AWS config")

	client := cloudfront.NewFromConfig(cfg)

	input := cloudfront.GetDistributionInput{
		Id: &distributionId,
	}

	resp, err := client.GetDistribution(context.Background(), &input)
	require.NoError(t, err, "failed to get CloudFront distribution")

	return resp.Distribution
}

func AssertOrigin(
	t *testing.T,
	origins []types.Origin,
	originId string,
	domain string,
	connectionAttempts int,
	connectionTimeout int,
	originPath string,
	hasOAC bool,
) {
	t.Helper()

	for _, origin := range origins {
		if origin.Id != nil && *origin.Id == originId {

			assert.Equal(t, originId, *origin.Id)
			assert.Equal(t, domain, *origin.DomainName)

			assert.Equal(t, int32(connectionAttempts), *origin.ConnectionAttempts)
			assert.Equal(t, int32(connectionTimeout), *origin.ConnectionTimeout)

			assert.Empty(t, origin.CustomHeaders.Items)

			assert.Equal(t, originPath, *origin.OriginPath)

			assert.Equal(t, hasOAC, origin.OriginAccessControlId != nil && *origin.OriginAccessControlId != "")
			return
		}
	}

	t.Fatalf("origin with id %q not found", originId)
}

var DefaultAllowedMethods = []string{"GET", "HEAD", "OPTIONS"}
var DefaultCachedMethods = []string{"GET", "HEAD"}
var DefaultViewerProtocolPolicy = "redirect-to-https"
var DefaultCachePolicyId = "658327ea-f89d-4fab-a63d-7e88639e58f6"

func AssertDefaultCacheBehaviour(
	t *testing.T,
	defaultBehavior *types.DefaultCacheBehavior,
	originId string,
	allowedMethods []string,
	cachedMethods []string,
	viewerProtocolPolicy string,
	cachePolicyId string,
	hasFunctionAssociations bool,
) {
	t.Helper()

	assert.Equal(t, originId, *defaultBehavior.TargetOriginId)

	expectedAllowedMethods := mapHTTPMethods(allowedMethods)
	assert.ElementsMatch(t, defaultBehavior.AllowedMethods.Items, expectedAllowedMethods)

	expectedCachedMethods := mapHTTPMethods(cachedMethods)
	assert.ElementsMatch(t, defaultBehavior.AllowedMethods.CachedMethods.Items, expectedCachedMethods)

	assert.Equal(t, types.ViewerProtocolPolicy(viewerProtocolPolicy), defaultBehavior.ViewerProtocolPolicy)

	assert.Equal(t, cachePolicyId, *defaultBehavior.CachePolicyId)

	if hasFunctionAssociations {
		assert.NotEmpty(t, defaultBehavior.FunctionAssociations.Items)
	} else {
		assert.Empty(t, defaultBehavior.FunctionAssociations.Items)
	}
}

func AssertOrderedCacheBehaviour(
	t *testing.T,
	cacheBehavior types.CacheBehavior,
	originId string,
	pathPattern string,
	allowedMethods []string,
	cachedMethods []string,
	viewerProtocolPolicy string,
	cachePolicyId string,
) {
	t.Helper()

	assert.Equal(t, originId, *cacheBehavior.TargetOriginId)

	assert.Equal(t, pathPattern, *cacheBehavior.PathPattern)

	expectedAllowedMethods := mapHTTPMethods(allowedMethods)
	assert.ElementsMatch(t, cacheBehavior.AllowedMethods.Items, expectedAllowedMethods)

	expectedCachedMethods := mapHTTPMethods(cachedMethods)
	assert.ElementsMatch(t, cacheBehavior.AllowedMethods.CachedMethods.Items, expectedCachedMethods)

	assert.Equal(t, types.ViewerProtocolPolicy(viewerProtocolPolicy), cacheBehavior.ViewerProtocolPolicy)

	assert.Equal(t, cachePolicyId, *cacheBehavior.CachePolicyId)

	assert.Empty(t, cacheBehavior.FunctionAssociations.Items)

}

func mapHTTPMethods(methods []string) []types.Method {
	result := make([]types.Method, len(methods))
	for i, m := range methods {
		result[i] = types.Method(m)
	}
	return result
}

func AssertDNSRecordsExist(
	t *testing.T,
	hostedZoneID string,
	names []string,
) {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(context.Background())
	require.NoError(t, err)

	client := route53.NewFromConfig(cfg)

	for _, name := range names {
		for _, rt := range []route53Types.RRType{route53Types.RRTypeA, route53Types.RRTypeAaaa} {

			resp, err := client.TestDNSAnswer(context.Background(), &route53.TestDNSAnswerInput{
				HostedZoneId: aws.String(hostedZoneID),
				RecordName:   aws.String(name),
				RecordType:   rt,
			})
			require.NoErrorf(t, err, "TestDNSAnswer failed for %s (%s)", name, rt)

			// If we get *any* answer back, the record effectively exists.
			assert.NotEmptyf(t, resp.RecordData,
				"expected an alias DNS answer for %s (%s)", name, rt)

			// Optional: ensure it's an ALIAS
			assert.Equalf(t, rt, resp.RecordType,
				"record type mismatch for %s", name)
		}
	}
}

func AssertCustomErrorResponse(
	t *testing.T,
	errorResponse types.CustomErrorResponse,
	errorCode int,
	responseCode string,
	responsePage string,
	minTTL int,
) {
	t.Helper()

	assert.Equal(t, int32(errorCode), *errorResponse.ErrorCode)

	assert.Equal(t, responseCode, *errorResponse.ResponseCode)

	assert.Equal(t, responsePage, *errorResponse.ResponsePagePath)

	assert.Equal(t, int64(minTTL), *errorResponse.ErrorCachingMinTTL)
}
