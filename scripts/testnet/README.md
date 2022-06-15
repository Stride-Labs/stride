# Stride Node Setup

This folder goes from an existing Stride image (from `make init`) and will construct 4 properly formatted docker images to seed the Stride testnet (3 validators, 1 seed). These images get launched on GCP through Terraform.

## High Level Path

High level, we need to accomplish
1. We set up the right backend state for the 3 validators and 1 seed node. Do this by `sh setup_testnet_state.sh`
2. We commit and push to GitHub to kick off image builds for all 4 nodes. This MUST occur after building backend state.
4. GitHub Actions will automatically upload the finished images to GCP Container Registry and Docker Hub.
5. User must Spin up GCP images using those images, kick them off (through `terraform apply`)

