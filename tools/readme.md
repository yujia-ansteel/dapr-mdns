
# Dapr 本地开发测试工具

本目录包含用于本地开发环境测试 Dapr 服务的工具脚本。

## 概述

这些工具用于在本地 Windows 环境下快速启动和停止 Dapr 服务实例，便于开发和测试。

## 前置条件

在运行脚本之前，请确保以下条件已满足：

1. **Dapr CLI** - 安装 Dapr 命令行工具
2. **Daprd** - Dapr 运行时，通常随 Dapr CLI 一起安装
3. **Python 环境** - 用于运行简单的 HTTP 服务器作为应用服务
4. **Windows 系统** - 这些脚本是 Windows 批处理文件

## 脚本说明

### start-dapr.bat

启动两个 Dapr 服务实例用于测试。

#### 用法
```batch
start-dapr.bat [appid] [app-port]
```

#### 参数
- `appid` - Dapr 应用的 AppID（例如：`shaogd-app`）
- `app-port` - 第一个应用的端口号（例如：`8100`）

#### 功能说明
该脚本将启动两个 Dapr 服务实例：
1. **App1** - 使用指定的 `appid` 和 `app-port`
2. **App2** - 使用相同的 `appid`，端口号为 `app-port + 1`

同时，每个实例都会启动一个对应的 Python HTTP 服务器，作为应用服务器。

#### 端口分配逻辑
脚本根据输入的 `app-port` 自动计算以下端口：
- Dapr HTTP 端口：`app-port - 4500`
- Dapr gRPC 端口：`52000 + (app-port - 8100 + 1)`
- Dapr 内部 gRPC 端口：`48444 + (app-port - 8100)`

#### 示例
```batch
start-dapr.bat shaogd-app 8100
```

这将启动：
- App1: Dapr 实例运行在端口 8100，对应的 Python HTTP 服务器也在端口 8100
- App2: Dapr 实例运行在端口 8101，对应的 Python HTTP 服务器也在端口 8101

### stop-dapr.bat

停止所有正在运行的 Dapr 服务实例。

#### 用法
```batch
stop-dapr.bat
```

#### 功能说明
- 强制终止所有 `daprd.exe` 进程
- 提供清理机制，避免端口占用冲突
- 提示手动关闭 Python HTTP 服务器窗口

## 使用流程

1. **启动测试环境**
   ```batch
   cd tools
   start-dapr.bat my-app 8100
   ```

2. **进行测试**
   - 使用 `dapr-mdns-resolver` 工具查找服务
   - 测试服务发现功能
   - 验证 mDNS 服务注册

3. **停止测试环境**
   ```batch
   stop-dapr.bat
   ```

## 注意事项

1. **端口冲突** - 确保使用的端口没有被其他应用占用
2. **Python 服务器** - Python HTTP 服务器窗口需要手动关闭
3. **日志级别** - 第一个实例使用 `--log-level=debug` 便于调试
4. **Metrics** - 所有实例都禁用 metrics (`--enable-metrics=false`)

## 与 dapr-mdns-resolver 配合使用

这些工具主要用于测试 `dapr-mdns-resolver` 项目的服务发现功能。启动 Dapr 服务后，可以使用以下命令测试服务发现：

```batch
dapr-mdns-resolver.exe -server -app-id-prefix shaogd -web-port 8081
```

然后访问 `http://localhost:8081` 查看服务发现结果。

## 故障排除

1. **脚本无法启动** - 检查 Dapr CLI 是否正确安装，确保 `daprd.exe` 在系统 PATH 中
2. **端口已被占用** - 修改 `app-port` 参数使用其他端口
3. **Python 未安装** - 安装 Python 或确保 `python` 命令可用