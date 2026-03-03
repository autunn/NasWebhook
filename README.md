
# NasWebhook

![GitHub](https://img.shields.io/badge/GitHub-Source%20Code-000000?style=flat-square&logo=github)
![Pulls](https://img.shields.io/docker/pulls/autunn/nas-webhook?style=flat-square&logo=docker&color=0db7ed)
![Size](https://img.shields.io/docker/image-size/autunn/nas-webhook/latest?style=flat-square&logo=docker)
![License](https://img.shields.io/badge/License-MIT-2ecc71?style=flat-square)

---

**NasWebhook** 是一个通用的 Webhook 转发服务，旨在将各品牌 NAS（如群晖、绿联 UGOS Pro、极空间等）及下载工具（如 qBittorrent）的系统通知推送到企业微信。

## 🛠️ 特性 (Features)

- **品牌通用**：适配所有支持自定义 JSON Webhook 的系统。
- **管理后台**：内置 Web 界面，支持在线配置企业微信参数。
- **安全加固**：支持自定义管理员密码登录，遵循加密协议。
- **自动化构建**：支持 Docker 多架构（amd64/arm64）自动构建。
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

1. **登录后台**：访问 `http://IP:5080`，默认密码为环境变量中的 `ADMIN_PASSWORD`。
2. **企业微信配置**：在后台填写 `CorpID`、`AgentID`、`Secret`、`Token` 和 `EncodingAESKey`。
3. **设置图片 API**：推荐填入 `https://api.vvhan.com/api/wallpaper/views?type=aliyun`。
4. **测试通知**：
```bash
curl -X POST -H "Content-Type: application/json" -d "{\"message\":\"NasWebhook 测试通知\"}" http://你的IP:5080/webhook

```



## 📁 挂载卷

* `/app/data`：存储 `config.json` 配置文件。
