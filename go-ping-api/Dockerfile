# Stage 1: Build
FROM golang:alpine AS builder

# Enable automatic toolchain downloads if go.mod requires a newer Go patch version
ENV GOTOOLCHAIN=auto

RUN apk add --no-cache protobuf-dev protoc git

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .

# Generate Protobuf stubs inside container
RUN mkdir -p pkg/grpc/pb && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    protoc --go_out=pkg/grpc/pb --go_opt=paths=source_relative \
           --go-grpc_out=pkg/grpc/pb --go-grpc_opt=paths=source_relative \
           proto/fractal.proto

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app .

# Stage 2: Runtime Minimal Container
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/app .

# Expose HTTP REST / gRPC-Web (8080) and Native gRPC (50051)
EXPOSE 8080
EXPOSE 50051

CMD ["./app"]