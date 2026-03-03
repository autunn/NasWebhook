FROM golang:1.21-alpine AS builder

# 接收构建参数
ARG APP_VERSION=v2026.03.03

WORKDIR /app
RUN apk add --no-cache git
COPY . .

RUN go mod init nas-webhook || true
RUN go mod tidy

# 修复点：去掉外层多余的双引号解析，确保 ldflags 作为一个整体被 Go 捕获
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.Version=${APP_VERSION}" \
    -o nas-webhook-app .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

COPY --from=builder /app/nas-webhook-app .
COPY templates ./templates
COPY logo.png ./

EXPOSE 5080
VOLUME ["/app/data"]
CMD ["./nas-webhook-app"]