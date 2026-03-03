FROM golang:1.21-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY . .
RUN go mod init nas-webhook || true
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o nas-webhook-app .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/nas-webhook-app .
COPY templates ./templates

# 必须复制图片到运行阶段
COPY logo.png ./

EXPOSE 5080
VOLUME ["/app/data"]
CMD ["./nas-webhook-app"]