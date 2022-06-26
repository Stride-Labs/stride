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
* The DNS setup is hit or miss and sometimes requires manual intervention. 
    * If it works properly, you should be able to successfully ping the endpoint (`ping stride-node1.{deployment_name}.stridenet.co`)
    * If that doesn't work, go to the Cloud DNS section and do the following:
        * Click on the managed zone that was created (named `{deployment_name}-stridenet`)
        * Click the edit button on the Type NS Record. There should be a warning message at the bottom that says the name servers might not have been configured properly, and there will be an option to restore them to defaults. Click this option.
        * Identify the letter ("a" through "e") that indicates the grouping of name servers (e.g. ns-cloud-e1.googledomains.com. => "e")
        * Then make all name servers consistent by replacing the "a" in each name server with this new letter (e.g. ns-cloud-a1.googledomains.com. should be changed to ns-cloud-e1.googledomains.com.). This should be done for:
            * The Type SOA record named `{deployment_name}.stridenet.co` in the `{deployment_name}-stridenet` managed zone
            * The Type NS Record named `{deployment_name}.stridenet.co` in the `stridenet` managed zone
## Shutting Down
* Terraform has trouble removing the DNS resources. To get around this, first manually delete the managed zone that was created (named `{deployment_name}-stridenet`) by deleting all records sets in the zone and then deleting the zone itself.
* Then run `terraform destroy` to remove the remaing resources
## Pending TODO
1. Create base images for each service that contains just the executable. That way, the image building step will simply have to copy the new state files in.
2. Link the terraform step in Github Actions so that it creates our nodes on push 


