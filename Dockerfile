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

WORKDIR /app
ENV GOPATH=/app
RUN go install github.com/DarthSim/overmind/v2@latest
RUN go install -tags extended github.com/gohugoio/hugo@latest

COPY Procfile .
COPY --from=builder /app/host host

COPY nginx.conf /etc/nginx/nginx.conf

# Ref https://fly.io/docs/app-guides/multiple-processes/#use-a-procfile-manager
ENTRYPOINT ["/app/bin/overmind", "start", "-N"]