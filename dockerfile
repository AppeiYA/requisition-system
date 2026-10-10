FROM golang:alpine
LABEL maintainer="appeimisic@gmail.com"

RUN apk add --no-cache curl

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app ./cmd/api
EXPOSE 8080
ENTRYPOINT [ "./app" ]