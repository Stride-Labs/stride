for f in $(find "../x" -type f -name "*.md" -not -path "../deps/*" \
            -not -path "../scripts/*" \
            -not -path "../readme-docs/*"
        )
do
        echo $f
        cp $f md
done