FROM golang:1.21-alpine AS builder

# 接收编译时的版本号参数
ARG APP_VERSION=v2026.03.03

WORKDIR /app
RUN apk add --no-cache git
COPY . .

RUN go mod init nas-webhook || true
RUN go mod tidy

# 注入 APP_VERSION 到 main.Version 变量
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X 'main.Version=${APP_VERSION}'" \
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