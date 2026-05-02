# Stage 1: test + build
FROM golang:1.25 AS builder

WORKDIR /src

# Copy dependency files first. This cached layer only busts on go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY cmd ./cmd
COPY internal ./internal
COPY api ./api
COPY db ./db

# Build args for version stamping
ARG APP_VERSION=dev
ARG GIT_SHA=local
ARG BUILD_TIME=unknown

# Run tests before building. Build fails if tests fail.
RUN CGO_ENABLED=0 go test ./... -v -count=1

# Build static binary. Runtime version metadata is supplied through env vars.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w" \
    -o /out/dronefleet \
    ./cmd/server

# Stage 2: production image
# distroless = no shell, no package manager, minimal attack surface
# Same base you were using; keeping the small distroless runtime.
FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/dronefleet /app/dronefleet

# Build args repeated because this stage sets runtime ENV.
ARG APP_VERSION=dev
ARG GIT_SHA=local
ARG BUILD_TIME=unknown

# Default environment. All values are overridable at runtime via K8s ConfigMap/Secret.
ENV APP_NAME=dronefleet \
    APP_ENV=production \
    APP_PORT=8080 \
    LOG_LEVEL=info \
    FEATURE_CACHE_WARM=true \
    MAX_FLEET_SIZE=50 \
    TELEMETRY_BUFFER=100 \
    DATABASE_URL="" \
    APP_VERSION=${APP_VERSION} \
    GIT_SHA=${GIT_SHA} \
    BUILD_TIME=${BUILD_TIME}

EXPOSE 8080

# Distroless has no shell so HEALTHCHECK CMD won't work here.
# Liveness/readiness is handled by K8s probes hitting /health and /ready.

# Non-root user. Distroless provides uid 65532 (nonroot).
USER nonroot:nonroot

ENTRYPOINT ["/app/dronefleet"]
