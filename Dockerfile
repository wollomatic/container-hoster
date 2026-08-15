# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.26.6-alpine3.24@sha256:af8d6740070b8906d12eae1c3e3ea0957fb63f492051ea05e354c38ef9fe88df AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY *.go ./
ARG TARGETOS
ARG TARGETARCH
ARG VERSION=version_not_set
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build --tags netgo -ldflags="-w -s -X 'main.version=${VERSION}'" -o /container-hoster .

FROM scratch
LABEL org.opencontainers.image.source=https://github.com/wollomatic/container-hoster
LABEL org.opencontainers.image.description="A simple 'etc/hosts' file injection tool to resolve names of local Docker containers on the host."
LABEL org.opencontainers.image.licenses=MIT
LABEL securitytxt="https://wollomatic.de/.well-known/security.txt"
VOLUME /var/run/docker.sock
VOLUME /hosts
ENTRYPOINT ["/container-hoster"]
WORKDIR /
COPY --from=build ./container-hoster /container-hoster
