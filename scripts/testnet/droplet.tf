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

data "google_compute_default_service_account" "default" {
}
resource "google_compute_address" "node1" {
  name   = "node1"
  region = "us-central1"
}
resource "google_compute_address" "node2" {
  name   = "node2"
  region = "europe-west6"
}
resource "google_compute_address" "node3" {
  name   = "node3"
  region = "us-east4"
}
resource "google_compute_address" "seed" {
  name   = "seed"
  region = "us-west1"
}
resource "google_compute_instance" "droplet-node1" {
  name                      = "droplet-node1"
  machine_type              = "e2-standard-4"
  zone                      = "us-central1-b"
  tags                      = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin            = "TRUE"
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
    scopes = ["https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/service.management.readonly",
    "https://www.googleapis.com/auth/trace.append"]
  }
}

resource "google_compute_instance" "droplet-node2" {
  name                      = "droplet-node2"
  machine_type              = "e2-standard-4"
  zone                      = "europe-west6-b"
  tags                      = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin            = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:droplet_node2'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.node2.address
    }
  }

  service_account {
    scopes = ["https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/service.management.readonly",
    "https://www.googleapis.com/auth/trace.append"]
  }
}

resource "google_compute_instance" "droplet-node3" {
  name                      = "droplet-node3"
  machine_type              = "e2-standard-4"
  zone                      = "us-east4-b"
  tags                      = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin            = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:droplet_node3'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.node3.address
    }
  }

  service_account {
    scopes = ["https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/service.management.readonly",
    "https://www.googleapis.com/auth/trace.append"]
  }
}

resource "google_compute_instance" "droplet-seed" {
  name                      = "droplet-seed"
  machine_type              = "e2-standard-4"
  zone                      = "us-west1-b"
  tags                      = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin            = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:droplet_seed'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.seed.address
    }
  }

  service_account {
    scopes = ["https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/service.management.readonly",
    "https://www.googleapis.com/auth/trace.append"]
  }
}


variable "regions" {
  type    = list(string)
  default = ["us-central1"]
}
variable "deployment_name" {
  type    = string
  default = "testnet"
}
variable "chain_name" {
  type    = string
  default = "stride"
}
resource "google_compute_address" "node-address" {
  name   = "${var.chain_name}-node1"
  region = regions[0]
}
resource "google_compute_instance" "test-nodes" {
  name                      = "${var.chain_name}-node1"
  machine_type              = "e2-standard-4"
  zone                      = regions[0]
  tags                      = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin            = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/${var.deployment_name}:${var.chain_name}-node1'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.test-node1.address
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
