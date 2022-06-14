# Stride Node Setup

This folder goes from an existing Stride image and will construct 
images specifically for Stride nodes.

## High Level Path

High level, we need to accomplish
1. Pull Stride image output from regular `make init`
2. Run local commands to create initial state for 4 nodes. 3 full nodes and 1 seed.
4. Push all of these containers out to Docker Registry.
5. Spin up GCP images using those images, kick them off (through Terraform)

