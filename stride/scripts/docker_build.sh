
btype=cached
while getopts f flag; do
    case "${flag}" in
        f) btype="full"
    esac
done

echo $btype

if [ "$btype" == "cached" ]; then
    docker build --tag stridezone:stride -f Dockerfile.stride .
    docker build --tag stridezone:gaia -f Dockerfile.gaia .
elif [ "$btype" == "full" ]; then
    docker build --no-cache --pull --tag stridezone:stride -f Dockerfile.stride .
    docker build --no-cache --pull --tag stridezone:gaia -f Dockerfile.gaia .
else
    echo "No docker build."
fi

