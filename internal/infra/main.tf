terraform {
  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = "~> 0.92.0"
    }
  }
}

provider "yandex" {
  token     = var.yc_oauth_token
  cloud_id  = var.yc_cloud_id
  folder_id = var.yc_folder_id
  zone      = var.yc_zone
}

# Create a new folder
resource "yandex_resourcemanager_folder" "new_folder" {
  cloud_id    = var.yc_cloud_id
  name        = "${var.project_name}-folder"
  description = "Folder for ${var.project_name} resources"
}

# Create network in the new folder
resource "yandex_vpc_network" "network" {
  name        = "${var.project_name}-network"
  folder_id   = yandex_resourcemanager_folder.new_folder.id
  description = "Network for ${var.project_name}"
}

# Create subnet in the new folder
resource "yandex_vpc_subnet" "subnet" {
  name           = "${var.project_name}-subnet"
  folder_id      = yandex_resourcemanager_folder.new_folder.id
  zone           = var.yc_zone
  network_id     = yandex_vpc_network.network.id
  v4_cidr_blocks = ["10.2.0.0/16"]
}

# Security group for VM
resource "yandex_vpc_security_group" "sg" {
  name        = "${var.project_name}-security-group"
  folder_id   = yandex_resourcemanager_folder.new_folder.id
  network_id  = yandex_vpc_network.network.id
  description = "Security group for ${var.project_name} VM"

  # Allow SSH
  ingress {
    protocol       = "TCP"
    description    = "SSH"
    port           = 22
    v4_cidr_blocks = [var.admin_ip_cidr]
  }

  # Allow HTTP
  ingress {
    protocol       = "TCP"
    description    = "HTTP"
    port           = 80
    v4_cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow HTTP
  ingress {
    protocol       = "TCP"
    description    = "HTTP"
    port           = 8080
    v4_cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow HTTPS
  ingress {
    protocol       = "TCP"
    description    = "HTTPS"
    port           = 443
    v4_cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all outbound traffic
  egress {
    protocol       = "ANY"
    description    = "All outgoing traffic"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}

# Get Ubuntu 24.04 image
data "yandex_compute_image" "ubuntu_image" {
  family = "ubuntu-2404-lts"
}

# Create VM in the new folder
resource "yandex_compute_instance" "vm" {
  name        = "${var.project_name}-vm"
  folder_id   = yandex_resourcemanager_folder.new_folder.id
  platform_id = "standard-v3"
  zone        = var.yc_zone

  resources {
    cores  = 2
    memory = 4
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu_image.id
      size     = 20
    }
  }

  network_interface {
    subnet_id          = yandex_vpc_subnet.subnet.id
    nat                = true
    security_group_ids = [yandex_vpc_security_group.sg.id]
  }

  metadata = {
    ssh-keys = "${var.vm_user}:${file(var.ssh_public_key_path)}"
  }
}

# Outputs
output "new_folder_id" {
  description = "ID of the created folder"
  value       = yandex_resourcemanager_folder.new_folder.id
}

output "vm_public_ip" {
  description = "Public IP address of the VM"
  value       = yandex_compute_instance.vm.network_interface[0].nat_ip_address
}

output "vm_ssh_command" {
  description = "SSH command to connect to the VM"
  value       = "ssh ${var.vm_user}@${yandex_compute_instance.vm.network_interface[0].nat_ip_address}"
}
