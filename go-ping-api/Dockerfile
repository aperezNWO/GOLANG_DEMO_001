# --- Build Stage ---
# Replace this line:
# FROM golang:1.22-alpine AS builder
# With this line:
FROM golang:1.25-alpine AS builder    
WORKDIR /app

COPY go.mod ./
# Download dependencies if go.sum exists
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# --- Minimal Runtime Stage ---
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
ENV PORT=8080

CMD ["./main"]