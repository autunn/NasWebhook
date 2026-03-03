<p align="center">
  <img src="https://raw.githubusercontent.com/autunn/NasWebhook/main/logo.png" width="180" alt="NasWebhook Logo" />
</p>

<p align="center">
  <h2>NasWebhook</h2>
  <p>连接各类品牌 NAS 与 企业微信的通用化消息桥梁</p>
</p>

![Pulls](https://img.shields.io/docker/pulls/autunn/nas-webhook)

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


## 简介 (Introduction)

NasWebhook 是一个通用的 Webhook 转发服务，旨在将各品牌 NAS（如群晖、绿联 UGOS Pro、极空间等）及下载工具（如 qBittorrent）的系统通知推送到企业微信。

## 特性 (Features)

- **品牌通用**：适配所有支持自定义 JSON Webhook 的系统。
- **可视化后台**：内置 Web 管理界面，轻松配置企业微信参数。
- **安全加固**：支持自定义管理员密码登录，并遵循企业微信加解密协议。
- **自动化构建**：支持 Docker 多架构（amd64/arm64）自动构建并推送到 Docker Hub。
- **个性化封面**：支持接入随机动漫图片 API，让通知卡片更精美。

## 快速启动 (Quick Start)

### Docker Compose (推荐)

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

### Docker CLI

```bash
docker run -d \
  --name nas-webhook \
  -p 5080:5080 \
  -v $(pwd)/data:/app/data \
  -e ADMIN_PASSWORD=admin \
  --restart always \
  autunn/nas-webhook:latest

```

## 配置指南 (Configuration)

1. **进入后台**：访问 `http://IP:5080`，使用密码登录（默认 `admin`）。
2. **填写参数**：在后台填写企业微信的 `CorpID`、`AgentID`、`Secret`、`Token` 和 `EncodingAESKey`。
3. **设置图片**：推荐在图片 API 栏填入 `https://api.vvhan.com/api/wallpaper/views?type=aliyun`。
4. **Webhook URL**：在 NAS 端填写 `http://你的访问地址/webhook`。

## 挂载卷 (Volumes)

* `/app/data`：用于持久化存储 `config.json` 配置文件。
