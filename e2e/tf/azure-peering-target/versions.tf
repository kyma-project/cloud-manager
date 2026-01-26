terraform {
  required_version = ">= 1.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.55.0"
    }
    random = {
      source = "hashicorp/random"
      version = "3.7.2"
    }
  }
}
