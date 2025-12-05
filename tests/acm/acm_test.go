package test

import (
	"context"
	"fmt"
	"os"
	"terraform-modules-core/tests/utils"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

var tld string
var zoneID string

func TestMain(m *testing.M) {
	utils.LoadTestEnv()

	tld = os.Getenv("TEST_TLD")
	zoneID = os.Getenv("TEST_HOSTED_ZONE_ID")

	// Skip if TEST_HOSTED_ZONE_ID is not set
	if zoneID == "" || tld == "" {
		fmt.Printf("Environment variables %v must be set.\n", []string{"TEST_HOSTED_ZONE_ID", "TEST_TLD"})
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestACMModule(t *testing.T) {
	scenarios := []struct {
		Name     string
		Vars     map[string]interface{}
		Expected map[string]interface{}
	}{
		{
			Name: "basic_acm_cert_us_east1",
			Vars: map[string]interface{}{
				"domain_name":    "test." + tld,
				"hosted_zone_id": zoneID,
				"region":         "us-east-1",
				"tags":           map[string]string{"CreatedBy": "terratest"},
			},
			Expected: map[string]interface{}{
				"domain_name": "test." + tld,
				"tags":        map[string]string{"CreatedBy": "terratest"},
				"region":      "us-east-1",
			},
		},
		{
			Name: "acm_cert_with_san",
			Vars: map[string]interface{}{
				"domain_name":               "test." + tld,
				"subject_alternative_names": []string{"www.test." + tld},
				"hosted_zone_id":            zoneID,
				"tags":                      map[string]string{"CreatedBy": "terratest"},
			},
			Expected: map[string]interface{}{
				"domain_name": "test." + tld,
				"tags":        map[string]string{"CreatedBy": "terratest"},
				"san":         []string{"www.test." + tld},
				"region":      "eu-west-2",
			},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.Name, func(t *testing.T) {
			terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
				TerraformDir: "../../modules/acm",
				Vars:         scenario.Vars,
				Logger:       logger.Default,
			})

			terraform.WorkspaceSelectOrNew(t, terraformOptions, scenario.Name)

			defer terraform.Destroy(t, terraformOptions)

			terraform.InitAndApply(t, terraformOptions)

			expected := scenario.Expected

			certArn := aws.GetAcmCertificateArn(t, expected["region"].(string), expected["domain_name"].(string))
			certArnOutput := terraform.Output(t, terraformOptions, "certificate_arn")
			assert.Equal(t, certArn, certArnOutput)

			cert := GetCertificate(t, expected["region"].(string), certArn)
			expectedDomainNames := []string{expected["domain_name"].(string)}
			if expected["san"] != nil {
				expectedDomainNames = append(expectedDomainNames, expected["san"].([]string)...)
			}
			assert.Equal(t, expectedDomainNames, cert.SubjectAlternativeNames)

			assert.Equal(t, types.CertificateStatusIssued, cert.Status)

			tags := ListTags(t, expected["region"].(string), certArn)
			assert.Equal(t, expected["tags"], tags)
		})
	}
}

func GetCertificate(t *testing.T, region string, certArn string) *types.CertificateDetail {
	client := aws.NewAcmClient(t, region)

	input := &acm.DescribeCertificateInput{
		CertificateArn: &certArn,
	}

	output, err := client.DescribeCertificate(context.Background(), input)
	if err != nil {
		t.Fatalf("Failed to describe certificate: %v", err)
	}

	return output.Certificate
}

func ListTags(t *testing.T, region string, certArn string) map[string]string {
	client := aws.NewAcmClient(t, region)

	input := &acm.ListTagsForCertificateInput{
		CertificateArn: &certArn,
	}

	output, err := client.ListTagsForCertificate(context.Background(), input)
	if err != nil {
		t.Fatalf("Failed to list tags: %v", err)
	}

	// Convert to map[string]string
	tagsMap := make(map[string]string, len(output.Tags))
	for _, tag := range output.Tags {
		tagsMap[*tag.Key] = *tag.Value
	}

	return tagsMap
}
