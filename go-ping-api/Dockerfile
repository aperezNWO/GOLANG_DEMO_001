# Stage 1: Build binary
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache protobuf-dev protoc git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate Protobuf Go stubs
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/fractal.proto

RUN CGO_ENABLED=0 GOOS=linux go build -o server .

# Stage 2: Minimal Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/server .

EXPOSE 8080
ENV PORT=8080

CMD ["./server"]