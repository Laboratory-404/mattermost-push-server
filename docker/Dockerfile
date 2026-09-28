# Stage 1: Build binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/mattermost-push-proxy ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/test-push ./cmd/test-push

# Stage 2: Production scratch image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/mattermost-push-proxy /usr/local/bin/mattermost-push-proxy
COPY --from=builder /app/test-push /usr/local/bin/test-push
COPY --from=builder /app/config/config.example.yaml /etc/mattermost-push-proxy/config.yaml

EXPOSE 8066

ENTRYPOINT ["/usr/local/bin/mattermost-push-proxy"]
CMD ["-config", "/etc/mattermost-push-proxy/config.yaml"]
