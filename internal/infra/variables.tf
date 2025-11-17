variable "yc_oauth_token" {
  type        = string
  description = "Yandex Cloud OAuth token"
  sensitive   = true
}

variable "yc_cloud_id" {
  type        = string
  description = "Yandex Cloud ID"
}

variable "yc_folder_id" {
  type        = string
  description = "Existing folder ID where new folder will be created"
}

variable "yc_zone" {
  type        = string
  description = "Yandex Cloud zone"
  default     = "ru-central1-a"
}

variable "project_name" {
  type        = string
  description = "Project name prefix for resources"
}

variable "admin_ip_cidr" {
  type        = string
  description = "Admin IP CIDR for SSH access"
  default     = "0.0.0.0/0"
}

variable "vm_user" {
  type        = string
  description = "VM user name"
  default     = "ubuntu"
}

variable "ssh_public_key_path" {
  type        = string
  description = "Path to SSH public key"
}
