# Both binaries are compiled once in the build stage and shipped by the
# runtime targets below:
#   docker build -f dockerfile -t requisition-system .                  # API (default)
#   docker build -f dockerfile --target worker -t requisition-system-worker .
FROM golang:alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o app ./cmd/api
RUN go build -o worker ./cmd/worker

# API image — the default target, so a plain `docker build` keeps behaving
# exactly as before.
FROM golang:alpine AS api
LABEL maintainer="appeimisic@gmail.com"

RUN apk add --no-cache curl

WORKDIR /src
COPY --from=build /src/app ./app
EXPOSE 8080
ENTRYPOINT [ "./app" ]

# Worker image — drains the asynq task queues (invitation emails, ...).
FROM alpine AS worker
LABEL maintainer="appeimisic@gmail.com"

WORKDIR /src
COPY --from=build /src/worker ./worker
ENTRYPOINT [ "./worker" ]

FROM api
