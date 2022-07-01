# Stride Node Setup

This folder contains the scripts necessary to launch a testnet on GCP. The main execution is handled by github actions, which builds the docker images for Stride, Gaia, Hermes and ICQ. Once the images are built, the network can be stood up with terraform (through `terraform apply`)

## Spinning Up
* Set the deployment name (`deployment_name`) and number of desired stride validator nodes (`num_stride_nodes`) in the github actions workflow (`testnet.yml`) and the terraform script (`main.tf`)
* The workflow must be triggered manually and it will:
    1. Compile Stride and Gaia Binaries
    2. Initialize Stride and Gaia state
    3. Create startup scripts for Hermes and ICQ
    4. Build Docker images for Stride, Gaia, Hermes and ICQ
* Once the images are built, all the resources can be deployed by running `terraform apply`. It will stand up:
    1. Stride, Gaia, Hermes, and ICQ Nodes
    2. Static Internal IP Addresses
    3. A DNS managed zone of the form `{deployment_name}.stridenet.co` (named `{deployment_name}-stridenet`)
    4. Endpoints for each node (e.g. `stride-node1.{deployment_name}.stridenet.co`)
    5. A DNS Name Service (NS) Record for `{deployment_name}.stridenet.co` in the `stridenet` managed zone
    6. A SOA Record and NS Record in the `{deployment_name}.stridenet.co` managed zone
        * The reason for this is to keep the name servers consistent 
* The DNS setup is hit or miss. To check if there's an error (and restart if there is), run `bash fix_dns.sh`
## Shutting Down
* Run `terraform destroy` to remove each resource.
* Terraform has trouble removing the DNS resources. To get around this, you'll have to manually delete the DNS managed zone (named `{deployment_name}-stridenet`) from GCP. This is best run after the above command so that the record sets in the managed zone are already deleted.
## Pending TODO
1. Create base images for each service that contains just the executable. That way, the image building step will simply have to copy the new state files in.
2. Link the terraform step in Github Actions so that it creates our nodes on push 


