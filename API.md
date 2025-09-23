### Utopia Server API 文档 (v2.0)

本文档遵循 OpenAPI 3.0 规范，详细描述了 `utopia-server` 的所有 API 端点。

---

#### **1. 认证 (Authentication)**

所有需要认证的端点都必须在 HTTP Header 中包含 `Authorization: Bearer <JWT>`。

##### **1.1 `POST /api/auth/register`**

*   **描述**: 注册一个新用户。
*   **请求体** (`application/json`):
    ```json
    {
      "username": "newuser",
      "password": "password123"
    }
    ```
*   **响应**:
    *   `201 Created`: 注册成功。
    *   `409 Conflict`: 用户名已存在。
    *   `400 Bad Request`: 请求体格式错误。

##### **1.2 `POST /api/auth/login`**

*   **描述**: 用户登录并获取 JWT。
*   **请求体** (`application/json`):
    ```json
    {
      "username": "testuser",
      "password": "password123"
    }
    ```
*   **响应**:
    *   `200 OK` (`application/json`): 登录成功。
        ```json
        {
          "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
        }
        ```
    *   `401 Unauthorized`: 用户名或密码错误。

---

#### **2. 节点管理 (Node Management)**

##### **2.1 `POST /api/nodes/register`**

*   **描述**: 供 `node-agent` 自动注册新节点。此端点无需认证。
*   **请求体** (`application/json`):
    ```json
    {
      "hostname": "gpu-node-01"
    }
    ```
*   **响应**:
    *   `201 Created` (`application/json`): 注册成功，返回分配给节点的唯一 ID。
        ```json
        {
          "id": "a1b2c3d4-e5f6-7890-1234-567890abcdef"
        }
        ```

##### **2.2 `GET /api/nodes/:id/status`**

*   **描述**: 获取指定节点的实时状态和指标。此接口会直接代理到 `node-agent` 的 `/api/v1/metrics` 端点。
*   **认证**: 需要 Bearer Token。
*   **路径参数**:
    *   `id` (string, required): 节点的唯一 ID。
*   **响应**:
    *   `200 OK` (`application/json`): 成功获取指标。响应体是 `node-agent` 返回的原始 JSON 数据。
        ```json
        {
          "node_id": "string",
          "cpu_usage_percent": "number",
          "memory_usage_percent": "number",
          "gpus": [ ... ],
          "system": { ... }
        }
        ```
    *   `401 Unauthorized`: 未提供或提供了无效的 JWT。
    *   `404 Not Found`: 指定的节点 ID 不存在。
    *   `409 Conflict`: 节点当前不是 `Online` 状态。
    *   `502 Bad Gateway`: 从 `node-agent` 获取指标失败。

---

#### **3. GPU 资源声明 (GPU Claims)**

**认证**: 需要 Bearer Token。

##### **3.1 `POST /api/gpu-claims`**

*   **描述**: 用户提交一个 GPU 资源申请。这是一个异步操作，服务器会接受请求并将其放入控制器的调和队列。
*   **请求体** (`application/json`):
    ```json
    {
      "spec": {
        "image": "nvidia/cuda:11.8.0-base-ubuntu22.04",
        "resources": {
          "gpuCount": 1
        }
      }
    }
    ```
*   **响应**:
    *   `202 Accepted` (`application/json`): 请求已被成功接受，并返回创建的 `GpuClaim` 的详细信息。
        ```json
        {
          "id": "claim-uuid-...",
          "user_id": 1,
          "spec": {
            "image": "nvidia/cuda:11.8.0-base-ubuntu22.04",
            "resources": {
              "gpuCount": 1
            }
          },
          "status": {
            "phase": "Pending",
            "nodeName": "",
            "containerID": "",
            "reason": ""
          },
          "created_at": "...",
          "updated_at": "..."
        }
        ```
    *   `401 Unauthorized`: 未提供或提供了无效的 JWT。
    *   `403 Forbidden`: 用户角色策略不允许此操作（例如，超出 GPU 配额）。
    *   `400 Bad Request`: 请求体格式错误。