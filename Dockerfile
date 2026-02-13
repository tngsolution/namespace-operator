# =========================
# 1️⃣ Build stage
# =========================
FROM golang:1.24.0-alpine AS builder

# Sécurité de base
RUN apk add --no-cache ca-certificates git

WORKDIR /workspace

# Copier les fichiers Go
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY api/ api/
COPY controllers/ controllers/

# Build du binaire
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -a -o manager main.go


# =========================
# 2️⃣ Runtime stage
# =========================
FROM gcr.io/distroless/static:nonroot

WORKDIR /

# Copier le binaire
COPY --from=builder /workspace/manager /manager

# Utilisateur non-root (sécurité)
USER 65532:65532

ENTRYPOINT ["/manager"]
