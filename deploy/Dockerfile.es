FROM ccr.ccs.tencentyun.com/lcalog/es:7.10

# 离线安装阿里云官方 repository-oss 插件（type=oss，AK/SK 直接放 settings，无需 keystore）
# 插件包必须与 ES 版本严格一致；下载地址：
#   https://github.com/aliyun/elasticsearch-repository-oss/releases
COPY deploy/es/elasticsearch-repository-oss-7.10.0.zip /tmp/oss-plugin.zip
RUN elasticsearch-plugin install --batch file:///tmp/oss-plugin.zip \
    && rm -f /tmp/oss-plugin.zip
