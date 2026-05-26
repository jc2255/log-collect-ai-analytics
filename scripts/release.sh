#!/bin/bash
set -e
cd "$(dirname "$0")/.."

echo "==> [1/3] 后端交叉编译 linux/amd64"
for app in admin apiserver logcollect logtransfer; do
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o release/bin/$app ./cmd/$app
done
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o release/bin/logcollect.exe ./cmd/logcollect_win

echo "==> [2/3] 前端构建"
(cd web && npm run build)
rm -rf release/web-dist/assets/
cp -r web/dist/* release/web-dist/

echo "==> [3/3] 完成。变更文件:"
git status --short