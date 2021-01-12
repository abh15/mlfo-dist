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
FROM alpine:3.9 
RUN apk add ca-certificates
RUN apk add --update --no-cache git bash tcpdump iperf busybox-extras iproute2 iputils


COPY --from=build_base /tmp/mlfo/out/mlfo /app/mlfo

# This container exposes port 8080 to the outside world
# EXPOSE 8000
# EXPOSE 9000

# COPY actuate /usr/local/bin/
# RUN chmod +x /usr/local/bin/actuate

# Run the binary program produced by `go install`
# CMD ["actuate&"]
# CMD ["/bin/bash"]
CMD ["/app/mlfo"]

