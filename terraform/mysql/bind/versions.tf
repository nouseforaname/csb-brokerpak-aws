terraform {
  required_providers {
    mysql = {
      source  = "hashicorp/mysql"
      version = ">= 1.9"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.3.2"
    }
  }
}