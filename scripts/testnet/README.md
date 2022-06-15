# Stride Node Setup

This folder goes from an existing Stride image (from `make init`) and will construct 4 properly formatted docker images to seed the Stride testnet (3 validators, 1 seed). These images get launched on GCP through Terraform.

## High Level Path

High level, we need to accomplish
1. We set up the right backend state for the 3 validators and 1 seed node. Do this by `sh setup_testnet_state.sh`
2. We commit and push to GitHub to kick off image builds for all 4 nodes. This MUST occur after building backend state.
4. GitHub Actions will automatically upload the finished images to GCP Container Registry and Docker Hub.
5. User must Spin up GCP images using those images, kick them off (through `terraform apply`)

## Pending TODO

Right now, this approach relies on the full state being pushed up to GitHub (so that GitHub actions can copy it over to the requisite image).

We should migrate to a solution where this isn't the case. Maybe: 
1. Github Actions creates a Stride "base image" on push to `main`.
2. In `setup_testnet_state.sh` we create new images with the additional state added.
3. Github Actions creates our 4 node images on push to `main-droplet`

We could also make step (2) above fully work on Github actions, so the full deployment pipeline happens automatically.