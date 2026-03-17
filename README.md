# NapSec — 隐私数据管家

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-blue)]()

> 本地优先、零依赖、开箱即用的文件隐私保护工具。

---

## 功能特性

| 功能 | 说明                              |
|------|---------------------------------|
| 实时监控 | 基于 fsnotify 监控目录变更              |
| 智能检测 | 内置 10+ 条正则规则，覆盖 API Key、私钥、身份证等 |
| AES-256 加密 | PBKDF2 密钥派生 + GCM 认证加密          |
| Git 审计日志 | 所有操作自动 Git commit，可追溯           |
| Web 仪表盘 | 内置可视化面板，无需额外依赖                  |
| 演习模式 | `--dry-run` 只检测不修改文件            |

---

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/liangach/napsec.git
cd napsec

# 编译
make build
```

### 基本用法

```bash
# 监控当前目录（演习模式）
napsec start . --dry-run

# 监控指定目录并启用加密保护
napsec start ~/Documents --password mypassword

# 查看状态
napsec status

# 列出已保护文件
napsec list

# 恢复文件
napsec recover ~/.napsec/vault/secret.txt.napsec

# 启动 Web 仪表盘
napsec web --port 8080
```

---

## 项目结构

```
napsec/
├── cmd/napsec/          # CLI 入口
│   └── commands/          # 子命令
├── internal/
│   ├── monitor/           # 文件监控（fsnotify）
│   ├── detector/          # 敏感信息检测（正则引擎）
│   ├── executor/          # 加密执行（AES-256-GCM）
│   ├── audit/             # 审计日志（Git）
│   ├── core/              # 核心引擎 + Web API
│   └── config/            # 配置管理
└── web/                   # Web 仪表盘前端
```

---

## 安全设计

- 所有加密操作使用 **AES-256-GCM**（认证加密）
- 密钥由 **PBKDF2**（100000 次迭代）从主密码派生
- 加密文件存储在独立的**隔离保险箱**目录
- **原始文件**在加密成功后被删除，原位置替换为提示文件
- Web 服务默认只监听 `localhost`，不暴露公网

## 操作指南
# NapSec 操作指南

## 一、基础操作

### 1. 目录监控与敏感文件保护

```bash
# 1.1 基础监控（演习模式，只检测不加密）
napsec start /path/to/monitor --dry-run

# 1.2 生产级监控（启用加密，自动保护敏感文件）
napsec start /path/to/monitor --password your-password

# 1.3 自定义监控配置（指定工作线程数）
napsec start /path/to/monitor \
  --password your-password \
  --workers 8
```

**参数说明：**
- `--password, -p`：加密密码（如不提供，会交互式提示输入）
- `--vault, -v`：加密保险箱路径（默认：~/.napsec/vault）
- `--dry-run, -d`：演习模式，只检测不执行保护操作
- `--workers, -w`：并发线程数（默认：4）

### 2. 加密文件管理

```bash
# 2.1 列出最近保护的加密文件
napsec list
napsec list --limit 50  # 显示最近50条记录（默认20条）

# 2.2 恢复加密文件到指定路径
napsec recover ~/.napsec/vault/secret.txt.napsec --output ~/recovered-secret.txt
```

**参数说明：**
- `list --limit, -n`：显示最近 N 条记录（默认20）
- `recover --output, -o`：恢复到指定路径（默认：~/Desktop/recovered）
- `recover --password, -p`：解密密码（如不提供，会交互式提示输入）

### 3. 审计与状态查看

```bash
# 3.1 查看 NapSec 运行状态
napsec status
```

**状态信息包含：**
- 总保护文件数
- 今日保护文件数
- 最后操作时间
- 审计日志总条数

## 二、Web 仪表盘操作

### 1. 启动 Web 服务

```bash
# 启动 Web 仪表盘（默认端口 8080）
napsec web

# 指定端口启动
napsec web --port 9090

# 开发模式启动
napsec web --dev
```

### 2. 访问 Web 界面

启动后访问：`http://localhost:8080`（或指定端口）

**Web 界面功能：**
- 实时统计看板（通过 `/api/stats` 接口）
- 审计记录列表（通过 `/api/records` 接口）
- 健康检查（通过 `/api/health` 接口）

![Web 界面](img.png)

### 3. Web API 接口

NapSec Web 服务提供以下 REST API：

- `GET /api/stats` - 获取实时统计信息
- `GET /api/records` - 获取最近50条审计记录
- `GET /api/health` - 健康检查

## 三、演习操作

### 1. 演习模式（Dry Run）

适合首次使用时验证规则有效性，仅检测敏感文件不执行加密/删除操作：

```bash
# 基础演习
napsec start /path/to/monitor --dry-run
```

### 2. 后台运行

#### Windows 系统
```bash
# 后台启动（使用 start /b）
start /b napsec start D:\work --password your-password
```

#### Linux/macOS 系统
```bash
# 使用 nohup 后台运行
nohup napsec start ~/work --password your-password > ~/.napsec/napsec.log 2>&1 &
```

## 四、常见问题

### 1. 密码输入问题

如果启动时未提供 `--password` 参数，程序会交互式提示输入：
```bash
napsec start ~/Documents
请输入加密密码：
请再次输入密码：
```

### 2. 停止监控

按 `Ctrl+C` 即可停止正在运行的监控服务。

### 3. 目录权限

NapSec 会自动创建所需目录（`~/.napsec/vault` 和 `~/.napsec/audit`），并设置为 `0700` 权限。

## 五、安全注意事项

1. **密码管理**：
    - 启动监控时输入的密码请妥善保管
    - 恢复文件时需要提供相同的密码
    - 忘记密码将无法恢复已加密文件

2. **加密文件存储**：
    - 加密文件保存在 `~/.napsec/vault` 目录
    - 建议定期备份此目录到安全位置

3. **审计日志**：
    - 所有操作记录在 `~/.napsec/audit` 目录
    - 使用 Git 进行版本管理，可查看操作历史

## 六、命令速查表

| 命令 | 用途 | 示例 |
|------|------|------|
| `napsec start` | 启动监控 | `napsec start ~/Documents -p 123456` |
| `napsec status` | 查看状态 | `napsec status` |
| `napsec list` | 列出记录 | `napsec list -n 50` |
| `napsec recover` | 恢复文件 | `napsec recover ~/.napsec/vault/file.napsec -o ~/file` |
| `napsec web` | 启动Web | `napsec web -p 8080` |
| `napsec --help` | 查看帮助 | `napsec --help` |
| `napsec [命令] --help` | 查看子命令帮助 | `napsec start --help` |

