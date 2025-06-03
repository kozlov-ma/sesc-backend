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

# Certificate Manager
resource "yandex_cm_certificate" "main_cert" {
  name    = "${var.project_name}-certificate"
  domains = concat([var.main_domain], var.additional_domains, ["*.${var.main_domain}"])

  managed {
    challenge_type = "DNS_CNAME"
  }
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
  name        = "${var.project_name}-security-group"
  network_id  = yandex_vpc_network.network.id

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

  # Allow all outbound traffic
  egress {
    protocol       = "ANY"
    description    = "All outgoing traffic"
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
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
      image_id = var.vm_image_id
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
      image_id = var.vm_image_id
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
      image_id = var.vm_image_id
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
  name        = "${var.project_name}-load-balancer"
  network_id  = yandex_vpc_network.network.id

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
        certificate_ids = [yandex_cm_certificate.main_cert.id]
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

# HTTP Router and backend groups
resource "yandex_alb_http_router" "router" {
  name = "${var.project_name}-router"
}

resource "yandex_alb_virtual_host" "virtual_host" {
  name           = "${var.project_name}-virtual-host"
  http_router_id = yandex_alb_http_router.router.id

  route {
    name = "frontend-route"
    http_route {
      http_route_action {
        backend_group_id = yandex_alb_backend_group.frontend_bg.id
      }
    }
  }
}

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

resource "yandex_alb_target_group" "frontend_tg" {
  name = "${var.project_name}-frontend-tg"

  target {
    subnet_id  = yandex_vpc_subnet.subnet.id
    ip_address = yandex_compute_instance.frontend.network_interface.0.ip_address
  }
}
