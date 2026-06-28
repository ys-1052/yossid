# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.26-bookworm AS builder

WORKDIR /build

# Cache go modules before copying source
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .

# Build for Linux amd64 — binary must be named `bootstrap` for provided.al2023
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o bootstrap ./cmd/lambda/main.go

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM public.ecr.aws/lambda/provided:al2023

# Copy the compiled binary to task root
COPY --from=builder /build/bootstrap /var/task/bootstrap

# Set the entrypoint to the bootstrap binary
ENTRYPOINT ["/var/task/bootstrap"]
