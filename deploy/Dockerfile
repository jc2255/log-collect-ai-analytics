# 使用预编译二进制，不依赖源码
# TARGETARCH 由 BuildKit / compose build-arg 注入（amd64 | arm64），不要写死默认 amd64
# 构建前: make ensure-binaries  或  bash scripts/ensure-binaries.sh

ARG APP_NAME=admin
ARG TARGETARCH

FROM docker.m.daocloud.io/library/alpine:3.19

ARG APP_NAME
ARG TARGETARCH

RUN apk add --no-cache ca-certificates tzdata curl \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 按构建目标架构复制二进制（release/bin/amd64|arm64/<app>）
COPY release/bin/${TARGETARCH}/${APP_NAME} ./server
RUN chmod +x ./server

EXPOSE 8080 8086 514

ENTRYPOINT ["./server"]
CMD ["-config", "/app/config.yaml"]
