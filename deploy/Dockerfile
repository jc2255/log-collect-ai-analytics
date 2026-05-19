# 使用预编译二进制，不依赖源码
# 通过 APP_NAME 参数指定要运行的服务（admin / apiserver / logtransfer / syslog）

ARG APP_NAME=admin

FROM alpine:3.19

# 重新声明 ARG（FROM 之后 ARG 作用域重置）
ARG APP_NAME

RUN apk add --no-cache ca-certificates tzdata curl \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

# 从预编译产物目录复制对应服务的二进制
COPY release/bin/${APP_NAME} ./server
RUN chmod +x ./server

EXPOSE 8080 8086 514

ENTRYPOINT ["./server"]
CMD ["-config", "/app/config.yaml"]
