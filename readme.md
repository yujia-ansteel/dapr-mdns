# dapr-mdns

`dapr-mdns` 是一个用于查找 Dapr 服务的工具，通过 mDNS (Multicast DNS) 协议来发现网络中的 Dapr 服务。该工具可以解析 mDNS 服务条目，并提取出有关 Dapr 服务的详细信息。

主要解决企业中，微服务系统每个单独使用 docker-compose 部署 Dapr 应用时，官方 Dashboard 无法知道全部应用以及实例的问题。该工具提供了实时服务发现和可视化界面，帮助企业更好地管理和监控 Dapr 微服务架构。

## 快速开始

### 安装

1. **克隆仓库**

   ```bash
   git clone https://github.com/yujia-ansteel/dapr-mdns.git
   cd dapr-mdns
   ```

2. **编译**

   你可以根据目标平台编译工具。以下是编译示例：

   - **Linux (ARM64)**

     ```bash
     GOOS=linux GOARCH=arm64 go build -o dapr-mdns-resolver-linux-arm64 ./cmd/dapr-mdns/main.go
     ```

   - **Linux (AMD64)**

     ```bash
     GOOS=linux GOARCH=amd64 go build -o dapr-mdns-resolver-linux-amd64 ./cmd/dapr-mdns/main.go
     ```

   - **Windows (AMD64)**

     ```bash
     GOOS=windows GOARCH=amd64 go build -o dapr-mdns-resolver-win-amd64.exe ./cmd/dapr-mdns/main.go
     ```

   - **macOS (AMD64)**

     ```bash
     GOOS=darwin GOARCH=amd64 go build -o dapr-mdns-resolver-darwin-amd64 ./cmd/dapr-mdns/main.go
     ```

   windows 环境下编译：


   -  **在 CMD（命令提示符）中**

      使用 `set` 命令设置环境变量，并用 `&&` 连接命令：
      
      ```cmd
      set GOOS=linux && set GOARCH=arm64 && go build -o dapr-mdns-resolver-linux-arm64 ./cmd/dapr-mdns/main.go
      ```
      
      或分步执行：
      
      ```cmd
      set GOOS=linux
      set GOARCH=arm64
      go build -o dapr-mdns-resolver-linux-arm64 ./cmd/dapr-mdns/main.go
      ```

   - **在 PowerShell 中**

      使用 `$env:` 前缀设置临时环境变量：
      
      ```powershell
      $env:GOOS="linux"; $env:GOARCH="arm64"; go build -o dapr-mdns-resolver-linux-arm64 ./cmd/dapr-mdns/main.go
      ```
      
      或更清晰的写法：
      
      ```powershell
      $env:GOOS = "linux"
      $env:GOARCH = "arm64"
      go build -o dapr-mdns-resolver-linux-arm64 ./cmd/dapr-mdns/main.go
      ```

$env:GOOS = "windows"
$env:GOARCH = "amd64"
3. **运行**

   工具支持两种运行模式：简单查找模式和完整服务模式。

   - **简单查找模式**（向后兼容）

     查找指定的 Dapr 服务：

     ```bash
     ./dapr-mdns-resolver-linux-arm64 -app-id <your-app-id> -timeout 2s
     ```

     其中 `<your-app-id>` 是你要查找的 Dapr 服务的 AppID。

   - **完整服务模式**（推荐）

     启动完整的服务发现系统，包括 mDNS 监听器、缓存刷新器和 Web UI：

     ```bash
     ./dapr-mdns-resolver-win-amd64.exe -server -app-id-prefix shaogd -web-port 8081 -web-base-path /dapr-mdns
     ```

     然后访问 `http://localhost:8081/dapr-mdns` 查看服务发现 Web 界面。

### 命令行参数

工具支持以下命令行参数：

| 参数 | 描述 | 默认值 | 适用模式 |
|------|------|--------|----------|
| `-app-id` | 指定要查找的 Dapr AppID（简单查找模式） | "" | 简单查找模式 |
| `-app-id-prefix` | 指定要监听的 Dapr AppID 前缀（服务模式） | "" | 服务模式 |
| `-server` | 启动完整服务模式（监听器 + 刷新器 + Web UI） | false | 服务模式 |
| `-web-port` | Web 服务器端口（仅在 `--server` 模式下有效） | 8080 | 服务模式 |
| `-web-base-path` | Web 服务器的基础路径（例如 /api/v1） | / | 服务模式 |
| `-refresh-interval` | 服务刷新间隔（仅在 `--server` 模式下有效） | 30s | 服务模式 |
| `-timeout` | 查询超时时间（简单查找模式） | 2s | 简单查找模式 |

**注意：** 如果同时指定了 `-server` 参数，工具将忽略其他简单查找参数并启动完整服务模式。

### 示例

#### 1. 简单查找模式

**查找指定 AppID 的 Dapr 服务**

```bash
./dapr-mdns-resolver-linux-amd64 -app-id lgdev-platform-iam-api2 -timeout 2s
```

输出示例：

```
正在查找 AppID: lgdev-platform-iam-api2 (超时: 2s)
✅ 找到服务
✅ 共找到 1 个服务

===== 服务 1 =====
========== Dapr 服务信息 ==========
AppID:                lgdev-platform-iam-api2
服务名称:              lgdev-platform-iam-api2._tcp.local
实例名称:              lgdev-platform-iam-api2
主机名:                lgdev-platform-iam-api2.local
Dapr 内部端口:         8080
IPv4 地址:            [10.151.47.35]
IPv6 地址:            []
文本记录:              []
==================================
✅ 结束服务
```

#### 2. 完整服务模式

**启动完整的服务发现系统**

```bash
./dapr-mdns-resolver-linux-amd64 -server -app-id-prefix shaogd -web-port 8081 -refresh-interval 30s
```

输出示例：

```
启动 Dapr mDNS 服务发现服务器
Web UI 地址: http://localhost:8081
服务刷新间隔: 30s
IPv4 mDNS 查询捕获器已启动
IPv6 mDNS 查询捕获器已启动
路由注册完成:
  /static/ -> static file server
  /api/services -> handleAPIServices
  /api/services/ -> handleAPIServiceDetail
  /api/search -> handleAPISearch
  / -> handleIndex
  /health -> handleHealth
Web 服务器启动，监听端口 8081
开始监听请求...
```

#### 3. 显示帮助信息

```bash
./dapr-mdns-resolver-linux-amd64 -h
```

输出示例：

```
Usage of ./dapr-mdns-resolver.exe:
  -app-id string
        指定要查找的 Dapr AppID
  -app-id-prefix string
        指定要查找的 Dapr AppID 前缀
  -refresh-interval duration
        服务刷新间隔（仅在 --server 模式下有效） (default 30s)
  -server
        启动完整服务模式（监听器 + 刷新器 + Web UI）
  -timeout duration
        查询超时时间 (default 2s)
  -web-port int
        Web 服务器端口（仅在 --server 模式下有效） (default 8080)
```

### 功能说明

1. **DaprServiceInfo 结构体**

   `DaprServiceInfo` 结构体用于存储 Dapr 服务的信息，包括 AppID、服务名称、实例名称、端口号、主机名、IPv4 和 IPv6 地址、以及文本记录。

2. **DaprMDNSParser 结构体**

   `DaprMDNSParser` 结构体提供 mDNS 解析功能，包括创建解析器、提取 AppID、解析服务信息、查找服务和打印服务信息等方法。

3. **查找服务**

   使用 `LookupDaprService` 方法查找指定 AppID 的 Dapr 服务。该方法会在指定的超时时间内查找并返回所有匹配的服务信息。

4. **打印服务信息**

   使用 `PrintDaprServiceInfo` 方法打印 Dapr 服务的详细信息。

### Web UI 使用指南

启动完整服务模式后，可以通过浏览器访问 Web UI（默认地址：`http://localhost:8080`）。

#### 主要功能

1. **服务概览**
   - 显示所有发现的 Dapr 服务按 AppID 分组
   - 每个服务实例显示详细信息：实例名称、主机名、端口、IP地址等
   - 实时状态指示（在线/离线）

2. **搜索功能**
   - 在顶部搜索框中输入 AppID 可以过滤显示的服务
   - 点击"搜索"按钮可以主动触发 mDNS 查询，查找指定 AppID 的服务
   - 支持回车键快速搜索

3. **状态指示**
   - **在线状态**：绿色标签，表示服务最近被发现（60秒内）
   - **离线状态**：红色标签，表示服务超过60秒未更新
   - 数据每5秒自动刷新

4. **时间信息**
   - 首次发现时间
   - 最后更新时间
   - 最后发现时间

5. **API 接口**
   Web UI 提供以下 REST API 接口：
   - `GET /api/services` - 获取所有服务
   - `GET /api/services/{appID}` - 获取指定 AppID 的服务
   - `GET /api/search?appID={appID}` - 搜索指定 AppID 的服务
   - `GET /health` - 健康检查

#### 界面截图

```
+----------------------------------------------+
| Dapr mDNS 服务发现                           |
| 发现的服务数量: 5     最后更新时间: ...       |
| [搜索框] [搜索] [清除] [手动刷新]            |
+----------------------------------------------+
| AppID: my-app (3个实例)                      |
| +----------------+  +----------------+       |
| | 实例1 (在线)    |  | 实例2 (在线)    |       |
| | 主机: ...      |  | 主机: ...      |       |
| | 端口: 8080     |  | 端口: 8081     |       |
| | IPv4: ...      |  | IPv4: ...      |       |
| +----------------+  +----------------+       |
| | 实例3 (离线)    |                          |
| | 主机: ...      |                          |
| | 端口: 8082     |                          |
| | IPv4: ...      |                          |
| +----------------+                          |
|                                            |
| AppID: another-app (2个实例)                |
| ...                                        |
+----------------------------------------------+
```

### 项目架构

`dapr-mdns` 采用模块化设计，主要包含以下核心组件：

#### 1. **命令行界面** (`cmd/dapr-mdns/main.go`)
   - 解析命令行参数
   - 支持简单查找模式和完整服务模式
   - 提供向后兼容性

#### 2. **解析器模块** (`internal/parser/`)
   - `DaprMDNSParser` - mDNS 解析器
   - 提取 Dapr AppID 从服务名称
   - 解析 mDNS 服务条目为结构化数据
   - 支持实时服务查找

#### 3. **缓存模块** (`internal/cache/`)
   - `ServiceCache` - 服务信息缓存
   - 按 AppID 分组存储服务实例
   - 支持自动过期清理（120秒）
   - 提供线程安全的读写操作

#### 4. **服务器模块** (`internal/server/`)
   - `MDNSListener` - mDNS 查询监听器
     - 捕获网络中的 mDNS 查询流量
     - 过滤特定 AppID 前缀的服务
     - 自动触发服务信息刷新
   - `Refresher` - 定时刷新器
     - 定期刷新缓存中的服务信息
     - 可配置刷新间隔（默认30秒）
     - 保持服务信息最新

#### 5. **Web 模块** (`internal/web/`)
   - `WebServer` - Web 服务器
     - 提供 RESTful API
     - 服务 HTML 模板界面
     - 支持服务搜索功能
   - 实时 Web UI 界面
   - 健康检查端点

#### 6. **类型定义** (`pkg/types/`)
   - `DaprServiceInfo` - Dapr 服务信息结构体
   - `CacheItem` - 缓存项结构体（包含时间戳）

#### 工作流程

1. **服务发现**：mDNS 监听器捕获网络中的 Dapr 服务查询
2. **信息获取**：解析器查询服务详细信息（IP地址、端口等）
3. **缓存更新**：服务信息存储到缓存中，带时间戳
4. **定期刷新**：刷新器定期更新缓存中的服务状态
5. **Web 展示**：Web UI 从缓存获取数据实时展示
6. **用户交互**：用户可通过 Web UI 搜索特定服务或手动刷新

### 使用场景

`dapr-mdns` 适用于以下场景：

1. **企业微服务管理**
   - 当企业使用多个独立的 docker-compose 部署 Dapr 应用时
   - 官方 Dapr Dashboard 无法跨部署单元发现服务
   - 需要统一视图查看所有 Dapr 服务实例

2. **开发测试环境**
   - 开发人员在本地启动多个 Dapr 服务进行测试
   - 需要快速验证服务发现是否正常
   - 查看服务实例的详细信息（IP、端口等）

3. **生产环境监控**
   - 监控 Dapr 服务的在线状态
   - 及时发现服务实例离线或异常
   - 统计服务实例数量和分布

4. **跨网络服务发现**
   - 在局域网内发现所有 Dapr 服务
   - 支持多播 DNS (mDNS) 协议
   - 自动发现新服务，无需手动配置

### 限制和注意事项

1. **网络要求**
   - 需要支持多播的网络环境
   - 防火墙需允许 UDP 5353 端口（mDNS）
   - 需要在同一局域网或 VLAN 中

2. **服务过滤**
   - 通过 `-app-id-prefix` 参数过滤服务
   - 只发现指定前缀的 Dapr 服务
   - 避免发现非相关的 mDNS 服务
   - 被动发现

3. **缓存机制**
   - 服务信息缓存 120 秒后自动过期
   - 离线服务标记为过期状态（60秒未更新）
   - 定期刷新保持信息最新

4. **平台支持**
   - 支持 Windows、Linux、macOS
   - 需要 Go 1.26.1 或更高版本编译
   - Web UI 需要现代浏览器支持

### 开发测试工具

项目还包含一个 `tools/` 目录，提供了本地测试 Dapr 服务的批处理脚本：

1. **start-dapr.bat** - 启动测试用 Dapr 服务实例
2. **stop-dapr.bat** - 停止所有 Dapr 服务实例
3. **readme.md** - 工具使用说明

这些工具用于在 Windows 环境下快速搭建测试环境，验证 `dapr-mdns` 的服务发现功能。

### 应用管理脚本

项目还提供了一个 Shell 脚本 `manage_dapr_mdns.sh`，用于在 Linux 环境下方便地管理 `dapr-mdns` 应用的启动、停止、重启和状态监控。

**脚本功能：**

1. **启动应用** (`start`) - 启动 `dapr-mdns` 服务
2. **停止应用** (`stop`) - 停止正在运行的 `dapr-mdns` 服务
3. **重启应用** (`restart`) - 重启 `dapr-mdns` 服务
4. **状态查看** (`status`) - 查看应用运行状态
5. **日志查看** (`log`) - 实时查看应用日志
6. **帮助信息** (`help`) - 显示脚本使用说明

**使用方法：**

```bash
# 启动应用
./manage_dapr_mdns.sh start

# 停止应用
./manage_dapr_mdns.sh stop

# 重启应用
./manage_dapr_mdns.sh restart

# 查看应用状态
./manage_dapr_mdns.sh status

# 查看应用日志（实时跟踪）
./manage_dapr_mdns.sh log

# 显示帮助信息
./manage_dapr_mdns.sh help
```

**脚本配置：**

脚本中的以下变量可以根据需要调整：

```bash
APP_NAME="dapr-mdns-resolver"              # 应用名称
APP_BIN="./dapr-mdns-resolver-linux-arm64" # 可执行文件路径
APP_ARGS="-server -app-id-prefix yj3dev -web-base-path /dapr-mdns -web-port 8881" # 启动参数
LOG_FILE="mdns.log"                        # 日志文件
PID_FILE="dapr_mdns.pid"                   # PID文件
```

**注意事项：**

- 确保脚本有执行权限：`chmod +x manage_dapr_mdns.sh`
- 脚本会检查应用二进制文件是否存在并自动添加执行权限
- 启动失败时会自动清理 PID 文件并显示错误信息
- 日志文件会记录应用的标准输出和标准错误
- 支持通过 PID 文件或进程名两种方式管理应用

### 贡献

欢迎贡献代码和提出问题。如果您有任何建议或发现任何问题，请提交 [issue](https://github.com/yujia-ansteel/dapr-mdns/issues) 或者提交 [pull request](https://github.com/yujia-ansteel/dapr-mdns/pulls)。

### 许可

`dapr-mdns` 项目遵循 [木兰宽松许可证，第 2 版](LICENSE)。

## 联系方式

- **GitHub**: [https://github.com/yujia-ansteel/dapr-mdns](https://github.com/yujia-ansteel/dapr-mdns)
- **作者**: [pyschemDa] (guangdashao1990@163.com)