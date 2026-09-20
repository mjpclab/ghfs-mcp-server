# GHFS MCP Server 开发计划

## 1. 项目概述

开发一个 MCP (Model Context Protocol) Server，作为 AI 模型与 GHFS (Go HTTP File Server) 之间的桥梁，使 AI 能够通过 MCP 协议对 GHFS 进行文件操作。

- **调试实例地址**：`http://localhost:8080/`
- **可写目录**：仅 `/ttt/` 目录有上传、建目录和删除权限，其余目录为只读
- **GHFS API 文档**：https://github.com/mjpclab/go-http-file-server/blob/main/doc/en-US/api.md
- **开发语言**：Go

---

## 2. GHFS REST API 梳理

根据 API 文档，本项目需要对接的 GHFS 接口如下：

### 2.1 列出目录（List Directory）

```
GET <path>[?sort=<sortBy>]
Accept: application/json
```

- 通过设置 `Accept: application/json` 请求头获取 JSON 格式的目录数据。
- 支持排序参数 `sort`，可选值：`n/N`(名称)、`e/E`(类型)、`s/S`(大小)、`t/T`(时间)、`_`(不排序)。
- 目录排序前缀：`/<key>` 目录在前、`<key>/` 目录在后。

### 2.2 上传文件（Upload）

```
POST <path>?upload
Content-Type: multipart/form-data
```

- 必须使用 `POST` 方法和 `multipart/form-data` 编码。
- 表单字段名：
  - `file`：普通上传
  - `dirfile`：上传至相对子路径（需开启 mkdir）
  - `innerdirfile`：类似 dirfile，但去掉第一层目录

### 2.3 创建目录（Mkdir）

```
POST <path>?mkdir
Accept: application/json

name=<dir1>&name=<dir2>&...
```

- 请求体为 `application/x-www-form-urlencoded`，包含一个或多个 `name` 参数。
- 支持多级目录创建，如 `foo/bar/baz`。

### 2.4 删除（Delete）

```
POST <path>?delete
Accept: application/json

name=<item1>&name=<item2>&...
```

- 请求体为 `application/x-www-form-urlencoded`，包含一个或多个 `name` 参数。
- 目录会被递归删除。

---

## 3. MCP Tools 设计

### 3.1 `ghfs_list` — 列出目录

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 要列出的目录路径，如 `/` 或 `/docs/` |
| `sort` | string | 否 | 排序方式，如 `/T`（目录在前，按时间倒序） |

**返回**：JSON 格式的目录内容列表（文件名、大小、修改时间、是否目录等）。

### 3.2 `ghfs_upload` — 上传文件/目录

支持上传单个文件、多个文件、以及包含目录结构的文件集合。

底层原理：GHFS 的 upload API 基于 `multipart/form-data`，每个文件作为一个 part。通过不同的表单字段名实现不同的上传行为：
- `file`：上传到目标目录下（平铺）
- `dirfile`：文件名中包含相对路径，自动创建子目录（需 mkdir 同时开启）

因此"上传目录"本质上是将目录展开为多个带相对路径前缀的文件，逐一作为 `dirfile` part 上传。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 上传目标目录路径，如 `/docs/` |
| `files` | object[] | 是 | 要上传的文件列表 |
| `files[].filepath` | string | 是 | 文件的相对路径（含文件名）。单文件如 `report.txt`；目录结构中的文件如 `mydir/sub/file.txt` |
| `files[].content` | string | 是 | 文件内容（base64 编码） |

**行为说明**：
- 若 `filepath` 不含 `/`（如 `hello.txt`），以 `file` 字段上传，文件直接存放在 `path` 下。
- 若 `filepath` 含 `/`（如 `subdir/hello.txt`），以 `dirfile` 字段上传，GHFS 会自动创建中间目录。

**上传示例场景**：

1. **上传单个文件**到 `/docs/`：
   ```json
   {
     "path": "/docs/",
     "files": [
       {"filepath": "readme.txt", "content": "base64..."}
     ]
   }
   ```

2. **上传多个文件**到 `/docs/`：
   ```json
   {
     "path": "/docs/",
     "files": [
       {"filepath": "a.txt", "content": "base64..."},
       {"filepath": "b.txt", "content": "base64..."}
     ]
   }
   ```

3. **上传目录结构**（将本地 `project/` 目录上传到 `/docs/`）：
   ```json
   {
     "path": "/docs/",
     "files": [
       {"filepath": "project/src/main.go", "content": "base64..."},
       {"filepath": "project/README.md", "content": "base64..."},
       {"filepath": "project/config/app.yaml", "content": "base64..."}
     ]
   }
   ```
   结果：GHFS 上会生成 `/docs/project/src/main.go` 等文件及对应目录。

**返回**：上传结果（成功/失败信息，含每个文件的状态）。

### 3.3 `ghfs_mkdir` — 创建目录

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 父目录路径 |
| `names` | string[] | 是 | 要创建的目录名列表，支持多级如 `foo/bar` |

**返回**：创建结果。

### 3.4 `ghfs_delete` — 删除文件或目录

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 所在目录路径 |
| `names` | string[] | 是 | 要删除的文件/目录名列表 |

**返回**：删除结果。

### 3.5 `ghfs_archive` — 打包下载

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 要打包的目录路径 |
| `format` | string | 是 | 打包格式：`tar`、`tgz` 或 `zip` |
| `names` | string[] | 否 | 指定打包的子项名称，留空则打包整个目录 |
| `filename` | string | 否 | 自定义下载文件名 |

**返回**：生成的打包下载 URL，用户可直接访问该 URL 下载归档文件。

---

## 4. 项目结构

```
ghfs-mcp-server/
├── go.mod
├── go.sum
├── plan.md
├── main.go              # 程序入口，解析命令行参数，启动 STDIO 或 HTTP 模式
├── server/
│   ├── ghfs.go          # GHFS HTTP 客户端，封装所有 REST API 调用
│   ├── ghfs_test.go     # GHFS 客户端测试
│   ├── tools.go         # MCP Tools 定义与注册
│   └── handler.go       # Tool 调用处理逻辑（内部创建 GHFS 客户端）
└── README.md
```

---

## 5. 技术选型

| 组件 | 选型 | 说明 |
|------|------|------|
| MCP SDK | [go-sdk](https://github.com/modelcontextprotocol/go-sdk) | MCP 官方 Go SDK，支持 STDIO 和 Streamable HTTP 传输 |
| HTTP 客户端 | `net/http` (标准库) | 用于调用 GHFS REST API |
| 命令行参数 | `flag` (标准库) | 解析运行模式和 GHFS 地址 |

---

## 6. 实现步骤

### 第一阶段：基础框架搭建

#### Step 1：初始化项目依赖

- 引入官方 Go SDK：`go get github.com/modelcontextprotocol/go-sdk`
- 确认项目可编译通过

#### Step 2：实现 GHFS HTTP 客户端 (`server/ghfs.go`)

编写 `ghfsClient` 结构体（私有），封装以下方法：

```go
// ghfsClient is an HTTP client for interacting with a GHFS server.
type ghfsClient struct {
    baseURL    string
    httpClient *http.Client
}

type uploadFile struct {
    filepath string // 相对路径，如 "file.txt" 或 "subdir/file.txt"
    content  []byte // 文件内容
}

func newGHFSClient(baseURL string) *ghfsClient
func (c *ghfsClient) list(path string, sort string) ([]byte, error)
func (c *ghfsClient) upload(path string, files []uploadFile) error
func (c *ghfsClient) mkdir(path string, names []string) error
func (c *ghfsClient) delete(path string, names []string) error
func (c *ghfsClient) archiveURL(path string, format string, names []string, filename string) string
```

**upload 实现要点**：
- 遍历 `files` 列表，对每个文件创建一个 multipart part
- 若 `filepath` 不含 `/`，使用表单字段名 `file`
- 若 `filepath` 含 `/`，使用表单字段名 `dirfile`，GHFS 会自动创建中间目录

#### Step 3：单独测试 GHFS 客户端

- 编写单元测试或简单 main 函数，验证对 `http://localhost:8080/` 的各项操作是否正常。

### 第二阶段：MCP Server 集成

#### Step 4：定义 MCP Tools (`server/tools.go`)

使用官方 Go SDK 的泛型 `mcp.AddTool` 注册 5 个 Tool：

- `ghfs_list`
- `ghfs_upload`
- `ghfs_mkdir`
- `ghfs_delete`
- `ghfs_archive`

每个 Tool 的输入参数用一个 Go 结构体描述，JSON Schema 由 SDK 自动推导：无 `omitempty` 的字段为必填，`jsonschema` 标签作为字段说明。枚举等标签无法表达的约束，通过 `jsonschema.For` 推导后再补充。

#### Step 5：实现 Tool Handler (`server/handler.go`)

`NewHandler` 直接接收 GHFS URL 字符串，内部创建客户端：
- 接收 SDK 已反序列化并校验过的类型化入参
- 调用 `ghfsClient` 对应方法
- 组装并返回 MCP 响应

#### Step 6：实现主入口 (`main.go`)

```
用法: ghfs-mcp-server [options]

参数:
  -ghfs-url string    GHFS 服务器地址 (默认 "http://localhost:8080")
  -mode string        运行模式: stdio 或 http (默认 "stdio")
  -addr string        HTTP 模式监听地址 (默认 ":8080")
```

- **STDIO 模式**：使用 `server.Run(ctx, &mcp.StdioTransport{})` 通过标准输入输出通信
- **HTTP 模式**：用 `mcp.NewStreamableHTTPHandler` 取得 `http.Handler`，再交给 `http.ListenAndServe`（或 `ListenAndServeTLS`）启动服务

### 第三阶段：测试与完善

#### Step 7：集成测试

- 使用 MCP Inspector 或 Claude Desktop 连接本 MCP Server
- 逐一测试 4 个 Tool 的正确性：
  - 列出根目录 `/`
  - 创建目录 `/test-dir/`
  - 上传文件到 `/test-dir/hello.txt`
  - 列出 `/test-dir/` 确认文件存在
  - 删除 `/test-dir/`
  - 列出根目录确认已删除

#### Step 8：错误处理与边界情况

- GHFS 服务不可达时的友好提示
- 路径不存在时的错误返回
- 上传空文件 / 大文件的处理
- 删除不存在的文件时的处理
- 参数校验（路径格式、必填参数检查）

#### Step 9：编写 README

- 项目简介
- 编译与安装方式
- 使用说明（STDIO / HTTP 两种模式）
- MCP Tools 说明
- 配置示例（Claude Desktop `claude_desktop_config.json`）

---

## 7. 运行与配置示例

### STDIO 模式

```bash
ghfs-mcp-server -ghfs-url http://localhost:8080
```

Claude Desktop 配置：
```json
{
  "mcpServers": {
    "ghfs": {
      "command": "/path/to/ghfs-mcp-server",
      "args": ["-ghfs-url", "http://localhost:8080"]
    }
  }
}
```

### HTTP 模式

```bash
ghfs-mcp-server -mode http -addr :9090 -ghfs-url http://localhost:8080
```

---

## 8. 里程碑

| 里程碑 | 内容 | 预计产出 |
|--------|------|----------|
| M1 | GHFS 客户端 + 基础测试 | `server/ghfs.go` 可独立运行 |
| M2 | MCP Server 框架 + 4 个 Tool 注册 | STDIO 模式可启动 |
| M3 | 完整功能 + HTTP 模式 | 两种模式均可工作 |
| M4 | 测试、错误处理、文档 | 可交付使用 |
