terraform {
  required_providers {
    yandex = {
      source  = "yandex-cloud/yandex"
      version = "~> 0.89.0"
    }
  }
}

provider "yandex" {
  token     = var.yc_oauth_token
  cloud_id  = var.yc_cloud_id
  folder_id = var.yc_folder_id
  zone      = var.yc_zone
}

# Certificate Manager
data "yandex_cm_certificate" "cert" {
  certificate_id = var.certificate_id
}

# Network resources
resource "yandex_vpc_network" "network" {
  name = "${var.project_name}-network"
}

resource "yandex_vpc_subnet" "subnet" {
  name           = "${var.project_name}-subnet"
  zone           = var.yc_zone
  network_id     = yandex_vpc_network.network.id
  v4_cidr_blocks = ["10.2.0.0/16"]
}

# Security group for VMs
resource "yandex_vpc_security_group" "sg" {
  name       = "${var.project_name}-security-group"
  network_id = yandex_vpc_network.network.id

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

  # Allow HTTPS
  ingress {
    protocol       = "TCP"
    description    = "HTTPS"
    port           = 443
    v4_cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow backend API port (adjust as needed)
  ingress {
    protocol       = "TCP"
    description    = "Backend API"
    port           = 8080
    v4_cidr_blocks = ["10.2.0.0/16"]
  }

  # Allow all outbound traffic
  egress {
    protocol       = "ANY"
    description    = "All outgoing traffic"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}

data "yandex_compute_image" "ubuntu_image" {
  family = "ubuntu-2404-lts"
}

# Create VMs
resource "yandex_compute_instance" "frontend" {
  name        = "${var.project_name}-frontend"
  platform_id = var.vm_platform_id
  zone        = var.yc_zone

  resources {
    cores  = var.frontend_cores
    memory = var.frontend_memory
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu_image.id
      size     = var.frontend_disk_size
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

resource "yandex_compute_instance" "backend" {
  name        = "${var.project_name}-backend"
  platform_id = var.vm_platform_id
  zone        = var.yc_zone

  resources {
    cores  = var.backend_cores
    memory = var.backend_memory
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu_image.id
      size     = var.backend_disk_size
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

resource "yandex_compute_instance" "kamal_accessories" {
  name        = "${var.project_name}-kamal-accessories"
  platform_id = var.vm_platform_id
  zone        = var.yc_zone

  resources {
    cores  = var.accessories_cores
    memory = var.accessories_memory
  }

  boot_disk {
    initialize_params {
      image_id = data.yandex_compute_image.ubuntu_image.id
      size     = var.accessories_disk_size
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

# Load Balancer
resource "yandex_alb_load_balancer" "lb" {
  name       = "${var.project_name}-load-balancer"
  network_id = yandex_vpc_network.network.id

  allocation_policy {
    location {
      zone_id   = var.yc_zone
      subnet_id = yandex_vpc_subnet.subnet.id
    }
  }

  listener {
    name = "https-listener"
    endpoint {
      address {
        external_ipv4_address {}
      }
      ports = [443]
    }
    tls {
      default_handler {
        certificate_ids = [data.yandex_cm_certificate.cert.id]
        http_handler {
          http_router_id = yandex_alb_http_router.router.id
        }
      }
    }
  }

  listener {
    name = "http-listener"
    endpoint {
      address {
        external_ipv4_address {}
      }
      ports = [80]
    }
    http {
      handler {
        http_router_id = yandex_alb_http_router.router.id
      }
    }
  }
}

# HTTP Router and virtual hosts
resource "yandex_alb_http_router" "router" {
  name = "${var.project_name}-router"
}

# Frontend virtual host for all domains
resource "yandex_alb_virtual_host" "frontend_host" {
  name           = "${var.project_name}-frontend-host"
  http_router_id = yandex_alb_http_router.router.id
  authority      = var.frontend_domains

  route {
    name = "frontend-route"
    http_route {
      http_route_action {
        backend_group_id = yandex_alb_backend_group.frontend_bg.id
      }
    }
  }
}

# API virtual host for all API subdomains
resource "yandex_alb_virtual_host" "api_host" {
  name           = "${var.project_name}-api-host"
  http_router_id = yandex_alb_http_router.router.id
  authority      = var.api_domains

  route {
    name = "api-route"
    http_route {
      http_route_action {
        backend_group_id = yandex_alb_backend_group.backend_bg.id
      }
    }
  }
}

# Frontend backend group
resource "yandex_alb_backend_group" "frontend_bg" {
  name = "${var.project_name}-frontend-bg"

  http_backend {
    name             = "frontend-backend"
    weight           = 1
    port             = 80
    target_group_ids = [yandex_alb_target_group.frontend_tg.id]
    healthcheck {
      timeout  = "10s"
      interval = "2s"
      http_healthcheck {
        path = "/"
      }
    }
  }
}

# Backend backend group
resource "yandex_alb_backend_group" "backend_bg" {
  name = "${var.project_name}-backend-bg"

  http_backend {
    name             = "backend-api"
    weight           = 1
    port             = 8080 # Adjust this port based on your backend application
    target_group_ids = [yandex_alb_target_group.backend_tg.id]
    healthcheck {
      timeout  = "10s"
      interval = "2s"
      http_healthcheck {
        path = "/health" # Adjust this path based on your backend health endpoint
      }
    }
  }
}

# Frontend target group
resource "yandex_alb_target_group" "frontend_tg" {
  name = "${var.project_name}-frontend-tg"

  target {
    subnet_id  = yandex_vpc_subnet.subnet.id
    ip_address = yandex_compute_instance.frontend.network_interface.0.ip_address
  }
}

# Backend target group
resource "yandex_alb_target_group" "backend_tg" {
  name = "${var.project_name}-backend-tg"

  target {
    subnet_id  = yandex_vpc_subnet.subnet.id
    ip_address = yandex_compute_instance.backend.network_interface.0.ip_address
  }
}

# Outputs
output "load_balancer_ip" {
  description = "External IP address of the load balancer"
  value       = yandex_alb_load_balancer.lb.listener[0].endpoint[0].address[0].external_ipv4_address[0].address
}

output "frontend_internal_ip" {
  description = "Internal IP address of the frontend VM"
  value       = yandex_compute_instance.frontend.network_interface.0.ip_address
}

output "backend_internal_ip" {
  description = "Internal IP address of the backend VM"
  value       = yandex_compute_instance.backend.network_interface.0.ip_address
}

output "dns_records_needed" {
  description = "DNS records that need to be configured"
  value = {
    load_balancer_ip = yandex_alb_load_balancer.lb.listener[0].endpoint[0].address[0].external_ipv4_address[0].address
    frontend_domains = var.frontend_domains
    api_domains      = var.api_domains
    note             = "All domains should point to the load_balancer_ip"
  }
}
