# Utopia GPU 调度平台 - 中央服务器

*   **项目代号**: Utopia
*   **组件**: `utopia-server`
*   **版本**: 2.0

## 1. 项目概述

`utopia-server` 是 Utopia GPU 调度平台的**大脑与中央神经系统**。它是一个独立的、多功能的 Go 应用程序，作为所有交互（用户、管理员、GPU 节点）的中心枢纽，并承担系统**唯一事实来源 (Single Source of Truth)** 的角色。

其核心任务是通过一个声明式的、自动化的工作流，高效、安全地管理和调度整个 GPU 资源池。

## 2. 核心特性

*   **声明式核心**: 通过持续运行的**调和循环 (Reconciliation Loop)**，将系统的实际状态驱动到用户期望的状态。
*   **API 优先**: 所有平台功能均通过内部 RESTful API 暴露。
*   **零配置节点接入**: GPU 节点可自动向平台注册并获取身份，实现“即插即用”。
*   **统一的健康与指标拉取**: 服务器主动轮询每个 `Node Agent` 的状态接口，完成健康检查和性能指标采集。
*   **可扩展的访问控制**: 基于用户-角色-策略的 RBAC 模型，支持灵活的权限控制。

## 3. 技术栈

*   **编程语言**: Go
*   **核心框架**: Gin (Web Framework)
*   **数据库**: MySQL 5.7+
*   **核心通信**: `frp` (反向隧道)
*   **配置管理**: Viper
*   **测试**: testify

## 4. 快速开始

### 4.1. 环境依赖

*   Go 1.18+
*   MySQL 5.7+
*   `frps` 可执行文件（如果需要由 `utopia-server` 托管）

### 4.2. 配置

1.  复制或重命名 `configs/config.yaml` 文件。
2.  修改文件中的配置项，特别是数据库 DSN：
    ```yaml
    database:
      dsn: "user:password@tcp(127.0.0.1:3306)/utopia?parseTime=true"
    ```

### 4.3. 运行服务

在项目根目录下执行以下命令：

```bash
go run cmd/utopia-server/main.go
```

服务启动时，它会自动执行以下操作：
1.  连接到 MySQL 并自动创建 `utopia` 数据库（如果不存在）。
2.  运行所有数据库迁移，创建所需的表。
3.  启动 `frps` 子进程。
4.  启动 API 服务器，默认监听 `8080` 端口。
5.  启动所有后台服务（控制器、节点发现、健康检查）。

### 4.4. 访问 Web UI

打开浏览器并访问 `http://localhost:8080`。您将被重定向到 Web UI 页面，在这里您可以注册、登录和申请 GPU 资源。

## 5. 运行测试

确保您的本地 MySQL 服务正在运行。然后，在项目根目录下执行：

```bash
# 运行指定包的测试
go test ./internal/api/...

# 运行所有测试
go test ./...
```

## 6. API 文档

详细的 API 接口文档请参见项目根目录下的 [`API.md`](./API.md) 文件。

## 7. 项目结构

```
.
├── cmd/utopia-server/      # 主程序入口
├── configs/                # 配置文件目录
├── internal/               # 项目私有代码
│   ├── api/                # API 网关、处理器、中间件
│   ├── auth/               # 用户认证与 RBAC 服务
│   ├── client/             # 与 node-agent 通信的客户端
│   ├── config/             # 配置加载逻辑
│   ├── controller/         # 声明式控制器与调和循环
│   ├── database/           # 数据库连接与迁移
│   ├── models/             # 核心数据模型
│   ├── node/               # 节点管理、发现、健康检查
│   ├── scheduler/          # 调度与分配逻辑
│   └── tunnel/             # frps 隧道服务管理
└── web/ui/                 # 前端静态文件