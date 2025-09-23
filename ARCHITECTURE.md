# Utopia Server - 架构与工作原理解析

本文档旨在深入剖析 `utopia-server` 的核心架构设计与关键工作流，为开发者和维护者提供一份清晰的内部指南。

## 1. 核心设计哲学：声明式系统

`utopia-server` 的根基是**声明式（Declarative）**设计，而非命令式（Imperative）。

*   **命令式**: 用户告诉系统“做什么”（例如，`POST /start-container`）。系统被动执行指令，但不关心最终结果。
*   **声明式**: 用户告诉系统“我想要什么”（例如，`POST /gpu-claims` 创建一个期望状态）。系统则主动、持续地工作，确保现实世界（Actual State）与用户的期望（Desired State）保持一致。

这种模式的核心优势在于其**健壮性和自愈能力**。如果一个容器意外崩溃，声明式系统会自动检测到“现实”与“期望”不符，并尝试重新创建容器，而无需人工干预。

## 2. 架构图

```mermaid
graph TD
    subgraph "外部用户与节点"
        User[<i class='fa fa-user'></i> 用户/管理员]
        NodeAgent["<i class='fa fa-server'></i> GPU 节点 [node-agent]"]
    end

    subgraph "Utopia Server [Go 应用]"
        direction LR

        subgraph "A. API 网关与 Web UI"
            APIGateway["Gin API 网关<br/>(JWT/RBAC 中间件)"]
        end

        subgraph "B. 核心服务"
            AuthService["认证与授权服务"]
            NodeService["节点管理服务"]
        end

        subgraph "C. 控制器与后台任务"
            Controller["声明式控制器<br/>(调和循环)"]
            Discovery["节点发现服务"]
            HealthChecker["节点健康检查"]
        end

        subgraph "D. 核心逻辑"
            Scheduler["调度与分配逻辑"]
            AgentClient["远程执行客户端"]
        end

        subgraph "E. 基础设施"
            FRPS["隧道服务宿主 [frps]"]
            DB[(<i class='fa fa-database'></i> MySQL 数据库)]
        end

        %% 内部模块交互
        APIGateway -- "路由请求" --> AuthService
        APIGateway -- "路由请求" --> NodeService
        APIGateway -- "路由请求" --> Controller

        Controller -- "1. 发现 Pending/Scheduled Claim" --> DB
        Controller -- "2. 请求调度" --> Scheduler
        Scheduler -- "3. 获取节点列表" --> DB
        Scheduler -- "4. 返回最佳节点" --> Controller
        Controller -- "5. 更新 Claim 状态为 'Scheduled'" --> DB
        Controller -- "6. 请求执行" --> AgentClient
        AgentClient -- "7. 通过隧道执行指令" --> NodeAgent

        Discovery -- "通过 Admin API 发现隧道" --> FRPS
        Discovery -- "更新节点端口/状态" --> DB
        HealthChecker -- "轮询 /status" --> AgentClient
        HealthChecker -- "更新节点 GPU/健康状态" --> DB
    end

    %% 外部交互
    User -- "登录, 提交 GpuClaim<br/>(REST API)" --> APIGateway
    NodeAgent -- "注册节点<br/>POST /nodes/register" --> APIGateway
    NodeAgent -- "建立反向隧道" --> FRPS
```

## 3. 关键工作流解析

### 工作流 1: 新节点自动上线 (Zero-Touch Onboarding)

1.  **注册**: 新 `node-agent` 启动，发现本地无 ID，于是调用 `utopia-server` 的 `POST /nodes/register` 接口。
2.  **分配身份**: `NodeService` 为其生成一个唯一的 UUID (`node-id`)，存入数据库，并将该 ID 返回给 `agent`。
3.  **建立隧道**: `agent` 保存 ID，用它动态生成 `frpc` 配置文件（例如，隧道名称为 `[node-id]_control`），并启动 `frpc` 子进程连接到服务器的 `frps` 服务。
4.  **服务发现**: `utopia-server` 的 `Discovery` 服务定期轮询 `frps` 的管理 API。它通过隧道名称识别出新节点，并解析出 `frps` 为其分配的公网端口 (`ControlPort`)。
5.  **上线**: `Discovery` 服务将节点的 `status` 更新为 `Online`，并将 `ControlPort` 存入数据库。至此，该节点正式加入资源池，可被调度。

### 工作流 2: GPU 资源声明与调和 (Reconciliation Loop)

1.  **声明期望**: 用户通过 UI 或 API 发送 `POST /api/gpu-claims` 请求，描述他们想要的容器镜像和 GPU 数量。
2.  **接受请求**: API 层通过认证和 RBAC 检查后，在数据库中创建一条 `GpuClaim` 记录，其 `status.phase` 初始为 `Pending`。
3.  **感知变化 (Perceive)**: `Controller` 的调和循环定期扫描数据库，发现了这条 `Pending` 的 `GpuClaim`。
4.  **决策 (Decide)**:
    *   控制器调用 `Scheduler`。
    *   `Scheduler` 从数据库获取所有 `Online` 状态的节点及其最新的 GPU 状态（由 `HealthChecker` 维护）。
    *   根据调度算法（例如，首次适应），选择一个最合适的节点。
    *   如果找不到合适的节点，本次调和结束，等待下一个周期重试。
5.  **行动 (Act)**:
    *   如果找到了节点，控制器将 `GpuClaim` 的 `status.phase` 更新为 `Scheduled`，并将 `status.nodeName` 设置为所选节点的 ID。
    *   在下一个调和周期，控制器发现这条 `Scheduled` 的 `Claim`。
    *   它调用 `AgentClient`，通过节点的 `ControlPort` 向 `node-agent` 的 `POST /api/v1/containers` 接口发送指令。
6.  **更新状态 (Update)**:
    *   `node-agent` 创建容器成功后，返回 `container_id`。
    *   `AgentClient` 将 `container_id` 返回给控制器。
    *   控制器最后一次更新 `GpuClaim`，将 `status.phase` 设置为 `Running`，并填入 `container_id`。
    *   如果 `node-agent` 执行失败，`phase` 则被设置为 `Failed`，并记录失败原因。

这个“感知-决策-行动-更新”的循环，就是 `utopia-server` 作为声明式系统自动化所有任务的核心工作原理。