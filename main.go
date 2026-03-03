package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

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
var accessToken string
var accessTokenExpiresAt int64
var sessionToken string

// App 信息
var AppName = "NAS Webhook"
var Version = "v3.3.1"

func init() {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		log.Fatal("Failed to generate session token")
	}
	sessionToken = hex.EncodeToString(b)
}

func main() {
	os.MkdirAll("data", 0755)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// 默认密码改为更通用的 admin123 [cite: 21]
	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "admin123" 
	}
	log.Printf("%s Version: %s", AppName, Version)

	r.GET("/login", func(c *gin.Context) {
		if checkCookie(c) {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.HTML(http.StatusOK, "login.html", gin.H{"version": Version})
	})

	r.POST("/login", func(c *gin.Context) {
		if c.PostForm("password") == adminPass {
			c.SetCookie("auth_session", sessionToken, 3600*24, "/", "", false, true)
			c.Redirect(http.StatusFound, "/")
		} else {
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "密码错误", "version": Version})
		}
	})

	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("auth_session", "", -1, "/", "", false, true)
		c.Redirect(http.StatusFound, "/login")
	})

	r.GET("/webhook", handleWebhookVerify)
	r.POST("/webhook", handleWebhookMsg)

	authorized := r.Group("/")
	authorized.Use(AuthMiddleware())
	{
		authorized.GET("/", func(c *gin.Context) {
			conf := loadConfig()
			c.HTML(http.StatusOK, "index.html", gin.H{
				"config":  conf,
				"success": c.Query("success"),
				"version": Version,
			})
		})
		authorized.POST("/save", handleSave)
	}

	log.Println("Server :5080 Started")
	r.Run(":5080")
}

// ... (AuthMiddleware, checkCookie, handleSave 等函数保持逻辑不变，仅需注意变量引用)

func sendToWeChat(conf Config, data map[string]interface{}) {
	token, err := getAccessToken(conf)
	if err != nil { return }
	
	content := "系统事件"
	if msg, ok := data["message"].(string); ok { content = msg }

	// 修改推送卡片标题为通用名称
	payload := map[string]interface{}{
		"touser":  "@all",
		"msgtype": "news",
		"agentid": conf.AgentID,
		"news": map[string]interface{}{
			"articles": []map[string]interface{}{
				{
					"title":       "NAS 系统通知", 
					"description": fmt.Sprintf("[%s]\n%s", time.Now().Format("15:04"), content),
					"url":         conf.NasURL,
					"picurl":      conf.PhotoURL,
				},
			},
		},
	}
    // ... 发送逻辑同原代码 [cite: 22]
}