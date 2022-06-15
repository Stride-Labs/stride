

# docker tag stridelabs/internal:droplet_node1 gcr.io/stride-nodes/testnet:droplet_node1
# docker tag stridelabs/internal:droplet_node2 gcr.io/stride-nodes/testnet:droplet_node2
# docker tag stridelabs/internal:droplet_node3 gcr.io/stride-nodes/testnet:droplet_node3
# docker tag stridelabs/internal:droplet_seed gcr.io/stride-nodes/testnet:droplet_seed

docker push gcr.io/stride-nodes/testnet:droplet_node1
docker push gcr.io/stride-nodes/testnet:droplet_node2
docker push gcr.io/stride-nodes/testnet:droplet_node3
docker push gcr.io/stride-nodes/testnet:droplet_seed