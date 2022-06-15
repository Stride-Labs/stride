# auth to GCP
terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
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

data "google_compute_default_service_account" "default" {
}

# define all of our IPs 
resource "google_compute_address" "node1" {
  name = "node1"
}
# define all of our IPs 
resource "google_compute_address" "node2" {
  name = "node2"
}
# define all of our IPs 
resource "google_compute_address" "node3" {
  name = "node3"
}
# define all of our IPs 
resource "google_compute_address" "seed" {
  name = "seed"
}

# Create a single Compute Engine instance
resource "google_compute_instance" "droplet-node1" {
  name         = "droplet-node1"
  machine_type = "e2-standard-4"
  zone         = "us-central1-c"
  tags         = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:droplet_node1'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.node1.address
    }
  }

  service_account {
    scopes = [ "https://www.googleapis.com/auth/devstorage.read_only",
        "https://www.googleapis.com/auth/logging.write",
        "https://www.googleapis.com/auth/monitoring.write",
        "https://www.googleapis.com/auth/servicecontrol",
        "https://www.googleapis.com/auth/service.management.readonly",
        "https://www.googleapis.com/auth/trace.append"]
  }
}