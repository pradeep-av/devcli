# --- Build Stage ---
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy project source
COPY . .

# Compile the devcli binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o devcli main.go

# Compile grpcurl in the builder stage to ensure a clean static binary
RUN go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# --- Final Stage ---
FROM alpine:latest

# Install runtime utilities: bash, curl, jq, and root CA certificates for HTTPS requests
RUN apk add --no-cache bash curl jq ca-certificates

# Copy compiled binaries from builder stage
COPY --from=builder /app/devcli /usr/local/bin/devcli
COPY --from=builder /go/bin/grpcurl /usr/local/bin/grpcurl

# Create a default workspace directory
WORKDIR /workspace

# Set the default entrypoint to devcli.
# Developers can still run other utilities by overriding the entrypoint:
# e.g., docker run --entrypoint grpcurl <image> ...
ENTRYPOINT ["devcli"]
