FROM golang:1.15-alpine AS build_base

RUN apk add --update --no-cache git bash curl tcpdump iperf busybox-extras iproute2 iputils

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

RUN apk add bash tcpdump iperf busybox-extras iproute2 iputils python3 python3-dev py3-pip g++ py3-psutil

# ENV GRPC_GO_LOG_VERBOSITY_LEVEL 99
# ENV GRPC_GO_LOG_SEVERITY_LEVEL info

COPY --from=build_base /tmp/mlfo/out/mlfo /app/mlfo
COPY build/prometheus_ss_exporter-1.0.0-py3-none-any.whl prometheus_ss_exporter-1.0.0-py3-none-any.whl
COPY ./monitoring/ss_exporter_src /app/ss_exporter_src
COPY ./monitoring/node_exporter /app/node_exporter
COPY build/entryexec.sh /app/entryexec.sh

RUN chmod +x /app/entryexec.sh
RUN pip install wheel pyroute2==0.5.10 PyYAML==5.1.1 prometheus-client 
# psutil==5.6.2
RUN python3 -m pip install --no-cache-dir 'prometheus_ss_exporter-1.0.0-py3-none-any.whl'



ENTRYPOINT ["bash", "/app/entryexec.sh"]
# ENTRYPOINT ["/app/mlfo"]


