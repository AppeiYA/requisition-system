# BUILD
FROM golang:alpine AS build

ARG VERSION=dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build \
    -ldflags "-X github.com/AppeiYA/requisition-system/internal/version.Version=${VERSION}" \
    -o app ./cmd/api

RUN go build \
    -ldflags "-X github.com/AppeiYA/requisition-system/internal/version.Version=${VERSION}" \
    -o worker ./cmd/worker


# WORKER
FROM alpine:3.24.4 AS worker

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=build --chown=app:app /src/worker ./worker

USER app

ENTRYPOINT ["./worker"]


# API
FROM alpine:3.24.4 AS api

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=build --chown=app:app /src/app ./app

USER app

EXPOSE 8080

HEALTHCHECK --interval=30s \
    --timeout=5s \
    --retries=3 \
    CMD wget -q -O /dev/null http://localhost:8080/healthz || exit 1

ENTRYPOINT ["./app"]