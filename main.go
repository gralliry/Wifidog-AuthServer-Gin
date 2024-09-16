package main

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
)

/*
Hostname                  (Mandatory; Default: NONE)
SSLAvailable              (Optional; Default: no; Possible values: yes, no)
SSLPort                   (Optional; Default: 443)
HTTPPort                  (Optional; Default: 80)

Path                      (Optional; Default: /wifidog/       Note:  The path must be both prefixed and suffixed by /.  Use a single / for server root.)
LoginScriptPathFragment   (Optional; Default: login/?         Note:  This is the script the user will be sent to for login.)
PortalScriptPathFragment  (Optional; Default: portal/?        Note:  This is the script the user will be sent to after a successfull login.)
MsgScriptPathFragment     (Optional; Default: gw_message.php? Note:  This is the script the user will be sent to upon error to read a readable message.)
PingScriptPathFragment    (Optional; Default: ping/?          Note:  This is the script the user will be sent to upon error to read a readable message.)
AuthScriptPathFragment    (Optional; Default: auth/?          Note:  This is the script the user will be sent to upon error to read a readable message.)
*/

// Config 定义配置
type Config struct {
	// 认证服务器
	Host string
	Port string
	// 脚本路径
	Path                     string
	LoginScriptPathFragment  string
	PortalScriptPathFragment string
	MsgScriptPathFragment    string
	PingScriptPathFragment   string
	AuthScriptPathFragment   string
	// 网关
	GWAddress string
	GWPort    string
	GWId      string
}

var conf Config

func routePath(rootPath string, scriptPath string) string {
	fullPath := rootPath + scriptPath
	return fullPath[:len(fullPath)-1]
}

func main() {
	_, err := toml.DecodeFile("config.toml", &conf)
	if err != nil {
		log.Fatal("配置读取失败")
	}

	// 设置为Release模式，这将禁用Logger和Recovery中间件, 但是会关闭所有日志输出
	//gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)
	// 创建一个新的Gin引擎实例
	r := gin.New()
	// 设置模板文件的路径
	r.LoadHTMLGlob("./pages/*.html")

	// 连接到 SQLite 数据库文件
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal("连接数据库时失败")
	}
	// 确保在函数退出时关闭数据库连接
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Println("数据库关闭时发生了错误")
		}
	}(db)

	// 面向user，登录页面
	r.Handle("GET", routePath(conf.Path, conf.LoginScriptPathFragment), func(c *gin.Context) {
		gwAddress := c.Query("gw_address")
		gwPort := c.Query("gw_port")
		gwId := c.Query("gw_id")
		if gwAddress != conf.GWAddress || gwPort != conf.GWPort || gwId != conf.GWId {
			c.String(http.StatusBadRequest, fmt.Sprintf("当前页面(%s, %s, %s)不属于你处在的认证网络(%s, %s, %s)", conf.GWAddress, gwPort, gwId, conf.GWAddress, conf.GWPort, conf.GWId))
			return
		}
		c.HTML(http.StatusOK, "login.html", gin.H{})
	})

	// 面向user，登录请求
	r.Handle("POST", routePath(conf.Path, conf.LoginScriptPathFragment), func(c *gin.Context) {
		// 获取表单参数
		username := c.PostForm("username")
		password := c.PostForm("password")
		// 获取query参数
		gwAddress := c.Query("gw_address")
		gwPort := c.Query("gw_port")
		gwId := c.Query("gw_id")
		ip := c.Query("ip")
		mac := c.Query("mac")
		// 因为后台有可能存在应用是使用http发送请求，所以这里的url不一定是用户打开浏览器访问的url，为了不误导用户，宁愿不要跳转
		//url := c.Query("url")

		if gwAddress != conf.GWAddress || gwPort != conf.GWPort || gwId != conf.GWId {
			c.String(http.StatusBadRequest, fmt.Sprintf("当前页面(%s, %s, %s)不属于你处在的认证网络(%s, %s, %s)", conf.GWAddress, gwPort, gwId, conf.GWAddress, conf.GWPort, conf.GWId))
			return
		}
		var userId int
		// 执行参数化查询
		res := db.QueryRow(
			"SELECT id FROM user_info where username = ? and password = ?",
			username, password,
		).Scan(&userId)
		if res != nil {
			c.HTML(http.StatusOK, "message.html", gin.H{
				"message": "账号不存在或密码错误",
			})
			return
		}
		// 更新用户信息
		token := uuid.New().String()
		_, err := db.Exec(
			"INSERT INTO connection (token, ip, mac, user_id) VALUES (?, ?, ?, ?)",
			token, ip, mac, userId,
		)
		if err != nil {
			c.HTML(http.StatusOK, "message.html", gin.H{
				"message": "登录失败",
			})
			return
		}
		// 成功重定向
		c.Redirect(http.StatusFound, fmt.Sprintf(
			"http://%s:%s/wifidog/auth?token=%s",
			conf.GWAddress, conf.GWPort, token))
	})

	// 面向user，成功登录页面
	r.Handle("GET", routePath(conf.Path, conf.PortalScriptPathFragment), func(c *gin.Context) {
		gwId := c.Query("gw_id")
		if gwId != conf.GWId {
			c.String(http.StatusBadRequest, fmt.Sprintf(
				"当前页面(%s)不属于你处在的认证网络(%s, %s, %s)",
				gwId, conf.GWAddress, conf.GWPort, conf.GWId))
			return
		}
		c.HTML(http.StatusOK, "portal.html", gin.H{})
	})

	// 面向user，提示信息
	r.Handle("GET", routePath(conf.Path, conf.MsgScriptPathFragment), func(c *gin.Context) {
		message := c.Query("message")
		c.HTML(http.StatusOK, "portal.html", gin.H{
			"message": message,
		})
	})

	// 面向Wifidog, Ping
	r.Handle("GET", routePath(conf.Path, conf.PingScriptPathFragment), func(c *gin.Context) {
		gwId := c.Query("gw_id")
		// 暂时未用到这里的数据
		//sysUptime := c.Query("sys_uptime")
		//sysMemfree := c.Query("sys_memfree")
		//sysLoad := c.Query("sys_load")
		//wifidogUptime := c.Query("wifidog_uptime")
		if gwId != conf.GWId {
			// 不回应非当前网络的请求
			//c.Status(http.StatusInternalServerError)
			return
		}
		c.String(http.StatusOK, "Pong")
	})

	// 面向Wifidog, 验证
	r.Handle("GET", routePath(conf.Path, conf.AuthScriptPathFragment), func(c *gin.Context) {
		//stage := c.Query("stage")
		ip := c.Query("ip")
		mac := c.Query("mac")
		token := c.Query("token")
		incoming := c.Query("incoming")
		outgoing := c.Query("outgoing")
		gwId := c.Query("gw_id")
		if gwId != conf.GWId {
			c.String(http.StatusOK, "Auth: 0")
			return
		}
		// 执行参数化查询
		res := db.QueryRow(
			"SELECT 1 FROM connection where token = ?",
			token,
		).Scan()
		if res != nil {
			c.String(http.StatusOK, "Auth: 0")
			return
		}
		// 更新连接信息
		_, err := db.Exec(
			"UPDATE connection SET ip = ?, mac = ?, incoming = ?, outgoing = ? WHERE token = ?",
			ip, mac, incoming, outgoing, token,
		)
		if err != nil {
			c.String(http.StatusOK, "Auth: 0")
			return
		}
		c.String(http.StatusOK, "Auth: 1")
	})
	// 启动服务，监听端口
	err = r.Run(conf.Host + ":" + conf.Port)
	if err != nil {
		log.Fatal("服务端启动失败")
	}
}
