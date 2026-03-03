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

var configPath = "data/config.json"
var accessToken string
var accessTokenExpiresAt int64
var sessionToken string

// App 信息
var AppName = "NAS Webhook"
var Version = "v4.0.1"

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

func AuthMiddleware() gin.HandlerFunc {
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
	cookie, err := c.Cookie("auth_session")
	if err != nil { return false }
	return cookie == sessionToken
}

func handleSave(c *gin.Context) {
	newConfig := Config{
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
	if newConfig.NasURL == "" { newConfig.NasURL = "http://localhost" }
	saveConfig(newConfig)
	c.Redirect(http.StatusSeeOther, "/?success=true")
}

func handleWebhookVerify(c *gin.Context) {
	conf := loadConfig()
	msgSignature := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	if msgSignature == "" {
		c.String(http.StatusBadRequest, "Invalid Request")
		return
	}
	if !verifySignature(conf.Token, timestamp, nonce, echostr, msgSignature) {
		c.String(http.StatusForbidden, "Sign Error")
		return
	}
	decryptedMsg, _ := decryptEchoStr(conf.EncodingAESKey, echostr)
	c.String(http.StatusOK, string(decryptedMsg))
}

func handleWebhookMsg(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON error"})
		return
	}
	conf := loadConfig()
	if !conf.Configured {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not configured"})
		return
	}
	go sendToWeChat(conf, data)
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func verifySignature(token, timestamp, nonce, echostr, msgSignature string) bool {
	params := []string{token, timestamp, nonce, echostr}
	sort.Strings(params)
	str := strings.Join(params, "")
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil)) == msgSignature
}

func decryptEchoStr(encodingAESKey, echostr string) ([]byte, error) {
	aesKey, _ := base64.StdEncoding.DecodeString(encodingAESKey + "=")
	cipherText, _ := base64.StdEncoding.DecodeString(echostr)
	block, _ := aes.NewCipher(aesKey)
	iv := aesKey[:16]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)
	pad := int(cipherText[len(cipherText)-1])
	if pad < 1 || pad > 32 { pad = 0 }
	cipherText = cipherText[:len(cipherText)-pad]
	msgLen := binary.BigEndian.Uint32(cipherText[16:20])
	return cipherText[20 : 20+int(msgLen)], nil
}

func loadConfig() Config {
	var conf Config
	data, err := os.ReadFile(configPath)
	if err != nil { return Config{Configured: false} }
	json.Unmarshal(data, &conf)
	return conf
}

func saveConfig(conf Config) {
	data, _ := json.MarshalIndent(conf, "", "  ")
	os.WriteFile(configPath, data, 0644)
}

func getAccessToken(conf Config) (string, error) {
	if accessToken != "" && accessTokenExpiresAt > time.Now().Unix()+60 {
		return accessToken, nil
	}
	baseURL := "https://qyapi.weixin.qq.com"
	if conf.ProxyURL != "" { baseURL = conf.ProxyURL }
	url := fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", baseURL, conf.CorpID, conf.CorpSecret)
	resp, err := http.Get(url)
	if err != nil { return "", err }
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if token, ok := result["access_token"].(string); ok {
		accessToken = token
		accessTokenExpiresAt = time.Now().Unix() + 7200
		return accessToken, nil
	}
	return "", fmt.Errorf("Token Error")
}

func sendToWeChat(conf Config, data map[string]interface{}) {
	token, err := getAccessToken(conf)
	if err != nil { return }
	content := "系统事件"
	if msg, ok := data["message"].(string); ok {
		content = msg
	} else if val, ok := data["data"].(map[string]interface{}); ok {
		if text, ok := val["text"].(string); ok { content = text }
	}

	baseURL := "https://qyapi.weixin.qq.com"
	if conf.ProxyURL != "" { baseURL = conf.ProxyURL }
	
	picURL := conf.PhotoURL
	if picURL == "" {
		picURL = fmt.Sprintf("https://picsum.photos/600/300?random=%d", time.Now().UnixNano())
	}
	agentID, _ := strconv.Atoi(conf.AgentID)

	payload := map[string]interface{}{
		"touser":  "@all",
		"msgtype": "news",
		"agentid": agentID,
		"news": map[string]interface{}{
			"articles": []map[string]interface{}{
				{
					"title":       "NAS 系统通知",
					"description": fmt.Sprintf("[%s]\n%s", time.Now().Format("15:04"), content),
					"url":         conf.NasURL,
					"picurl":      picURL,
				},
			},
		},
	}
	body, _ := json.Marshal(payload)
	postURL := fmt.Sprintf("%s/cgi-bin/message/send?access_token=%s", baseURL, token)
	http.Post(postURL, "application/json", bytes.NewBuffer(body))
}