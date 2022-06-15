sh setup_testnet_images.sh
sh push_to_gcr.sh
terraform apply -replace=google_compute_instance.droplet-node1
# docker run -it gcr.io/stride-nodes/testnet:droplet_node1
