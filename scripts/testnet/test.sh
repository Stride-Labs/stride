
docker pull stridelabs/internal:droplet_node1

docker tag stridelabs/internal:droplet_node1 gcr.io/stride-nodes/testnet:droplet_node1

docker push gcr.io/stride-nodes/testnet:droplet_node1

terraform apply