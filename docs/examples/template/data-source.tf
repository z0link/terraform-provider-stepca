data "stepca_template" "ssh_admin" {
  name = "ssh-admin"
}

output "ssh_admin_principals" {
  value = jsondecode(data.stepca_template.ssh_admin.body).principals
}

output "ssh_admin_metadata" {
  value = data.stepca_template.ssh_admin.metadata
}
