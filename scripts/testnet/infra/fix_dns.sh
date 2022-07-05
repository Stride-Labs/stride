#!/bin/bash

set -eu

# This script replaces the managed DNS subzone's NS record in the parent zone
# E.g. replaces the testnet.stridenet.co NS server IN stridenet.co

deployment_name="$1"
parent_zone_name="stridenet"
sub_zone_name="${deployment_name}-${parent_zone_name}"
record_set_name="${deployment_name}.${parent_zone_name}.co"

# First figure out what region the managed zone is in by first grabbing the routing policy of the SOA record (e.g. ns-cloud-c1.googledomains.com.)
example_name_server=$(gcloud dns record-sets describe $record_set_name --type=SOA --zone=$sub_zone_name --format=text | grep rrdatas | awk '{print $2}')
region_prefix=${example_name_server:0:10} # e.g. ns-cloud-c

# Use the prefix to create the desired policy
desired_rrdatas="${region_prefix}1.googledomains.com.,${region_prefix}2.googledomains.com.,${region_prefix}3.googledomains.com.,${region_prefix}4.googledomains.com."

# Then get the current policy
current_rrdatas=$(gcloud dns record-sets describe $record_set_name --type=NS --zone=$parent_zone_name --format='value[no-heading](DATA)')

printf "%s\n\t%s\n" "Current Policy in Parent Zone" $current_rrdatas
printf "%s\n\t%s\n" "Desired Policy in Parent Zone" $desired_rrdatas


if [ "$current_rrdatas" == ${desired_rrdatas} ]; then
	echo "Policies match. No action necessary."
else
	# Update NS record in parent zone
	printf "\n%s\n" "Policies do not match. Updating policy."
	gcloud dns record-sets update $record_set_name --zone=$parent_zone_name --type=NS --ttl 300 --rrdatas=$desired_rrdatas

	# Get the list of machines to reset
	printf "\n%s\n" "Resetting machines in background threads..."
	gcloud compute instances list --format='value[no-heading](NAME,ZONE)' | grep $deployment_name | while read line; do
		node=$(echo $line | awk '{print $1}')
		zone=$(echo $line | awk '{print $2}')

		printf "\t%s\n" "Resetting $node in $zone"
		gcloud compute instances reset $node --zone $zone &
	done

	# Wait for background processes to finish
	while pgrep -f -q "google-cloud-sdk"; do
		sleep 2
	done

	echo "Done"
fi

