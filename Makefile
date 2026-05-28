.PHONY: all build clean admin apiserver logcollect logcollect_win logtransfer syslog

BINARY_DIR=bin
GO=go
GOFLAGS=-v

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

# 交叉编译 Linux amd64 发布版本
release:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o release/bin/admin ./cmd/admin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o release/bin/apiserver ./cmd/apiserver
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o release/bin/logcollect ./cmd/logcollect
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o release/bin/logtransfer ./cmd/logtransfer
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o release/bin/syslog ./cmd/syslog

# 开发模式运行
run-admin:
	$(GO) run ./cmd/admin -config configs/admin.yaml

run-apiserver:
	$(GO) run ./cmd/apiserver -config configs/apiserver.yaml

run-logtransfer:
	$(GO) run ./cmd/logtransfer -config configs/logtransfer.yaml

# ==================== Docker ====================

# 启动所有服务
docker-up:
	docker compose up -d

# 停止所有服务
docker-down:
	docker compose down

# 只启动基础设施（开发模式）
docker-infra:
	docker compose up -d mysql redis kafka zookeeper elasticsearch

# 重新构建并启动
docker-rebuild:
	docker compose up -d --build admin apiserver logtransfer web

# 查看服务日志
docker-logs:
	docker compose logs -f --tail=100

# 清理所有Docker数据
docker-clean:
	docker compose down -v --rmi local
