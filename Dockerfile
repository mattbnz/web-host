FROM golang:1.19 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./
RUN go build -v -o host


FROM ubuntu:kinetic

# Big & slow, keep separate for good caching.
RUN apt-get update && apt-get install -y npm

# All the other packages we need
RUN apt-get update && \
    apt-get install -y \
        bash \
        tmux \
        curl \
        nginx \
        golang-go

RUN curl -1sLf 'https://repositories.timber.io/public/vector/gpg.3543DB2D0A2BC4B8.key' |  gpg --dearmor > /usr/share/keyrings/timber-vector-archive-keyring.gpg
RUN curl -1sLf 'https://repositories.timber.io/public/vector/config.deb.txt?distro=ubuntu&codename=kinetic' > /etc/apt/sources.list.d/timber-vector.list
RUN apt-get update &&  apt-get install -y vector

WORKDIR /app
ENV GOPATH=/app
RUN go install github.com/DarthSim/overmind/v2@latest
RUN go install -tags extended github.com/gohugoio/hugo@latest

COPY Procfile .
COPY --from=builder /app/host host

COPY nginx.conf /etc/nginx/nginx.conf
COPY vector.toml /etc/vector/

# Ref https://fly.io/docs/app-guides/multiple-processes/#use-a-procfile-manager
ENTRYPOINT ["/app/bin/overmind", "start", "-N"]