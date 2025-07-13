# Build stage
FROM golang:1.23 as builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o redash-mcp-server main.go

# Runtime stage
FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/redash-mcp-server /app/redash-mcp-server
ENTRYPOINT ["/app/redash-mcp-server"]
