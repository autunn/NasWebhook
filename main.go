package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// Config 结构体对应你的 config.json
type Config struct {
	CorpID         string `json:"corpid"`
	AgentID        string `json:"agentid"`
	CorpSecret     string `json:"corpsecret"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encoding_aes_key"`
	ProxyURL       string `json:"proxy_url"`
	NasURL         string `json:"nas_url"`
	PhotoURL       string `json:"photo_url"`
	Configured     bool   `json:"configured"`
}

var configPath = "data/config.json"

func main() {
	// 设置为发布模式减少日志干扰
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 确保数据目录存在
	os.MkdirAll("data", 0755)

	// 核心路由：支持所有请求方式
	r.Any("/webhook", handleMessage)

	// 配置管理界面（供你访问 5080 端口使用）
	r.GET("/", func(c *gin.Context) {
		conf := loadConfig()
		c.JSON(200, conf)
	})

	fmt.Println("NAS Webhook v4.2.0 (Super Compatible) Started on :5080")
	r.Run(":5080")
}

func handleMessage(c *gin.Context) {
	var text string

	// 1. 尝试从 URL 参数读取 (?text=xxx)
	text = c.Query("text")

	// 2. 如果没有，尝试从 POST 表单读取
	if text == "" {
		text = c.PostForm("text")
	}

	// 3. 如果还是没有，尝试解析 JSON Body
	if text == "" {
		// 先把 Body 读出来，防止解析 JSON 后 Body 消失
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		if len(bodyBytes) > 0 {
			var jsonData map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
				if t, ok := jsonData["text"].(string); ok {
					text = t
				}
			}
			// 调试日志：如果 text 还是空的，打印原始 Body 看看 qB 到底发了什么
			if text == "" {
				fmt.Printf("[DEBUG] Received raw unknown body: %s\n", string(bodyBytes))
			}
		}
	}

	// 打印接收日志，方便你用 docker logs -f 查看
	fmt.Printf("[INFO] Received Request - IP: %s, Text: '%s'\n", c.ClientIP(), text)

	if text == "" {
		c.JSON(200, gin.H{"status": "no_content_received"})
		return
	}

	conf := loadConfig()
	if conf.Configured {
		fmt.Printf("[ACTION] Pushing to WeChat: %s\n", text)
		go pushToWeChat(conf, text)
		c.JSON(200, gin.H{"status": "ok"})
	} else {
		fmt.Println("[WARN] Config not active (configured: false)")
		c.JSON(200, gin.H{"status": "not_configured"})
	}
}

func pushToWeChat(conf Config, content string) {
	// 这里的逻辑简化为发送文本，确保成功率
	// 实际生产中你可以根据需要恢复你的 News 图文消息逻辑
	tokenUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", conf.CorpID, conf.CorpSecret)
	
	resp, err := http.Get(tokenUrl)
	if err != nil {
		fmt.Printf("[ERROR] Get Token Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		ErrCode     int    `json:"errcode"`
	}
	json.NewDecoder(resp.Body).Decode(&tokenRes)

	if tokenRes.AccessToken == "" {
		fmt.Printf("[ERROR] Access Token is empty, ErrCode: %d\n", tokenRes.ErrCode)
		return
	}

	sendUrl := "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=" + tokenRes.AccessToken
	msg := map[string]interface{}{
		"touser":  "@all",
		"msgtype": "text",
		"agentid": conf.AgentID,
		"text": map[string]string{
			"content": content,
		},
	}

	body, _ := json.Marshal(msg)
	http.Post(sendUrl, "application/json", bytes.NewBuffer(body))
}

func loadConfig() Config {
	var conf Config
	file, err := os.ReadFile(configPath)
	if err != nil {
		return Config{Configured: false}
	}
	json.Unmarshal(file, &conf)
	return conf
}