variable "yc_oauth_token" {
  description = "Yandex Cloud OAuth token"
  type        = string
  sensitive   = true
}

variable "yc_cloud_id" {
  description = "Yandex Cloud ID"
  type        = string
}

variable "yc_folder_id" {
  description = "Yandex Cloud Folder ID"
  type        = string
}

variable "yc_zone" {
  description = "Yandex Cloud availability zone"
  type        = string
  default     = "ru-central1-a"
}

variable "project_name" {
  description = "Project name to be used in resource names"
  type        = string
}

variable "main_domain" {
  description = "Main domain for SSL certificate"
  type        = string
}

variable "additional_domains" {
  description = "Additional domains for SSL certificate"
  type        = list(string)
  default     = []
}

variable "admin_ip_cidr" {
  description = "CIDR block for admin access (for SSH)"
  type        = string
  default     = "0.0.0.0/0"  # Warning: It's better to restrict this in production
}

variable "vm_platform_id" {
  description = "Yandex Compute platform ID"
  type        = string
  default     = "standard-v1"
}

variable "vm_image_id" {
  description = "Yandex Compute image ID"
  type        = string
  # Default is Ubuntu 20.04
  default     = "fd80mrhj8fl2oe87o4e1"
}

variable "vm_user" {
  description = "Username for VM SSH access"
  type        = string
  default     = "ubuntu"
}

variable "ssh_public_key_path" {
  description = "Path to the SSH public key"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

# Frontend VM specs
variable "frontend_cores" {
  description = "Number of vCPUs for frontend VM"
  type        = number
  default     = 2
}

variable "frontend_memory" {
  description = "RAM size for frontend VM in GB"
  type        = number
  default     = 4
}

variable "frontend_disk_size" {
  description = "Disk size for frontend VM in GB"
  type        = number
  default     = 10
}

# Backend VM specs
variable "backend_cores" {
  description = "Number of vCPUs for backend VM"
  type        = number
  default     = 2
}

variable "backend_memory" {
  description = "RAM size for backend VM in GB"
  type        = number
  default     = 4
}

variable "backend_disk_size" {
  description = "Disk size for backend VM in GB"
  type        = number
  default     = 10
}

# Kamal accessories VM specs
variable "accessories_cores" {
  description = "Number of vCPUs for Kamal accessories VM"
  type        = number
  default     = 2
}

variable "accessories_memory" {
  description = "RAM size for Kamal accessories VM in GB"
  type        = number
  default     = 4
}

variable "accessories_disk_size" {
  description = "Disk size for Kamal accessories VM in GB"
  type        = number
  default     = 20
}
