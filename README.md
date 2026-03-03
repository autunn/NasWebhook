<p align="center">
  <a href="https://github.com/autunn/NasWebhook">
    <img src="https://raw.githubusercontent.com/autunn/NasWebhook/main/logo.png" width="180" alt="NasWebhook Logo" />
  </a>
</p>

<p align="center">
  <h2 align="center">NasWebhook</h2>
  <p align="center">连接各类品牌 NAS 与 企业微信的通用化消息桥梁</p>
</p>

<p align="center">
  <a href="https://github.com/autunn/NasWebhook">
    <img src="https://img.shields.io/badge/GitHub-Source%20Code-000000?style=flat-square&logo=github" alt="GitHub" />
  </a>
  <a href="https://hub.docker.com/r/autunn/nas-webhook">
    <img src="https://img.shields.io/docker/pulls/autunn/nas-webhook?style=flat-square&logo=docker&color=0db7ed" alt="Docker Pulls" />
  </a>
  <a href="https://hub.docker.com/r/autunn/nas-webhook">
    <img src="https://img.shields.io/docker/image-size/autunn/nas-webhook/latest?style=flat-square&logo=docker" alt="Docker Image Size" />
  </a>
  <img src="https://img.shields.io/badge/License-MIT-2ecc71?style=flat-square" alt="License" />
</p>

---

**NasWebhook** 是一个通用的 Webhook 转发服务，旨在将各品牌 NAS（如群晖、绿联 UGOS Pro、极空间等）及下载工具（如 qBittorrent）的系统通知推送到企业微信。

## 🛠️ 特性 (Features)

- **品牌通用**：适配所有支持自定义 JSON Webhook 的系统。
- **可视化后台**：内置 Web 管理界面，轻松配置企业微信参数。
- **多架构支持**：支持 Docker 多架构（amd64/arm64）自动构建。
- **个性化封面**：支持接入随机动漫图片 API。

## 🚀 快速启动

### Docker Compose
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
      - ADMIN_PASSWORD=admin # 修改为你自己的后台密码
      - TZ=Asia/Shanghai
    restart: always

```

## 📝 配置指南

1. **登录后台**：访问 `http://IP:5080`，使用环境变量中的 `ADMIN_PASSWORD` 登录。
2. **企业微信配置**：在后台填写 `CorpID`、`AgentID`、`Secret`、`Token` 和 `EncodingAESKey`。
3. **Webhook URL**：在 NAS 端设置 Webhook 目标地址为 `http://你的访问地址/webhook`。
