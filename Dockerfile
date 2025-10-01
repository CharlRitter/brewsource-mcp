# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/brewsource-mcp ./app/cmd/server
RUN chmod +x /app/brewsource-mcp

# Runtime stage
FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /app/brewsource-mcp ./brewsource-mcp
COPY app/data/ ./data/
USER nonroot:nonroot
EXPOSE 8080
CMD ["./brewsource-mcp", "-mode=websocket", "-port=8080"]
