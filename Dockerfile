# syntax = docker/dockerfile:1

ARG GO_VERSION="1.21"
ARG RUNNER_IMAGE_VERSION="3.17"

FROM golang:${GO_VERSION}-alpine${RUNNER_IMAGE_VERSION} as builder

WORKDIR /opt
RUN apk add --no-cache make git gcc musl-dev openssl-dev linux-headers ca-certificates build-base

COPY go.mod .
COPY go.sum .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

RUN WASMVM_VERSION=$(cat go.mod | grep github.com/CosmWasm/wasmvm | awk '{print $2}') \
    && wget https://github.com/CosmWasm/wasmvm/releases/download/$WASMVM_VERSION/libwasmvm_muslc.$(uname -m).a \
    -O /lib/libwasmvm_muslc.a

COPY . .

RUN BUILD_TAGS=muslc LINK_STATICALLY=true make build

# Add to a distroless container
FROM alpine:${RUNNER_IMAGE_VERSION}

COPY --from=builder /opt/build/strided /usr/local/bin/strided
RUN apk add bash vim sudo dasel jq \
    && addgroup -g 1000 stride \
    && adduser -S -h /home/stride -D stride -u 1000 -G stride 

RUN mkdir -p /etc/sudoers.d \
    && echo '%wheel ALL=(ALL) ALL' > /etc/sudoers.d/wheel \
    && echo "%wheel ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers \
    && adduser stride wheel 

USER 1000
ENV HOME /home/stride
WORKDIR $HOME

EXPOSE 26657 26656 1317 9090

CMD ["strided", "start"]