package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	// ListenHost 认证服务器
	ListenHost   string
	ListenPort   int
	ConfigPath   string
	DatabasePath = "./data/database.db"
	SchemePath   = "./data/scheme.sql"
	// HtmlPath 页面认证
	HtmlPath        = "./pages/*.html"
	LoginHtmlPath   = "login.html"
	MessageHtmlPath = "message.html"
	PortalHtmlPath  = "portal.html"
)

var pathConf struct {
	// 脚本路径
	LoginScriptPath  string
	PortalScriptPath string
	MsgScriptPath    string
	PingScriptPath   string
	AuthScriptPath   string
	//
	IsSSL bool
}

func main() {
	// 使用flag包解析参数
	flag.StringVar(&ListenHost, "h", "0.0.0.0", "Host address")
	flag.IntVar(&ListenPort, "p", 8003, "Port number")
	flag.StringVar(&ConfigPath, "c", "config.toml", "Path of config file")
	// 解析命令行参数
	flag.Parse()
	_, err := toml.DecodeFile(ConfigPath, &pathConf)
	if err != nil {
		log.Fatal("配置读取失败")
	}
	// 设置为Release模式，这将禁用Logger和Recovery中间件, 但是会关闭所有日志输出
	// 应该由编译或运行时指定
	// 创建一个新的Gin引擎实例
	var r = gin.New()
	// 设置模板文件的路径
	r.LoadHTMLGlob(HtmlPath)
	// 设置代理
	if r.SetTrustedProxies(nil) != nil {
		log.Fatal("信任代理错误")
	}
	// 使用 os.Stat 检查文件是否存在
	_, err = os.Stat(DatabasePath)
	var isNotExist = os.IsNotExist(err)
	// 连接到 SQLite 数据库文件
	// 确保在函数退出时关闭数据库连接 log.Println("数据库关闭时发生了错误")
	db, err := gorm.Open(sqlite.Open(DatabasePath), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库时失败")
	}
	// database时自动创建
	if isNotExist {
		// 读取 schema.sql 文件
		sqlFile, err := os.ReadFile(SchemePath)
		if err != nil {
			log.Fatal("读取 schema.sql 文件时失败: ", err)
		}
		// 执行 SQL 文件中的内容
		sqlStatements := string(sqlFile)
		if err := db.Exec(sqlStatements).Error; err != nil {
			log.Fatal("执行 SQL 文件时失败: ", err)
		}
		log.Println("数据库结构已成功创建！")
	}
	// 测试数据库 // 检查结果
	if db.Exec("UPDATE conn SET is_expire = 1 WHERE is_expire = 0").Error != nil {
		log.Fatal("数据库测试连接失败")
	}

	// 面向user，登录页面
	r.Handle("GET", pathConf.LoginScriptPath,
		func(context *gin.Context) {
			// http://127.0.0.1:8003/wifidog/login/?gw_address=10.10.10.1&gw_port=2060&gw_id=64644ADFE3CE&ip=10.10.10.131&mac=fc:5b:8c:86:be:92
			// 网关信息
			var (
				gwAddress = context.Query("gw_address")
				gwPort    = context.Query("gw_port")
				gwId      = context.Query("gw_id")
			)
			context.HTML(http.StatusOK, LoginHtmlPath, gin.H{
				"gw_address": gwAddress,
				"gw_port":    gwPort,
				"gw_id":      gwId,
			})
		})

	// 面向user，登录请求
	r.Handle("POST", pathConf.LoginScriptPath,
		func(context *gin.Context) {
			var (
				// 获取表单参数
				username = context.PostForm("username")
				password = context.PostForm("password")
				// 获取query参数
				gwAddress = context.Query("gw_address")
				gwPort    = context.Query("gw_port")
				gwId      = context.Query("gw_id")
				ip        = context.Query("ip")
				mac       = context.Query("mac")
				// 因为后台有可能存在应用是使用http发送请求，所以这里的url不一定是用户打开浏览器访问的url
				//url       = context.Query("url")
			)
			var result *gorm.DB

			// ----------账号认证----------
			var userId int
			// 查询用户是否存在
			result = db.Raw(
				"SELECT id FROM user where account = ? and password = ?",
				username, password,
			).Scan(&userId)
			if result.Error != nil {
				context.Status(http.StatusInternalServerError)
				return
			}
			if result.RowsAffected != 1 {
				context.HTML(http.StatusOK, MessageHtmlPath, gin.H{
					"message": "账号不存在或密码错误",
				})
				return
			}
			// 查询网络是否存在(可以分开两个，查询是否存在再查询是否匹配)
			var netId int
			result = db.Raw(
				"SELECT id FROM net where address = ? and port = ? and sid = ?",
				gwAddress, gwPort, gwId,
			).Scan(&netId)
			if result.Error != nil {
				context.Status(http.StatusInternalServerError)
				return
			}
			if result.RowsAffected != 1 {
				context.HTML(http.StatusOK, MessageHtmlPath, gin.H{
					"message": "你正在连接的网络不受当前认证服务器管辖",
				})
				return
			}
			// 更新连接信息
			var token = uuid.New().String()
			result = db.Exec(
				"INSERT INTO conn(token, user_id, net_id, ip, mac) VALUES (?, ?, ?, ?, ?)",
				token, userId, netId, ip, mac,
			)
			if result.Error != nil || result.RowsAffected != 1 {
				context.Status(http.StatusInternalServerError)
			}
			// 成功重定向
			context.Redirect(http.StatusFound, fmt.Sprintf("http%s://%s:%s/wifidog/auth?token=%s", "",
				gwAddress, gwPort, token))
		})

	// 面向user，成功登录页面
	r.Handle("GET", pathConf.PortalScriptPath,
		func(context *gin.Context) {
			// 这个请求是wifidog重定向给用户的，本质是用户请求的，不用对其身份验证
			var (
				gwId = context.Query("gw_id")
			)
			context.HTML(http.StatusOK, PortalHtmlPath, gin.H{
				"gw_id": gwId,
			})
		})

	// 面向user，提示信息
	r.Handle("GET", pathConf.MsgScriptPath,
		func(context *gin.Context) {
			// 这个请求是wifidog重定向给用户的，本质是用户请求的，不用对其身份验证
			var (
				message = context.Query("message")
			)
			// denied
			context.HTML(http.StatusOK, MessageHtmlPath, gin.H{
				"message": message,
			})
		})

	// 面向Wifidog, Ping
	r.Handle("GET", pathConf.PingScriptPath,
		func(context *gin.Context) {
			// 需要防止外部请求这个接口导致外部修改系统信息 todo
			var (
				gwId          = context.Query("gw_id")
				sysUptime     = context.Query("sys_uptime")
				sysMemfree    = context.Query("sys_memfree")
				sysLoad       = context.Query("sys_load")
				wifidogUptime = context.Query("wifidog_uptime")
			)
			var result *gorm.DB
			// 查询网络是否存在，注意address如果采用别的看门狗可能不一定是ip（至少wifidog是ip）
			var netId int
			result = db.Raw(
				"SELECT id FROM net where sid = ?",
				gwId,
			).Scan(&netId)
			if result.Error != nil || result.RowsAffected != 1 {
				// 不回应非当前网络的请求
				context.Status(http.StatusInternalServerError)
				return
			}
			// 更新网络信息，忽略更新失败的情况
			db.Exec(
				"UPDATE net SET sys_uptime = ?, sys_memfree = ?, sys_load = ?, wifidog_uptime = ? WHERE id = ?",
				sysUptime, sysMemfree, sysLoad, wifidogUptime, netId,
			)
			context.String(http.StatusOK, "Pong")
		})

	// 面向Wifidog, 验证Auth
	r.Handle("GET", pathConf.AuthScriptPath,
		func(context *gin.Context) {
			var (
				stage    = context.Query("stage")
				ip       = context.Query("ip")
				mac      = context.Query("mac")
				token    = context.Query("token")
				incoming = context.Query("incoming")
				outgoing = context.Query("outgoing")
				gwId     = context.Query("gw_id")
			)
			var result *gorm.DB
			// 查询连接
			// 用户是可以拿到token的，为了防止用户在多台设备使用相同mac，这里条件要加上mac
			// 可以加上ip，伪造的可能性更小，但是如果切换vpn可能会导致断开
			var connId int
			result = db.Raw(`
				SELECT conn.id FROM conn 
				LEFT JOIN net ON net.id = conn.net_id 
				WHERE conn.token = ? and net.sid = ? and conn.ip = ? and conn.mac = ? and conn.is_expire = 0
			`,
				token, gwId, ip, mac,
			).Scan(&connId)
			if result.Error != nil {
				context.Status(http.StatusInternalServerError)
				return
			}
			if result.RowsAffected != 1 {
				context.String(http.StatusOK, "Auth: 0")
				return
			}

			// 当前时间戳（秒）
			var timestamp = time.Now().Unix()
			// 更新连接信息，忽略更新失败的情况
			if stage == "login" {
				result = db.Exec(
					"UPDATE conn SET incoming = ?, outgoing = ?, start_time = ?, end_time = ? WHERE id = ?",
					incoming, outgoing, timestamp, timestamp, connId,
				)
				if result.Error != nil {
					context.Status(http.StatusInternalServerError)
					return
				}
				if result.RowsAffected == 1 {
					// 认证成功
					context.String(http.StatusOK, "Auth: 1")
				} else {
					context.String(http.StatusOK, "Auth: 0")
				}
				return
			} else if stage == "counters" {
				result = db.Exec(
					"UPDATE conn SET incoming = ?, outgoing = ?, end_time = ? WHERE id = ?",
					incoming, outgoing, timestamp, connId,
				)
				if result.Error != nil {
					context.Status(http.StatusInternalServerError)
					return
				}
				if result.RowsAffected == 1 {
					// 认证成功
					context.String(http.StatusOK, "Auth: 1")
				} else {
					context.String(http.StatusOK, "Auth: 0")
				}
				return
			} else if stage == "logout" {
				// 退出
				db.Exec(
					"UPDATE conn SET is_expire = 1 WHERE id = ?",
					connId,
				)
				context.String(http.StatusOK, "Auth: 0")
				return
			} else {
				// 不存在当前的stage
				context.String(http.StatusOK, "Auth: 0")
				return
			}
		})

	// 启动服务，监听端口
	err = r.Run(ListenHost + ":" + strconv.Itoa(ListenPort))
	if err != nil {
		log.Fatal("服务端启动失败")
	}
}
