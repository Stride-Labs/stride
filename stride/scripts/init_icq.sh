#!/bin/bash

rm scripts/state/ICQ/config.yaml

# CACHE (no changes to interchain-queries GH repo)
# docker build --tag stridezone:interchain-queries -f Dockerfile.icq .
# NO CACHE (changes to interchain-queries GH repo)
docker build --no-cache --pull --tag stridezone:interchain-queries -f Dockerfile.icq .

# docker-compose build icq --no-cache

docker-compose run icq /bin/sh