# auth to GCP
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.24.0"
    }
  }
}

provider "google" {
  credentials = file("~/.gcp/terraform.json")

  project = "stride-nodes"
  region  = "us-central1"
  zone    = "us-central1-b"
}

data "google_compute_default_service_account" "default" {}

variable "regions" {
  type    = list(string)
  default = ["us-central1"]
}
variable "deployment_name" {
  type    = string
  default = "testnet"
}
variable "network_name" {
  type    = string
  default = "stride"
}

variable "num_nodes" {
  type    = number
  default = 3
}

locals {
  node_names = [
    for i in range(1, var.num_nodes + 1) : "${var.network_name}-node${i}"
  ]
}

module "images" {
  source  = "terraform-google-modules/container-vm/google"
  version = "~> 2.0"

  count = length(local.node_names)
  container = {
    image = "gcr.io/stride-nodes/${var.deployment_name}:${local.node_names[count.index]}"
  }
  restart_policy = "Always"
}

resource "google_compute_address" "internal-addresses" {
  count  = length(local.node_names)
  name   = local.node_names[count.index]
  region = element(var.regions, count.index)
}

resource "google_compute_instance" "nodes" {
  count                     = length(local.node_names)
  name                      = local.node_names[count.index]
  machine_type              = "e2-standard-4"
  zone                      = "${element(var.regions, count.index)}-a"
  tags                      = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin            = "TRUE"
    gce-container-declaration = module.images[count.index].metadata_value
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.internal-addresses[count.index].address
    }
  }

  service_account {
    scopes = [
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/trace.append"
    ]
  }
}

# Newtork Managed Zone: e.g. testnet.stridenet.co
resource "google_dns_managed_zone" "stridenet-network-sub-zone" {
  name     = "${var.deployment_name}-stridenet"
  dns_name = "${var.deployment_name}.stridenet.co."
}

# Sub-Zone NS Record in Parent Zone: e.g testnet.stridenet.co IN stridenet.co
resource "google_dns_record_set" "stridenet-sub-zone-name-service-in-parent" {
  name = google_dns_managed_zone.stridenet-network-sub-zone.dns_name
  type = "NS"
  ttl  = 300

  managed_zone = "stridenet"

  rrdatas = [
    "ns-cloud-a1.googledomains.com.", "ns-cloud-a2.googledomains.com.", "ns-cloud-a3.googledomains.com.", "ns-cloud-a4.googledomains.com."
  ]
}

# Type SOA (Start of Authority) Record for Managed Zone: e.g testnet.stridenet.co
resource "google_dns_record_set" "stridenet-sub-zone-name-service" {
  name = google_dns_managed_zone.stridenet-network-sub-zone.dns_name
  type = "SOA"
  ttl  = 21600

  managed_zone = google_dns_managed_zone.stridenet-network-sub-zone.name

  rrdatas = [
    "ns-cloud-a1.googledomains.com. cloud-dns-hostmaster.google.com. 1 21600 3600 259200 300"
  ]
}

# Type NS (Name Service) Record for Managed Zone: e.g testnet.stridenet.co
resource "google_dns_record_set" "stridenet-sub-zone-start-of-authority" {
  name = google_dns_managed_zone.stridenet-network-sub-zone.dns_name
  type = "NS"
  ttl  = 300

  managed_zone = google_dns_managed_zone.stridenet-network-sub-zone.name

  rrdatas = [
    "ns-cloud-a1.googledomains.com.", "ns-cloud-a2.googledomains.com.", "ns-cloud-a3.googledomains.com.", "ns-cloud-a4.googledomains.com."
  ]
}

# Type A (Static Hostname) for each node in the network: e.g. stride-node1.testnet.stridenet.co
resource "google_dns_record_set" "external-addresses" {
  count = length(local.node_names)
  name  = "${local.node_names[count.index]}.${google_dns_managed_zone.stridenet-network-sub-zone.dns_name}"
  type  = "A"
  ttl   = 300

  managed_zone = google_dns_managed_zone.stridenet-network-sub-zone.name

  rrdatas = [google_compute_instance.nodes[count.index].network_interface[0].access_config[0].nat_ip]
}


