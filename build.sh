#!/bin/bash

# 构建前端
cd web
npm run build

# 创建静态文件目录
cd ../internal/handler
mkdir -p static

# 复制构建文件
cp -r ../../web/dist/* static/

# 返回项目根目录
cd ../../

# 构建 Go 应用
go build -o gomyredis_darwin_arm64 cmd/server/main.go

# 运行 Go 应用
./gomyredis_darwin_arm64