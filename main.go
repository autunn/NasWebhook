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
	"fmt"
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

var (
	configPath           = "data/config.json"
	accessToken          string
	accessTokenExpiresAt int64
	sessionToken         string
	Version              = "v4.1.0"
)

func init() {
	b := make([]byte, 16)
	rand.Read(b)
	sessionToken = hex.EncodeToString(b)
}

func main() {
	_ = os.MkdirAll("data", 0755)
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 核心修改：托管本地 logo 映射
	r.StaticFile("/logo.png", "./logo.png")

	r.LoadHTMLGlob("templates/*")

	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "admin"
	}

	r.GET("/login", func(c *gin.Context) {
		if checkCookie(c) {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.HTML(http.StatusOK, "login.html", gin.H{"version": Version})
	})

	r.POST("/login", func(c *gin.Context) {
		if c.PostForm("password") == adminPass {
			c.SetCookie("auth_session", sessionToken, 86400, "/", "", false, true)
			c.Redirect(http.StatusFound, "/")
		} else {
			c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "密码错误", "version": Version})
		}
	})

	r.GET("/logout", func(c *gin.Context) {
		c.SetCookie("auth_session", "", -1, "/", "", false, true)
		c.Redirect(http.StatusFound, "/login")
	})

	r.GET("/webhook", handleVerify)
	r.POST("/webhook", handleMessage)

	auth := r.Group("/")
	auth.Use(authMiddleware())
	{
		auth.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"config":  loadConfig(),
				"success": c.Query("success"),
				"version": Version,
			})
		})
		auth.POST("/save", handleSave)
	}

	log.Printf("NAS Webhook %s Started", Version)
	r.Run(":5080")
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkCookie(c) {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func checkCookie(c *gin.Context) bool {
	cookie, _ := c.Cookie("auth_session")
	return cookie == sessionToken
}

func loadConfig() Config {
	var conf Config
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &conf)
	}
	return conf
}

func handleSave(c *gin.Context) {
	conf := Config{
		CorpID:         c.PostForm("corpid"),
		AgentID:        c.PostForm("agentid"),
		CorpSecret:     c.PostForm("corpsecret"),
		Token:          c.PostForm("token"),
		EncodingAESKey: c.PostForm("encoding_aes_key"),
		ProxyURL:       strings.TrimRight(c.PostForm("proxy_url"), "/"),
		NasURL:         strings.TrimRight(c.PostForm("nas_url"), "/"),
		PhotoURL:       c.PostForm("photo_url"),
		Configured:     true,
	}
	data, _ := json.MarshalIndent(conf, "", "  ")
	os.WriteFile(configPath, data, 0644)
	c.Redirect(http.StatusSeeOther, "/?success=true")
}

func handleVerify(c *gin.Context) {
	conf := loadConfig()
	msgSig := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	params := []string{conf.Token, timestamp, nonce, echostr}
	sort.Strings(params)
	h := sha1.New()
	h.Write([]byte(strings.Join(params, "")))
	if fmt.Sprintf("%x", h.Sum(nil)) != msgSig {
		c.AbortWithStatus(403)
		return
	}

	aesKey, _ := base64.StdEncoding.DecodeString(conf.EncodingAESKey + "=")
	cipherText, _ := base64.StdEncoding.DecodeString(echostr)
	block, _ := aes.NewCipher(aesKey)
	mode := cipher.NewCBCDecrypter(block, aesKey[:16])
	mode.CryptBlocks(cipherText, cipherText)
	msgLen := binary.BigEndian.Uint32(cipherText[16:20])
	c.String(200, string(cipherText[20:20+msgLen]))
}

func handleMessage(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err == nil {
		conf := loadConfig()
		if conf.Configured {
			go pushToWeChat(conf, data)
		}
	}
	c.JSON(200, gin.H{"status": "ok"})
}

func pushToWeChat(conf Config, data map[string]interface{}) {
	baseURL := "https://qyapi.weixin.qq.com"
	if conf.ProxyURL != "" {
		baseURL = conf.ProxyURL
	}

	token := getWeChatToken(conf, baseURL)
	if token == "" {
		return
	}

	content := "收到来自 NAS 的通知"
	if m, ok := data["message"].(string); ok {
		content = m
	}
	if t, ok := data["text"].(string); ok {
		content = t
	}

	picURL := conf.PhotoURL
	if picURL == "" {
		picURL = fmt.Sprintf("https://picsum.photos/600/300?random=%d", time.Now().Unix())
	}

	agentID, _ := strconv.Atoi(conf.AgentID)
	payload := map[string]interface{}{
		"touser":  "@all",
		"msgtype": "news",
		"agentid": agentID,
		"news": map[string]interface{}{
			"articles": []map[string]interface{}{{
				"title":       "NAS 通知中心",
				"description": fmt.Sprintf("时间: %s\n内容: %s", time.Now().Format("15:04:05"), content),
				"url":         conf.NasURL,
				"picurl":      picURL,
			}},
		},
	}

	body, _ := json.Marshal(payload)
	http.Post(fmt.Sprintf("%s/cgi-bin/message/send?access_token=%s", baseURL, token), "application/json", bytes.NewBuffer(body))
}

func getWeChatToken(conf Config, baseURL string) string {
	if accessToken != "" && accessTokenExpiresAt > time.Now().Unix() {
		return accessToken
	}
	resp, err := http.Get(fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", baseURL, conf.CorpID, conf.CorpSecret))
	if err != nil {
		return ""
	}
	var res struct {
		Token string `json:"access_token"`
		Exp   int64  `json:"expires_in"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	accessToken = res.Token
	accessTokenExpiresAt = time.Now().Unix() + res.Exp - 60
	return accessToken
}