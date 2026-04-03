# dapr-mdns

`dapr-mdns` 是一个用于查找 Dapr 服务的工具，通过 mDNS (Multicast DNS) 协议来发现网络中的 Dapr 服务。该工具可以解析 mDNS 服务条目，并提取出有关 Dapr 服务的详细信息。

## 快速开始

### 安装

1. **克隆仓库**

   ```bash
   git clone https://github.com/your-repo/dapr-mdns.git
   cd dapr-mdns
   ```

2. **编译**

   你可以根据目标平台编译工具。以下是编译示例：

   - **Linux (ARM64)**

     ```bash
     GOOS=linux GOARCH=arm64 go build -o dapr-mdns-arm64 main.go
     ```

   - **Linux (AMD64)**

     ```bash
     GOOS=linux GOARCH=amd64 go build -o dapr-mdns-amd64 main.go
     ```

3. **运行**

   使用以下命令运行工具：

   ```bash
   ./dapr-mdns-* -app-id <your-app-id> -lookup
   ```

   其中 `<your-app-id>` 是你要查找的 Dapr 服务的 AppID。

### 命令行参数

| 参数       | 描述                                       | 默认值        |
|------------|--------------------------------------------|---------------|
| `app-id`   | 要查找的 Dapr 服务的 AppID                 | (必填)        |
| `lookup`   | 是否执行查找操作                           | `false`       |
| `timeout`  | 查询超时时间，单位为秒                     | `10`          |

### 示例

1. **查找指定 AppID 的 Dapr 服务**

   ```bash
   ./dapr-mdns-arm64 -app-id lgdev-platform-iam-api2 -lookup
   ```

   输出示例：

   ```
   正在查找 AppID: lgdev-platform-iam-api2 (超时: 10s)
   ✅ 找到服务
   ✅ 共找到 1 个服务

   ===== 服务 1 =====
   ========= Dapr 服务信息 =========
   AppID:                lgdev-platform-iam-api2
   服务名称:              lgdev-platform-iam-api2._tcp.local
   实例名称:              lgdev-platform-iam-api2
   主机名:                lgdev-platform-iam-api2.local
   Dapr 内部端口:         8080
   IPv4 地址:            [10.151.47.35]
   IPv6 地址:            []
   文本记录:              []
   =================================

   ✅ 结束服务
   ```

2. **显示帮助信息**

   ```bash
   ./dapr-mdns-arm64
   ```

   输出示例：

   ```
   Usage of ./dapr-mdns-arm64:
     -app-id string
            指定要查找的 Dapr AppID
     -lookup
            是否查找指定的 AppID
     -timeout duration
            查询超时时间 (default 10s)
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

### 贡献

欢迎贡献代码和提出问题。如果您有任何建议或发现任何问题，请提交 [issue](https://github.com/your-repo/dapr-mdns/issues) 或者提交 [pull request](https://github.com/your-repo/dapr-mdns/pulls)。

### 许可

`dapr-mdns` 项目遵循 [木兰宽松许可证，第 2 版](LICENSE)。

## 联系方式

- **GitHub**: [https://github.com/yujia-ansteel/dapr-mdns](https://github.com/yujia-ansteel/dapr-mdns)
- **作者**: [pyschemDa] (guangdashao1990@163.com)