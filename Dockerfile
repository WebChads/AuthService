FROM golang:1.24.3-bookworm AS builder

# Build dependencies
RUN apt-get update && \
    apt-get install -y \
    librdkafka-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /build
COPY . .

# Building with dynamic linking
RUN CGO_ENABLED=1 GOOS=linux go build -o auth_service

# Final stage (with running binary app)
FROM debian:bookworm-slim

# Runtime dependencies
RUN apt-get update && \
    apt-get install -y \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /build/auth_service .
COPY --from=builder /build/configs ./configs

CMD ["./auth_service"]