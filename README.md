# Go-My-Redis

Go-My-Redis 是一个轻量级的Redis Web Admin管理工具，提供了直观的图形界面和强大的功能来管理和监控 Redis 服务器。

## 功能特点

- 🚀 现代化的用户界面
- 🔐 安全的连接管理和历史会话管理
- 📊 实时监控 Redis 服务器状态
- 🔍 强大的键值搜索和过滤功能
- 📝 支持多种数据类型的可视化展示(目前仅支持String Hash SET ZSET LIST)
- ⚡ 高性能的后端处理
- 🔄 实时数据更新
- 🖥 Xterm终端使用，后续将提供队列实时监控

## 效果图

![介绍页面](https://i.imgur.com/xJZ7w0A.png)
![连接页面](https://i.imgur.com/78B5mum.png)
![连接页面](https://i.imgur.com/aAFBt7y.png)
![终端操作](https://i.imgur.com/d6FBivI.png)
![终端操作](https://i.imgur.com/ODSgxRw.png)
![主题切换](https://i.imgur.com/78B5mum.png)


## 技术栈

### 后端
- Golang
- Redis 客户端库
- SSE [暂不支持]

### 前端
- Vue.js 3
- Naive UI 框架
- TypeScript

## 快速开始(本地需具备Go与Node环境)

```shell
chmod +x build.sh
./build.sh

# 交叉编译
# mac
go build -o gomyredis_darwin_arm64 cmd/server/main.go
# linux
GOOS=linux GOARCH=amd64 go build -o gomyredis_linux_amd64 cmd/server/main.go
# windows
GOOS=windows GOARCH=amd64 go build -o gomyredis_windows_x64.exe cmd/server/main.go
```

## 下载运行

### windows 可能会被防火墙阻止，请允许后，再使用管理员身份运行

[软件地址](https://github.com/voyager-go/go-my-redis/releases/tag/1.1)

## 贡献指南

欢迎提交 Pull Request 或创建 Issue。在提交代码前，请确保：

1. 代码符合项目的编码规范
2. 所有测试通过
3. 提交信息清晰明了

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交 Issue 或联系项目维护者。 