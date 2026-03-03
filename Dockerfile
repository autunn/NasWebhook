# 阶段一：编译环境
FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY . .

ENV GOPROXY=https://proxy.golang.org,direct
RUN go mod tidy

# 编译为 nas-webhook
RUN CGO_ENABLED=0 GOOS=linux go build -o nas-webhook .

# 阶段二：运行环境
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/nas-webhook .
COPY templates ./templates

EXPOSE 5080
VOLUME ["/app/data"]

CMD ["./nas-webhook"]