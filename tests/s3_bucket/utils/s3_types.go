package utils

type BucketPolicyDocument struct {
	Statement []struct {
		Effect    string                 `json:"Effect"`
		Principal string                 `json:"Principal"`
		Action    interface{}            `json:"Action"`
		Resource  interface{}            `json:"Resource"`
		Condition map[string]interface{} `json:"Condition"`
	} `json:"Statement"`
}
