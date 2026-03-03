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

var (
	configPath           = "data/config.json"
	accessToken          string
	accessTokenExpiresAt int64
	sessionToken         string
	Version              = "v4.2.2-Stable"
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

	// 1. 静态资源映射（确保 Logo 正常显示）
	r.StaticFile("/logo.png", "./logo.png")
	r.LoadHTMLGlob("templates/*")

	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "admin"
	}

	// 2. 身份验证路由
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

	// 3. Webhook 核心路由
	r.GET("/webhook", handleVerify)
	r.POST("/webhook", handleMessage)

	// 4. 管理后台
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

	log.Printf("NAS Webhook %s 已启动 (端口 :5080)", Version)
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
	echostr := c.Query("echostr")

	// 兼容：非验证请求但带参数时，直接转交给消息处理逻辑
	if echostr == "" && (c.Query("text") != "" || c.Query("message") != "" || c.Query("task") != "") {
		handleMessage(c)
		return
	}

	conf := loadConfig()
	msgSig := c.Query("msg_signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")

	params := []string{conf.Token, timestamp, nonce, echostr}
	sort.Strings(params)
	h := sha1.New()
	h.Write([]byte(strings.Join(params, "")))
	if fmt.Sprintf("%x", h.Sum(nil)) != msgSig {
		log.Printf("[Verify] 签名校验失败")
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
	data := make(map[string]interface{})

	// 合并 URL 参数
	for k, v := range c.Request.URL.Query() {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}

	// 合并 JSON Body
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	if len(bodyBytes) > 0 {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
			for k, v := range jsonData {
				data[k] = v
			}
		}
	}

	if len(data) > 0 {
		conf := loadConfig()
		if conf.Configured {
			log.Printf("[Webhook] 触发消息发送: %v", data)
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
	if token == "" { return }

	// 描述构造
	var description strings.Builder
	description.WriteString(fmt.Sprintf("时间: %s", time.Now().Format("15:04:05")))
	for k, v := range data {
		description.WriteString(fmt.Sprintf("\n%s: %v", k, v))
	}

	// 核心逻辑：随机老婆 API + 纳秒级去缓存
	picURL := conf.PhotoURL
	if picURL == "" {
		// 默认使用国内加速二次元接口，拼接纳秒级随机数
		picURL = fmt.Sprintf("https://api.vvhan.com/api/wallpaper/acg?rand=%d", time.Now().UnixNano())
	} else {
		// 如果用户填写了自定义 API，强制追加随机参数防止微信缓存旧图
		connector := "?"
		if strings.Contains(picURL, "?") { connector = "&" }
		picURL = fmt.Sprintf("%s%sv=%d", picURL, connector, time.Now().UnixNano())
	}

	agentID, _ := strconv.Atoi(conf.AgentID)
	payload := map[string]interface{}{
		"touser":  "@all",
		"msgtype": "news",
		"agentid": agentID,
		"news": map[string]interface{}{
			"articles": []map[string]interface{}{{
				"title":       "NAS 通知中心",
				"description": description.String(),
				"url":         conf.NasURL,
				"picurl":      picURL,
			}},
		},
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/cgi-bin/message/send?access_token=%s", baseURL, token)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err == nil {
		defer resp.Body.Close()
		resBody, _ := io.ReadAll(resp.Body)
		log.Printf("[WeChat] 发送回执: %s", string(resBody))
	} else {
		log.Printf("[WeChat] 网络故障: %v", err)
	}
}

func getWeChatToken(conf Config, baseURL string) string {
	if accessToken != "" && accessTokenExpiresAt > time.Now().Unix() {
		return accessToken
	}
	resp, err := http.Get(fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", baseURL, conf.CorpID, conf.CorpSecret))
	if err != nil { return "" }
	defer resp.Body.Close()

	var res struct {
		Token string `json:"access_token"`
		Exp   int64  `json:"expires_in"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	accessToken = res.Token
	accessTokenExpiresAt = time.Now().Unix() + res.Exp - 60
	return accessToken
}