.PHONY: all build clean admin apiserver logcollect logcollect_win logtransfer syslog \
	release release-native release-amd64 release-arm64 ensure-binaries ensure-oss-plugin \
	prepare-host docker-up docker-up-all docker-down docker-rm docker-switch-ha docker-infra docker-rebuild docker-rebuild-agent \
	docker-up-ha docker-logs docker-clean

BINARY_DIR=bin
GO=go
GOFLAGS=-v
GOARCH_NATIVE=$(shell bash scripts/goarch.sh)
export LCA_TARGETARCH=$(GOARCH_NATIVE)

all: build

build: admin apiserver logcollect logcollect_win logtransfer syslog

admin:
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/admin ./cmd/admin

apiserver:
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/apiserver ./cmd/apiserver

logcollect:
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/logcollect ./cmd/logcollect

logcollect_win:
	GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_DIR)/logcollect.exe ./cmd/logcollect_win

logtransfer:
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/logtransfer ./cmd/logtransfer

syslog:
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/syslog ./cmd/syslog

clean:
	rm -rf $(BINARY_DIR)

# 交叉编译双架构发布版本（amd64 + arm64）
release:
	bash scripts/release.sh

# 仅编译当前宿主机架构（Docker 本机构建前用）
release-native:
	bash scripts/release.sh --native

release-amd64:
	bash scripts/release.sh --arch amd64

release-arm64:
	bash scripts/release.sh --arch arm64

ensure-binaries:
	bash scripts/ensure-binaries.sh

# ES OSS 快照插件：交付包应自带 zip；源码仓缺失时自动下载整理
ensure-oss-plugin:
	@if [ -f deploy/es/elasticsearch-repository-oss-7.17.25.zip ]; then \
		echo "[ensure-oss-plugin] 已存在 deploy/es/elasticsearch-repository-oss-7.17.25.zip"; \
	elif [ -f scripts/prepare-oss-plugin.sh ]; then \
		bash scripts/prepare-oss-plugin.sh; \
	else \
		echo "缺少 OSS 插件包 deploy/es/elasticsearch-repository-oss-7.17.25.zip" >&2; exit 1; \
	fi

# 有源码时重新构建前端；交付包没有 web/src，直接使用 release/web-dist
ensure-web:
	@if [ -d web/src ]; then bash scripts/build.sh --web --no-win; \
	elif [ -f release/web-dist/index.html ]; then echo "[ensure-web] 使用已打包的 release/web-dist"; \
	else echo "缺少 release/web-dist，无法启动页面" >&2; exit 1; fi

# Linux：没有 Docker 则安装；vm.max_map_count 不足 262144 则写入。已满足则跳过。
prepare-host:
	bash scripts/prepare-host.sh

# ==================== Docker ====================
# 按本机架构编译二进制；通过 LCA_TARGETARCH + DOCKER_DEFAULT_PLATFORM 对齐镜像架构
# （当前 docker compose build 不支持 --platform，故用环境变量）
# 注意：页面改动在 web/，必须 ensure-web，否则 docker 里仍是旧前端

DOCKER_PLATFORM=linux/$(GOARCH_NATIVE)

docker-up: prepare-host ensure-binaries ensure-oss-plugin ensure-web
	LCA_TARGETARCH=$(GOARCH_NATIVE) DOCKER_DEFAULT_PLATFORM=$(DOCKER_PLATFORM) \
		docker compose build --build-arg TARGETARCH=$(GOARCH_NATIVE)
	LCA_TARGETARCH=$(GOARCH_NATIVE) DOCKER_DEFAULT_PLATFORM=$(DOCKER_PLATFORM) \
		docker compose up -d --force-recreate

# 应用服务列表（不含 logcollect；Agent 用 deploy/logcollect 独立部署）
COMPOSE_APPS=admin apiserver logtransfer syslog web

docker-up-all: prepare-host ensure-binaries ensure-web
	LCA_TARGETARCH=$(GOARCH_NATIVE) DOCKER_DEFAULT_PLATFORM=$(DOCKER_PLATFORM) \
		docker compose build --build-arg TARGETARCH=$(GOARCH_NATIVE) $(COMPOSE_APPS)
	LCA_TARGETARCH=$(GOARCH_NATIVE) DOCKER_DEFAULT_PLATFORM=$(DOCKER_PLATFORM) \
		docker compose up -d --force-recreate $(COMPOSE_APPS)

docker-up-ha: prepare-host ensure-binaries ensure-oss-plugin ensure-web
	LCA_TARGETARCH=$(GOARCH_NATIVE) DOCKER_DEFAULT_PLATFORM=$(DOCKER_PLATFORM) \
		docker compose -f docker-compose.ha.yaml build --build-arg TARGETARCH=$(GOARCH_NATIVE)
	# HA 本身不含 logcollect；前端由 nginx 挂载 ./release/web-dist
	LCA_TARGETARCH=$(GOARCH_NATIVE) DOCKER_DEFAULT_PLATFORM=$(DOCKER_PLATFORM) \
		docker compose -f docker-compose.ha.yaml up -d --force-recreate

docker-down:
	docker compose down

# 只删除单机编排（make docker-up）的容器和网络。
# 不删镜像，也不删数据卷。切到高可用不要用这条，用 docker-switch-ha。
docker-rm:
	docker compose down --remove-orphans
	@echo "已删除单机容器。镜像和数据卷都还在。切高可用请执行: make docker-switch-ha && make docker-up-ha"

# 从单机切到高可用。
# 单机 ES 不会进入高可用集群，旧 MySQL 里的日志库在新 ES 中没有索引，查询会 404。
# 停掉单机和高可用容器，并删除本项目的 MySQL、Redis 数据卷，下次启动回到初始库。
# 不删除镜像，也不删除高可用自己的 Kafka / ES 数据卷。
docker-switch-ha:
	docker compose -f docker-compose.ha.yaml down --remove-orphans
	docker compose down --remove-orphans
	@project=$$(docker compose config --format json | python3 -c 'import json,sys; print(json.load(sys.stdin)["name"])'); \
	for vol in mysql_data redis_data; do \
	  docker volume rm -f "$${project}_$${vol}" >/dev/null 2>&1 || true; \
	  echo "已删除数据卷 $${project}_$${vol}"; \
	done; \
	echo "MySQL 与 Redis 已回到未初始化状态。镜像仍在。接下来执行: make docker-up-ha"

docker-infra: prepare-host
	docker compose up -d mysql redis kafka zookeeper elasticsearch

docker-rebuild: prepare-host ensure-binaries ensure-web
	LCA_TARGETARCH=$(GOARCH_NATIVE) DOCKER_DEFAULT_PLATFORM=$(DOCKER_PLATFORM) \
		docker compose build --no-cache --build-arg TARGETARCH=$(GOARCH_NATIVE) $(COMPOSE_APPS)
	LCA_TARGETARCH=$(GOARCH_NATIVE) DOCKER_DEFAULT_PLATFORM=$(DOCKER_PLATFORM) \
		docker compose up -d --force-recreate $(COMPOSE_APPS)

# Agent 请用: cd deploy/logcollect && bash deploy.sh

docker-logs:
	docker compose logs -f --tail=100

docker-clean:
	docker compose down -v --rmi local

run-admin:
	$(GO) run ./cmd/admin -config configs/admin.yaml

run-apiserver:
	$(GO) run ./cmd/apiserver -config configs/apiserver.yaml

run-logtransfer:
	$(GO) run ./cmd/logtransfer -config configs/logtransfer.yaml
