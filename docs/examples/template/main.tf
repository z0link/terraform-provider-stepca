terraform {
  required_providers {
    stepca = {
      source  = "z0link/stepca"
      version = "0.1.0"
    }
  }
}

provider "stepca" {
  ca_url      = "https://localhost:9000"
  admin_name  = "admin"
  admin_key   = file("/etc/step-ca/admin.jwk")
  token       = var.token
  admin_token = var.admin_token
}

resource "stepca_template" "ssh_admin" {
  name = "ssh-admin"
  body = jsonencode({
    principals = ["admin"]
    extensions = {
      forceCommand = "sudo -l"
    }
  })
  metadata = {
    type    = "ssh"
    version = "1"
  }
}
