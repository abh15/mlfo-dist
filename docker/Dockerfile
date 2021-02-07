FROM golang:1.15-alpine AS build_base

RUN apk add --update --no-cache git bash tcpdump iperf busybox-extras iproute2 iputils

# Set the Current Working Directory inside the container
WORKDIR /tmp/mlfo

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -o ./out/mlfo .

# Start fresh from a smaller image
FROM alpine:latest

RUN apk add ca-certificates
RUN apk add --update --no-cache bash tcpdump iperf busybox-extras iproute2 iputils

# ENV GRPC_GO_LOG_VERBOSITY_LEVEL 99
# ENV GRPC_GO_LOG_SEVERITY_LEVEL info

COPY --from=build_base /tmp/mlfo/out/mlfo /app/mlfo

ENTRYPOINT ["/app/mlfo"]


