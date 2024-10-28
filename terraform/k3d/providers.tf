terraform {
  required_providers {
    k3d = {
      source = "nikhilsbhat/k3d"
      version = "0.0.2"
    }
  }
}

provider "k3d" {}
