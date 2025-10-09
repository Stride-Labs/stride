# #!/bin/bash

# NOTE: This script is run from inside the k8s pod

set -eu

if [ $# -ne 2 ]; then
    echo "Error: Both upgrade_name and upgrade_height are required"
    echo "Usage: $0 <upgrade_name> <upgrade_height>"
    exit 1
fi

upgrade_name=$1
upgrade_height=$2

cat > proposal.json << EOF
{
  "title": "Upgrade $upgrade_name",
  "summary": "Upgrade $upgrade_name",
  "metadata": "",
  "messages": [
    {
      "@type": "/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade",
      "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
      "plan": {
        "name": "$upgrade_name",
        "height": "$upgrade_height"
      }
    }
  ],
  "deposit": "2000000000ustrd"
}
EOF

strided tx gov submit-proposal proposal.json --from val1 -y 