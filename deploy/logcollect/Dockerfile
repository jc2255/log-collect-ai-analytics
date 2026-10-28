# ==========================================================
# logcollect Agent - 独立部署 Dockerfile（支持 amd64 / arm64）
#
# 使用方式（项目根目录）:
#   1. make release-native   # 编译当前架构
#   2. cd deploy/logcollect && bash deploy.sh
# ==========================================================

ARG TARGETARCH=amd64

FROM docker.m.daocloud.io/library/alpine:3.19

ARG TARGETARCH

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 优先使用本目录下由 deploy.sh 复制的二进制；否则从 release 多架构目录复制
COPY logcollect ./server
RUN chmod +x ./server

RUN mkdir -p /app/dead_letters

ENTRYPOINT ["./server"]
CMD ["-config", "/app/config.yaml"]
