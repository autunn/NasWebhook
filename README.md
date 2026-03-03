# NasWebhook

连接各类品牌 NAS 与 企业微信的通用化消息桥梁。

[![GitHub Source](https://img.shields.io/badge/GitHub-Source%20Code-000000?style=flat-square&logo=github)](https://github.com/autunn/NasWebhook)
[![Docker Pulls](https://img.shields.io/docker/pulls/autunn/nas-webhook.svg?style=flat-square&logo=docker&color=0db7ed)](https://hub.docker.com/r/autunn/nas-webhook)
[![Docker Image Size](https://img.shields.io/docker/image-size/autunn/nas-webhook/latest?style=flat-square&logo=docker)](https://hub.docker.com/r/autunn/nas-webhook)
[![License](https://img.shields.io/badge/License-MIT-2ecc71?style=flat-square)](https://github.com/autunn/NasWebhook/blob/main/LICENSE)

---

## 📖 简介 (Introduction)

**NasWebhook** 是一个通用的 Webhook 转发服务。它可以接收来自 **绿联 UGOS Pro、群晖 DSM、极空间** 等 NAS 系统，或 **qBittorrent** 等下载工具的 Webhook 信号，并统一推送到企业微信。

## ✨ 特性 (Features)

- **全品牌适配**：支持所有可配置标准 JSON Webhook 的系统。
- **Web 管理界面**：内置可视化配置后台，无需手动修改代码。
- **安全可靠**：集成官方加解密协议，支持自定义后台管理密码。
- **多架构构建**：原生支持 `linux/amd64` 和 `linux/arm64`。
- **图文美化**：支持调用第三方 API 接口显示动态动漫封面。

## 🚀 快速启动

### 1. 使用 Docker Compose (推荐)
```yaml
version: '3'
services:
  webhook:
    image: autunn/nas-webhook:latest
    container_name: nas-webhook
    ports:
      - "5080:5080"
    volumes:
      - ./data:/app/data
    environment:
      - ADMIN_PASSWORD=admin # 修改为你的后台登录密码
      - TZ=Asia/Shanghai
    restart: always

```

### 2. 使用 Docker CLI

```bash
docker run -d --name nas-webhook -p 5080:5080 -v $(pwd)/data:/app/data -e ADMIN_PASSWORD=admin autunn/nas-webhook:latest

```

## 📝 使用指南

1. **登录**：访问 `http://NAS_IP:5080`。
2. **配置**：填写企业微信 `CorpID`、`Secret`、`AgentID` 等信息。
3. **Webhook 地址**：在 NAS 端填写 `http://你的访问地址/webhook`。
4. **数据存储**：所有配置保存在挂载的 `/app/data` 目录中。
