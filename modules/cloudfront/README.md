# cloudfront Module

## Description
Creates a CloudFront distribution with configurable origins, caching rules, SSL support, and optional custom domains.

---

## Inputs

| Name    | Description            | Type   | Default | Required |
|---------|------------------------|--------|---------|----------|
| input_1 | TODO: describe input_1 | string | -       | yes      |
| input_2 | TODO: describe input_2 | string | -       | no       |

---

## Outputs

| Name        | Description                    |
|-------------|--------------------------------|
| output_1    | TODO: describe output_1        |
| output_2    | TODO: describe output_2        |

---

## Example Usage

```hcl
module "cloudfront" {
  source = "git::ssh://github.com/bilalyasin889/terraform-modules-core.git//modules/cloudfront?ref=vX.Y.Z"

  # TODO: provide actual values
  input_1 = "example_value_1"
  input_2 = "example_value_2"
  input_3 = "example_value_3"
}
```

---

## Testing
[//]: # (TODO: This section is intended for module developers and should be removed before publishing.)

The generated module includes a Go test file for Terratest: `cloudfront_test.go`.

### Prerequisites

Before running the tests, make sure you have:

- [Go](https://golang.org/dl/) installed (version 1.21+ recommended)
- `GOPATH` configured and added to your `PATH` (if needed)
- Terraform installed and available in your `PATH`
- AWS credentials configured if testing AWS resources

### Setup

1. Initialize a Go module in the `test` directory if not already initialized:

    ```bash
    cd test
    go mod init cloudfront_test
    go mod tidy
    ```
    - This will download the Terratest dependencies used in the test.


2. Run the tests using the Go tool:
    ```bash
    go test -v ./cloudfront_test.go
    ```
    - -v enables verbose output
      - Make sure your Terraform backend and AWS environment are properly configured
