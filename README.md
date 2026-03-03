<p align="center">
  <img src="https://raw.githubusercontent.com/autunn/nas-webhook/main/logo.png" width="180" alt="NAS Webhook Logo" />
</p>

<p align="center">
  <h2>NAS Webhook</h2>
  <p>连接 NAS 与 企业微信的现代化桥梁</p>
</p>

<p align="center">
  <a href="https://github.com/autunn/nas-webhook">
    <img src="https://img.shields.io/badge/GitHub-Source%20Code-000000?style=flat-square&logo=github" />
  </a>
  <a href="https://hub.docker.com/r/autunn/nas-webhook">
    <img src="https://img.shields.io/docker/pulls/autunn/nas-webhook?style=flat-square&logo=docker&color=0db7ed" />
  </a>
  <a href="https://hub.docker.com/r/autunn/nas-webhook">
    <img src="https://img.shields.io/docker/image-size/autunn/nas-webhook?style=flat-square&logo=docker" />
  </a>
  <img src="https://img.shields.io/badge/License-MIT-2ecc71?style=flat-square" />
</p>


## 简介 (Introduction)

[cite_start]NAS Webhook 是一个用于将各类 NAS (如 UGREEN UGOS Pro, Synology, QNAP 等) 消息推送到企业微信的 Webhook 服务 [cite: 21, 25]。

## 特性 (Features)

- [cite_start]**精美 UI**：响应式管理后台，配置简单直观 。
- [cite_start]**安全验证**：支持自定义管理员密码与企业微信回调签名验证 。
- [cite_start]**多架构支持**：适配主流 NAS 环境 [cite: 21]。
- [cite_start]**图文通知**：支持带图片、标题和跳转链接的微信卡片消息 。

## 快速启动 (Quick Start)

### Docker CLI

```bash
docker run -d \
  --name nas-webhook \
  -p 5080:5080 \
  -v $(pwd)/data:/app/data \
  -e ADMIN_PASSWORD=这里填你的强密码 \
  --restart always \
  autunn/nas-webhook:latest
```

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
      - ADMIN_PASSWORD=这里填你的强密码
    restart: always
```

## 配置指南 (Configuration)

1、在企业微信后台创建自建应用，获取 CorpID 和 Secret 。

2、在本项目后台配置 Token 与 EncodingAESKey 。

3、在 NAS 的通知设置中添加 Webhook，URL 填写 http://你的IP:5080/webhook 。

## 挂载卷 (Volumes)

- `/app/data`：持久化配置与数据

## License

MIT License
