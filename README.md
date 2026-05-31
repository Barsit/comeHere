# ComeHere ⚡

**劫持闭源应用中硬编码的第三方 API 调用，透明转发到你的本地代理。**

[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache--2.0-blue)](LICENSE)

## 解决的问题

像 OpenAI Codex Desktop 这类闭源应用，硬编码了 API 端点地址（`https://api.openai.com/v1/responses`），用户无法通过常规配置改为自己的代理或第三方 API。

ComeHere 在网络层拦截这些请求——无需修改应用本身。

## 原理

```
应用（如 Codex）→ api.openai.com:443
                  ↓ (hosts 重定向)
            127.0.0.1:443  ← ComeHere TLS 代理
                  ↓ (转发)
            127.0.0.1:3000  ← 你的代理（ccx 等）
                  ↓
            DeepSeek / 任意 API
```

三步：
1. **hosts 文件** — 把目标域名指向 `127.0.0.1`
2. **自签名证书** — 自动签发域名证书并安装到系统信任存储
3. **TLS 代理** — 解密 HTTPS 请求，转发到你的目标地址

## 快速开始

### 下载

从 [Releases](https://github.com/Barsit/comeHere/releases) 下载最新版 `comehere.exe`。

### 运行

```bash
comehere.exe
```

首次运行会弹出 UAC 请求管理员权限（需要修改 hosts 文件和安装证书）。之后自动打开管理界面。

### WebUI 管理

访问 `http://localhost:8848`

### 添加劫持规则

添加一条规则即可拦截 `api.openai.com` 的请求：

| 字段 | 值 | 说明 |
|------|-----|------|
| 源域名 | `api.openai.com` | 要劫持的域名 |
| 源端口 | `443` | HTTPS 默认端口 |
| 目标地址 | `localhost:3000` | 你的 ccx 代理地址 |

点击「启用」→ hosts 自动写入 + 证书签发 → 开始拦截。

> ⚠️ 关闭程序后 hosts 会残留，请在退出前先暂停所有规则。

## 搭配 ccx 使用 Codex Desktop

这是 ComeHere 最典型的场景——让 OpenAI Codex Desktop 通过 ccx 代理使用 DeepSeek。

### 完整链路

```
Codex Desktop → api.openai.com:443
     ↓ ComeHere 劫持
127.0.0.1:443 → ComeHere TLS 代理
     ↓ 转发 HTTP
localhost:3000 → ccx 代理
     ↓ Responses → Chat Completions 翻译
api.deepseek.com → DeepSeek v4
```

### 步骤

**1. 启动 ccx 代理**

```bash
# 确保 ccx 运行在 localhost:3000
# ccx 需要配置 DeepSeek 上游
```

ccx 配置 (`D:\ccx\.config\config.json`) 参考：

```json
{
  "upstream": [{
    "baseUrl": "https://api.deepseek.com",
    "apiKeys": ["sk-你的DeepSeek密钥"],
    "serviceType": "openai",
    "name": "deepseek-chat",
    "status": "active"
  }],
  "responsesUpstream": [{
    "baseUrl": "https://api.deepseek.com",
    "apiKeys": ["sk-你的DeepSeek密钥"],
    "serviceType": "openai",
    "name": "deepseek-responses",
    "status": "active"
  }]
}
```

**2. 启动 ComeHere**

```bash
comehere.exe
# UAC 弹窗 → 确认 → WebUI 自动打开
```

**3. 添加劫持规则**

在 WebUI 中添加：

| 源域名 | 目标地址 |
|--------|---------|
| `api.openai.com` | `localhost:3000` |

点击「启用」。

**4. 打开 Codex**

Codex 的 API 调用会被透明劫持到 ccx → DeepSeek。无需修改 Codex 的任何配置。

## 单独使用（不搭配 ccx）

如果目标后端支持 HTTPS，可以直接转发：

| 源域名 | 目标地址 | 目标 HTTPS |
|--------|---------|-----------|
| `api.openai.com` | `api.deepseek.com:443` | ✅ |

这样 ComeHere 会直接把 HTTPS 请求转发到 DeepSeek（跳过 ccx）。

## 技术架构

```
┌─────────────────────────────────────┐
│          ComeHere.exe               │
│                                     │
│  ┌──────────┐  ┌──────────────┐    │
│  │ WebUI    │  │ Gin API      │    │
│  │ Vue 3    │◀─┤ :8848        │    │
│  │ (嵌入式)  │  │ 9 个 REST   │    │
│  └──────────┘  └──────┬───────┘    │
│                       │            │
│              ┌────────┴────────┐   │
│              │    规则引擎      │   │
│              │ SNI 匹配 + 转发  │   │
│              └────────┬────────┘   │
│                       │            │
│              ┌────────┴────────┐   │
│              │   TLS 代理 :443  │   │
│              └────────┬────────┘   │
│                       │            │
│  ┌──────────┐  ┌──────┴───────┐   │
│  │ hosts 管理│  │  证书管理(CA) │   │
│  └──────────┘  └──────────────┘   │
└─────────────────────────────────────┘
```

- **语言**: Go 1.22+
- **Web 框架**: Gin
- **前端**: Vue 3（嵌入式 SPA，`//go:embed`）
- **TLS**: 标准库 `crypto/tls`，自动签发域名证书
- **持久化**: `~/.comehere/config.json`
- **分发**: 单二进制

## 构建

```bash
# 构建前端
cd web && npm install && npm run build && cd ..

# 编译
go build -o comehere.exe .
```

## 对比其他方案

| 方案 | 原理 | 需要修改 Codex？ |
|------|------|-----------------|
| **ccx** | API 翻译代理 | ✅ 需要 Codex 主动发请求 |
| **codex-relay** | Responses ↔ Chat 翻译 | ✅ 需要 Codex 主动发请求 |
| **ComeHere** 🏆 | **网络层拦截** | ❌ **不需要** |

其他方案都依赖 Codex 自愿把请求发给代理。ComeHere 在网络层拦截，Codex 无法绕过。

## 许可证

Apache-2.0
