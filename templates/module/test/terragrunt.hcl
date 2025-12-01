# Terragrunt test configuration for <module_name>

terraform {
  source = "../.."

  backend "local" {
    path = "terraform.tfstate"
  }
}

inputs = {
  # TODO: Provide values for required module inputs
  # input_1 = "example"
  # input_2 = "example"
}