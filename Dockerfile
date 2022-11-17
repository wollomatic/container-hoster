FROM golang:1.19-bullseye AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go ./

RUN GOOS=linux GOARCH=amd64 go build --tags netgo -ldflags="-w -s" -o /container-hoster .


FROM scratch

LABEL org.opencontainers.image.source=https://github.com/wollomatic/container-hoster
LABEL org.opencontainers.image.description="A simple 'etc/hosts' file injection tool to resolve names of local Docker containers on the host."
LABEL org.opencontainers.image.licenses=MIT

WORKDIR /

COPY --from=build ./container-hoster /container-hoster
COPY ./README.md /README.md

VOLUME /var/run/docker.sock
VOLUME /hosts

ENTRYPOINT ["/container-hoster"]
