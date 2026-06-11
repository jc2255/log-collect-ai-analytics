# ==========================================================
# logcollect Agent - 独立部署 Dockerfile
# 
# 使用方式（项目根目录）:
#   1. make release                         # 交叉编译 Linux 二进制
#   2. cd deploy/logcollect && bash deploy.sh   # 一键部署
#
# 或直接 docker compose:
#   cd deploy/logcollect
#   docker compose up -d --build
# ==========================================================

FROM docker.m.daocloud.io/library/alpine:3.19

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 复制当前目录下的预编译二进制（deploy.sh 会自动复制到此处）
COPY logcollect ./server
RUN chmod +x ./server

# 创建死信目录
RUN mkdir -p /app/dead_letters

ENTRYPOINT ["./server"]
CMD ["-config", "/app/config.yaml"]
