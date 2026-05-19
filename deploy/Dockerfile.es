FROM ccr.ccs.tencentyun.com/lcalog/es:7.10

# 安装 repository-s3 插件（阿里云 OSS 兼容 S3 协议）
RUN elasticsearch-plugin install --batch repository-s3
