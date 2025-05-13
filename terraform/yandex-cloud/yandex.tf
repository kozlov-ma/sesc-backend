# This OpenTofu script creates a virtual machine on Yandex Cloud.

# Variables for sensitive data and configuration

variable "ssh_pubkey_path" {
  description = "Path to SSH public key"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "yc_token" {
  description = "Yandex Cloud API token"
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
  description = "Yandex Cloud Zone"
  type        = string
  default     = "ru-central1-a"
}

variable "vm_name" {
  description = "Name of the VM"
  type        = string
  default     = "coolify-host"
}

variable "ssh_username" {
  description = "SSH username for the VM"
  type        = string
  default     = "ubuntu"
}

variable "coolify_email" {
  description = "Email for Coolify admin account"
  type        = string
  default     = ""
}

variable "coolify_password" {
  description = "Password for Coolify admin account"
  type        = string
  sensitive   = true
  default     = ""
}

# Configure the Yandex Cloud Provider
terraform {
  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = "~> 0.95.0"
    }
  }
}

provider "yandex" {
  token     = var.yc_token
  cloud_id  = var.yc_cloud_id
  folder_id = var.yc_folder_id
  zone      = var.yc_zone
}

data "yandex_compute_image" "ubuntu_image" {
  family = "ubuntu-2404-lts"
}

resource "yandex_vpc_subnet" "subnet" {
  name           = "${var.vm_name}-subnet"
  zone           = var.yc_zone
  network_id     = yandex_vpc_network.network.id
  v4_cidr_blocks = ["192.168.10.0/24"]
}

resource "yandex_compute_instance" "coolify_vm" {
  name        = var.vm_name
  platform_id = "standard-v1"
  zone        = var.yc_zone

  resources {
    cores  = 4
    memory = 4
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu_image.id
      size     = 40
    }
  }

  network_interface {
    subnet_id          = yandex_vpc_subnet.subnet.id
    nat                = true
    security_group_ids = [yandex_vpc_security_group.coolify_sg.id]
  }

  metadata = {
    ssh-keys = "${var.ssh_username}:${file(var.ssh_pubkey_path)}"
  }
}

resource "yandex_vpc_network" "network" {
  name = "${var.vm_name}-network"
}

resource "yandex_vpc_security_group" "coolify_sg" {
  name        = "coolify-security-group"
  description = "Security group for Coolify server"
  network_id  = yandex_vpc_network.network.id

  ingress {
    protocol       = "TCP"
    description    = "SSH"
    v4_cidr_blocks = ["0.0.0.0/0"]
    port           = 22
  }

  ingress {
    protocol       = "TCP"
    description    = "Coolify Dashboard HTTP"
    v4_cidr_blocks = ["0.0.0.0/0"]
    port           = 8000
  }

  ingress {
    protocol       = "TCP"
    description    = "HTTPS"
    v4_cidr_blocks = ["0.0.0.0/0"]
    port           = 443
  }

  ingress {
    protocol       = "TCP"
    description    = "HTTP"
    v4_cidr_blocks = ["0.0.0.0/0"]
    port           = 80
  }

  egress {
    protocol       = "ANY"
    description    = "All outgoing traffic"
    v4_cidr_blocks = ["0.0.0.0/0"]
    from_port      = 0
    to_port        = 65535
  }
}

output "public_ip" {
  value = yandex_compute_instance.coolify_vm.network_interface[0].nat_ip_address
}

output "ssh_command" {
  value = "ssh ${var.ssh_username}@${yandex_compute_instance.coolify_vm.network_interface[0].nat_ip_address}"
}
