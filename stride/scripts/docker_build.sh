
btype=cached
while getopts f flag; do
    case "${flag}" in
        f) btype=full
    esac
done

if [ "$btype" == "cached" ]; then
    docker build --tag stridezone:stride -f Dockerfile.stride .
elif [ "$btype" == "full" ]; then
    docker build --no-cache --pull --tag stridezone:stride -f Dockerfile.stride .
else
    echo "No docker build."
fi
