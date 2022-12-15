# syntax = docker/dockerfile:1

ARG GO_VERSION="1.19"
ARG RUNNER_IMAGE="alpine:3.16"

FROM golang:${GO_VERSION}-alpine as builder-deps

WORKDIR /opt
RUN apk add --no-cache make git gcc musl-dev openssl-dev linux-headers

COPY go.mod .
COPY go.sum .

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/go/pkg/mod \
    go mod download

# Copy the remaining files
COPY . .

FROM builder-deps as production-builder

RUN LINK_STATICALLY=true make build

FROM builder-deps as debug-builder

RUN LINK_STATICALLY=true DOCKER_DEBUG=true make build
RUN CGO_ENABLED=0 go install -ldflags "-s -w -extldflags '-static'" -v github.com/go-delve/delve/cmd/dlv@latest

# Add to a distroless container
FROM ${RUNNER_IMAGE} as stride

RUN apk add bash vim sudo dasel \
    && addgroup -g 1000 stride \
    && adduser -S -h /home/stride -D stride -u 1000 -G stride 

RUN mkdir -p /etc/sudoers.d \
    && echo '%wheel ALL=(ALL) ALL' > /etc/sudoers.d/wheel \
    && echo "%wheel ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers \
    && adduser stride wheel 

USER 1000
ENV HOME /home/stride
WORKDIR $HOME

FROM stride as debug

COPY --from=debug-builder /opt/build/strided /usr/local/bin/strided
COPY --from=debug-builder /go/bin/dlv /usr/local/bin/dlv

EXPOSE 26657 26656 1317 9090 2345

CMD ["strided", "start"]

FROM stride as production

COPY --from=production-builder /opt/build/strided /usr/local/bin/strided

EXPOSE 26657 26656 1317 9090

CMD ["strided", "start"]
