# 可选：在官方 ES 镜像上安装阿里云 repository-oss 插件（OSS 快照备份用）
# 默认 HA/单机 compose 直接使用官方镜像，不走本 Dockerfile。
#
# 用法：
#   1. 从 https://github.com/aliyun/elasticsearch-repository-oss/releases
#      下载与 ES 版本一致的插件包到 deploy/es/
#   2. 修改下方 COPY 文件名后：
#      docker compose -f docker-compose.ha.yaml build es01
FROM docker.elastic.co/elasticsearch/elasticsearch:7.17.25

# COPY deploy/es/elasticsearch-repository-oss-<ver>.zip /tmp/oss-plugin.zip
# RUN elasticsearch-plugin install --batch file:///tmp/oss-plugin.zip \
#     && rm -f /tmp/oss-plugin.zip
