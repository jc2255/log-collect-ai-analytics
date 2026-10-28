# LCA Elasticsearch 镜像：官方 7.17.25 + 阿里云 repository-oss（OSS 快照备份）
# 插件包由 scripts/prepare-oss-plugin.sh 生成到 deploy/es/
#
# 构建（在仓库根目录）:
#   docker compose -f docker-compose.ha.yaml build es01
# 或:
#   docker build -f deploy/Dockerfile.es -t lca-elasticsearch:7.17.25 .
ARG ES_VERSION=7.17.25
FROM docker.elastic.co/elasticsearch/elasticsearch:${ES_VERSION}

ARG ES_VERSION=7.17.25
USER root
COPY deploy/es/elasticsearch-repository-oss-${ES_VERSION}.zip /tmp/oss-plugin.zip
RUN elasticsearch-plugin install --batch file:///tmp/oss-plugin.zip \
    && rm -f /tmp/oss-plugin.zip \
    && elasticsearch-plugin list | grep -q elasticsearch-repository-oss
USER elasticsearch
