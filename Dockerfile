# Simple runtime-only Dockerfile
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy the pre-built binary
COPY app/bin/brewsource-mcp ./brewsource-mcp

# Copy BJCP data files
COPY app/data/ ./data/

# Make sure it's executable (though this should already be set)
USER nonroot:nonroot

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./brewsource-mcp", "-mode=websocket", "-port=8080"]
