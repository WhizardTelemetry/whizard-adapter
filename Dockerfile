FROM golang:1.19.1 as builder
ARG GOPROXY

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY apis/ apis
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY vendor/ vendor

# Build -mod=vendor
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -o adapter cmd/adapter.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/adapter .
USER 65532:65532

ENTRYPOINT ["/adapter"]
