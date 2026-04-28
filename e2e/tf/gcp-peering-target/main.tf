
data "google_client_config" "current" {}

locals {
  network_name = coalesce(var.network_name, var.name)
  project_id = data.google_client_config.current.project
  subnetwork_name = coalesce(var.subnetwork_name, var.name)
  instance_name = coalesce(var.instance_name, var.name)
}

data "google_tags_tag_key" "tag_key"{
  parent = "projects/${local.project_id}"
  short_name = "e2e"
}
data "google_tags_tag_value" "tag_value"{
  parent = data.google_tags_tag_key.tag_key.id
  short_name = "None"
}

resource "google_compute_network" "vpc_network" {
  name                    = local.network_name
  auto_create_subnetworks = false
  params {
    resource_manager_tags = {
      "${data.google_tags_tag_key.tag_key.id}" = "${data.google_tags_tag_value.tag_value.id}"
    }
  }
}

resource "google_compute_subnetwork" "vpc_subnetwork" {
  name          = local.subnetwork_name
  ip_cidr_range = var.subnet_cidr
  network       = google_compute_network.vpc_network.id
  region        = "${var.location}"
}

resource "google_compute_instance" "vm_instance" {
  name         = local.instance_name
  machine_type = "e2-micro"
  zone         = "${var.location}-b"

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-13"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.vpc_subnetwork.id
  }

  metadata = {
    enable-oslogin = "FALSE"
  }
}

resource "google_compute_firewall" "allow-ssh-on-peering-target" {
   name    = "${local.network_name}-allow-ssh"
   network = google_compute_network.vpc_network.id

   allow {
     protocol = "tcp"
     ports    = ["22"]
   }

   source_ranges = ["10.250.0.0/16"]
}
