# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o crawler666 .

FROM alpine:latest
RUN apk --no-cache add ca-certificates chromium
WORKDIR /root/

COPY --from=builder /app/crawler666 .
COPY --from=builder /app/config ./config/
COPY --from=builder /app/web ./web/

ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/lib/chromium/

EXPOSE 8080

CMD ["./crawler666"]
