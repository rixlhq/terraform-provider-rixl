terraform {
  required_providers {
    rixl = {
      source  = "rixlhq/rixl"
      version = ">= 0.0.0"
    }
  }
}

provider "rixl" {
  api_key = var.rixl_api_key
}
