# ---------- builder ----------
FROM golang:1.21-alpine AS builder
WORKDIR /src
COPY . .
RUN apk add --no-cache bash git && \
    go build -o /crawler cmd/main.go

# ---------- runtime ----------
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /crawler /app/crawler
COPY config.yaml .
ENTRYPOINT ["/app/crawler"]
