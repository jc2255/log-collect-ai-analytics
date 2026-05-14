# 多阶段构建 - Go 应用
# 通过 APP_NAME 参数指定要构建的应用

# 阶段1: 构建
FROM golang:1.21-alpine AS builder

ARG APP_NAME

RUN apk add --no-cache gcc musl-dev

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/${APP_NAME}

# 阶段2: 运行
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080 8086 514

ENTRYPOINT ["./server"]
CMD ["-config", "/app/config.yaml"]
