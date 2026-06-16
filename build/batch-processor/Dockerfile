# Multi-stage build
# Stage 1: Build

FROM golang:1.24-alpine AS builder

WORKDIR /build


# Cache das dependências
COPY go.mod ./

RUN go mod download


# Código fonte
COPY pkg/ ./pkg/
COPY cmd/batch-processor/ ./cmd/batch-processor/


# Build estático
RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -a \
    -ldflags="-w -s" \
    -o batch-processor \
    ./cmd/batch-processor/


# Stage 2: Runtime distroless

FROM gcr.io/distroless/static:nonroot


WORKDIR /


COPY --from=builder \
    /build/batch-processor \
    /batch-processor


USER nonroot:nonroot


EXPOSE 8080


ENTRYPOINT ["/batch-processor"]
