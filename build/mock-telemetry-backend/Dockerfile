# Multi-stage build
# Stage 1: Compilação

FROM golang:1.24-alpine AS builder

WORKDIR /build


# Cache das dependências
COPY go.mod ./

RUN go mod download


# Código fonte
COPY pkg/ ./pkg/
COPY cmd/mock-telemetry-backend/ ./cmd/mock-telemetry-backend/


# Build binário estático
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -a \
    -ldflags="-w -s" \
    -o mock-telemetry-backend \
    ./cmd/mock-telemetry-backend/


# Stage 2: Runtime distroless

FROM gcr.io/distroless/static:nonroot


WORKDIR /


COPY --from=builder \
    /build/mock-telemetry-backend \
    /mock-telemetry-backend


USER nonroot:nonroot


EXPOSE 8080


ENTRYPOINT ["/mock-telemetry-backend"]
