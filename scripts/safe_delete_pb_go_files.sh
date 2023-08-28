#!/bin/bash

# Find all .pb.go files and loop through them
find . -name '*.pb.go' | while read -r file; do
    # Check if the file's path contains "/migrations/"
    if [[ $file != *"/migrations/"* ]]; then
        # Delete the file if it's not in a migrations directory
        echo "Deleting $file"
        rm "$file"
    else
        echo "Skipping $file"
    fi
done
