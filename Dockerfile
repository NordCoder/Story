FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w' -o /app/bin/feed-api cmd/story/main.go

FROM alpine:3.18 AS runner
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /app/bin/feed-api /app/feed-api
COPY config/ config/

ARG HTTP_PORT=8080
EXPOSE ${HTTP_PORT}


ENTRYPOINT ["/app/feed-api"]