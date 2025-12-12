package utils

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/require"
)

// GetAWSAccountID retrieves the AWS account ID for the current credentials.
func GetAWSAccountID(region string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sts.NewFromConfig(cfg)

	out, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("failed to get caller identity: %w", err)
	}

	return *out.Account, nil
}

func GetDefaultHostedZone() (zoneID string, zoneName string) {
	const defaultZoneName = "bilalyasin.com"

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	client := route53.NewFromConfig(cfg)
	resp, err := client.ListHostedZonesByName(context.Background(), &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(defaultZoneName),
	})
	if err != nil || len(resp.HostedZones) == 0 {
		panic(fmt.Sprintf("no hosted zone found for %s", defaultZoneName))
	}

	return strings.TrimPrefix(*resp.HostedZones[0].Id, "/hostedzone/"), defaultZoneName
}

func GetWildCardCertificateARN() string {
	const defaultDomain = "*.bilalyasin.com"

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		panic(fmt.Sprintf("failed to load AWS config: %v", err))
	}

	client := acm.NewFromConfig(cfg)

	resp, err := client.ListCertificates(context.Background(), &acm.ListCertificatesInput{})
	if err != nil {
		panic(fmt.Sprintf("failed to list ACM certificates: %v", err))
	}

	for _, cert := range resp.CertificateSummaryList {
		if aws.ToString(cert.DomainName) == defaultDomain {
			return aws.ToString(cert.CertificateArn)
		}
	}
	panic(fmt.Sprintf("no certificate found for domain %s", defaultDomain))
	return "" // never reached
}

func CreateS3Bucket(t *testing.T, bucketName string, region string) string {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	require.NoError(t, err, "failed to load AWS config")

	client := s3.NewFromConfig(cfg)

	input := &s3.CreateBucketInput{
		Bucket: &bucketName,
	}

	if region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}

	_, err = client.CreateBucket(context.Background(), input)
	require.NoError(t, err, "failed to create S3 Bucket")

	return fmt.Sprintf("arn:aws:s3:::%s", bucketName)
}

func DeleteS3Bucket(t *testing.T, bucketName string, region string) {
	t.Helper()

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	require.NoError(t, err, "failed to load AWS config")

	client := s3.NewFromConfig(cfg)

	input := &s3.DeleteBucketInput{
		Bucket: &bucketName,
	}

	_, err = client.DeleteBucket(context.Background(), input)
	require.NoError(t, err, "failed to delete S3 Bucket")
}
