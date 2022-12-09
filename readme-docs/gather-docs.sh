#!/bin/bash
for f in $(find "./x" \
                -type f \
                -name "*.md" )
                # -not -path "scripts/*" \
                # -not -path "readme-docs/*"\
        # )
do
        echo $f
        start=$(echo $f | cut -d'/' -f3- | cut -d/ -f1)
        filename=$(echo $f | cut -d'/' -f3- | cut -d/ -f2)
        newfilename=$start"_"$filename
        cp $f readme-docs/md/$newfilename
done