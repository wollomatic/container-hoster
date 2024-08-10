FROM --platform=$BUILDPLATFORM golang:1.22.6-alpine3.20 AS build
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
