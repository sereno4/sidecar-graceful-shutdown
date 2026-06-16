# Multi-stage build
# Stage 1: Compilação

FROM golang:1.24-alpine AS builder

WORKDIR /build


# Cache das dependências
COPY go.mod ./

RUN go mod download


# Código fonte
COPY pkg/ ./pkg/
COPY cmd/telemetry-sidecar/ ./cmd/telemetry-sidecar/


# Build binário estático
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -a \
    -ldflags="-w -s" \
    -o telemetry-sidecar \
    ./cmd/telemetry-sidecar/


# Stage 2: Runtime distroless

FROM gcr.io/distroless/static:nonroot


WORKDIR /


COPY --from=builder \
    /build/telemetry-sidecar \
    /telemetry-sidecar


USER nonroot:nonroot


EXPOSE 8081


ENTRYPOINT ["/telemetry-sidecar"]
