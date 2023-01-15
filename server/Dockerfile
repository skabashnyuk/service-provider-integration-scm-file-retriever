FROM golang:1.18 as builder
ENV GOPATH=/go/
USER root

WORKDIR /spi-file

# Copy the Go Modules manifests
COPY ./go.mod go.mod
COPY ./go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go sources
COPY ./server.go server.go
COPY ./websocket websocket
#COPY ../gitfile gitfile
COPY static/index.html static/index.html

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o spi-file server.go

# Compose the final image
FROM registry.access.redhat.com/ubi8/ubi-minimal:8.7-1049

# Install the 'shadow-utils' which contains `adduser` and `groupadd` binaries
RUN microdnf install shadow-utils \
	&& groupadd --gid 65532 nonroot \
	&& adduser \
		--no-create-home \
		--no-user-group \
		--uid 65532 \
		--gid 65532 \
		nonroot

COPY --from=builder /spi-file/spi-file /spi-file
COPY --from=builder /spi-file/static/index.html /static/index.html

WORKDIR /
USER 65532:65532

ENTRYPOINT ["/spi-file"]

