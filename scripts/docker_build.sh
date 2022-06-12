
btype=cached
while getopts fsa flag; do
    case "${flag}" in
        f) btype="full" ;;
        s) btype="stride" ;;
        a) btype="strideall" ;;
    esac
done

echo $btype

if [ "$btype" == "cached" ]; then
    docker build --tag stridezone:stride -f Dockerfile.stride .
    docker build --tag stridezone:gaia -f Dockerfile.gaia .
    docker build --tag stridezone:hermes -f Dockerfile.hermes .
    docker-compose build icq 
elif [ "$btype" == "full" ]; then
    docker build --no-cache --pull --tag stridezone:stride -f Dockerfile.stride .
    docker build --no-cache --pull --tag stridezone:gaia -f Dockerfile.gaia .
    docker build --pull --tag stridezone:hermes -f Dockerfile.hermes .
    docker-compose build icq --no-cache --pull
elif [ "$btype" == "stride" ]; then
    docker build --tag stridezone:stride -f Dockerfile.stride .
elif [ "$btype" == "strideall" ]; then
    docker build --no-cache --pull --tag stridezone:stride -f Dockerfile.stride .
else
    echo "No docker build."
fi

