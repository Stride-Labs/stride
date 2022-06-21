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
  zone    = "us-central1-a"
}

data "google_compute_default_service_account" "default" {
}
resource "google_compute_address" "node1-internal" {
  name   = "node1-internal"
  region = "us-central1"
}
resource "google_compute_address" "node2-internal" {
  name = "node2-internal"
  region = "europe-west6"
}
resource "google_compute_address" "node3-internal" {
  name   = "node3-internal"
  region = "us-east4"
}
resource "google_compute_address" "seed-internal" {
  name   = "seed-internal"
  region = "us-west1"
}
resource "google_compute_address" "gaia-internal" {
  name   = "gaia-internal"
  region = "us-west1"
}
resource "google_compute_address" "hermes-internal" {
  name   = "hermes-internal"
  region = "us-west1"
}

resource "google_compute_address" "icq-internal" {
  name   = "icq-internal"
  region = "us-west1"
}

resource "google_compute_instance" "internal-node1" {
  name         = "internal-node1"
  machine_type = "e2-standard-2"
  zone         = "us-central1-a"
  tags         = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:internal_node1'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.node1-internal.address
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

resource "google_compute_instance" "internal-node2" {
  name         = "internal-node2"
  machine_type = "e2-standard-2"
  zone         = "europe-west6-a"
  tags         = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:internal_node2'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.node2-internal.address
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

resource "google_compute_instance" "internal-node3" {
  name         = "internal-node3"
  machine_type = "e2-standard-2"
  zone         = "us-east4-a"
  tags         = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:internal_node3'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.node3-internal.address
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

resource "google_compute_instance" "internal-seed" {
  name         = "internal-seed"
  machine_type = "e2-standard-2"
  zone         = "us-west1-a"
  tags         = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:internal_seed'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.seed-internal.address
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


resource "google_compute_instance" "internal-gaia" {
  name         = "internal-gaia"
  machine_type = "e2-standard-2"
  zone         = "us-west1-a"
  tags         = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:internal_gaia'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.gaia-internal.address
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

resource "google_compute_instance" "internal-hermes" {
  name         = "internal-hermes"
  machine_type = "e2-standard-2"
  zone         = "us-west1-c"
  tags         = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:internal_hermes'\n      stdin: false\n      tty: false\n  restartPolicy: Always\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }

  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.hermes-internal.address
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

resource "google_compute_instance" "internal-icq" {
  name         = "internal-icq"
  machine_type = "e2-standard-2"
  zone         = "us-west1-c"
  tags         = ["ssh"]
  allow_stopping_for_update = true

  metadata = {
    enable-oslogin = "TRUE"
    gce-container-declaration = "spec:\n  containers:\n    - name: node\n      image: 'gcr.io/stride-nodes/testnet:internal_icq'\n      stdin: false\n      tty: false\n  restartPolicy: Never\n"
  }
  boot_disk {
    initialize_params {
      image = "cos-cloud/cos-97-lts"
    }
  }
  
  network_interface {
    network = "default"
    access_config {
      nat_ip = google_compute_address.icq-internal.address
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
