package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
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
}

func routePath(rootPath string, scriptPath string) string {
	fullPath := rootPath + scriptPath
	return fullPath[:len(fullPath)-1]
}

func main() {
	var conf Config
	_, err := toml.DecodeFile("config.toml", &conf)
	if err != nil {
		log.Fatal("配置读取失败")
	}
	// 设置为Release模式，这将禁用Logger和Recovery中间件, 但是会关闭所有日志输出
	// 应该由编译或运行时指定
	//gin.SetMode(gin.ReleaseMode)
	//gin.SetMode(gin.DebugMode)
	// 创建一个新的Gin引擎实例
	var r = gin.New()
	// 设置模板文件的路径
	r.LoadHTMLGlob("./pages/*.html")

	// 连接到 SQLite 数据库文件
	conn, err := sqlite.OpenConn("./database.db")
	if err != nil {
		log.Fatal("连接数据库时失败")
	}
	// 确保在函数退出时关闭数据库连接
	defer func(conn *sqlite.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("数据库关闭时发生了错误")
		}
	}(conn)

	// 面向user，登录页面
	r.Handle("GET", routePath(conf.Path, conf.LoginScriptPathFragment), func(context *gin.Context) {
		// 网关信息
		var (
			gwAddress = context.Query("gw_address")
			gwPort    = context.Query("gw_port")
			gwId      = context.Query("gw_id")
		)
		context.HTML(http.StatusOK, "login.html", gin.H{
			"address": gwAddress,
			"port":    gwPort,
			"id":      gwId,
		})
	})

	// 面向user，登录请求
	r.Handle("POST", routePath(conf.Path, conf.LoginScriptPathFragment), func(context *gin.Context) {
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
			// // 因为后台有可能存在应用是使用http发送请求，所以这里的url不一定是用户打开浏览器访问的url
			url = context.Query("url")
		)
		var isExist bool
		// 查询用户是否存在
		var userId int
		isExist = false
		err := sqlitex.ExecuteTransient(conn,
			"SELECT id FROM user_info where username = ? and password = ? LIMIT 1",
			&sqlitex.ExecOptions{
				Args: []any{username, password},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					userId = stmt.ColumnInt(0)
					isExist = true
					return nil
				},
			},
		)
		if err != nil {
			context.Status(http.StatusInternalServerError)
			return
		}
		if !isExist {
			context.HTML(http.StatusUnauthorized, "message.html", gin.H{
				"message": "账号不存在或密码错误",
			})
			return
		}
		// 查询网络是否存在(可以分开两个，查询是否存在再查询是否匹配)
		var netId string
		isExist = false
		err = sqlitex.ExecuteTransient(conn,
			"SELECT id FROM net_info where address = ? and port = ? and id = ? LIMIT 1",
			&sqlitex.ExecOptions{
				Args: []any{gwAddress, gwPort, gwId},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					userId = stmt.ColumnInt(0)
					isExist = true
					return nil
				},
			},
		)
		if err != nil {
			context.Status(http.StatusInternalServerError)
			return
		}
		if !isExist {
			context.HTML(http.StatusForbidden, "message.html", gin.H{
				"message": "你正在连接的网络不受当前认证服务器管辖",
			})
			return
		}
		// 更新用户信息
		var token = uuid.New().String()
		err = sqlitex.ExecuteTransient(conn,
			"INSERT INTO connection (token, user_id, net_id, ip, mac) VALUES (?, ?, ?, ?, ?)",
			&sqlitex.ExecOptions{
				Args: []any{token, userId, netId, ip, mac},
			},
		)
		if err != nil {
			context.HTML(http.StatusInternalServerError, "message.html", gin.H{
				"message": "登录失败",
			})
			return
		}
		// 设置url的cookies
		context.SetCookie("url", url, 60, "/", "", false, false)
		// 成功重定向
		context.Redirect(http.StatusFound, fmt.Sprintf("http%s://%s:%s/wifidog/auth?token=%s", "", gwAddress, gwPort, token))
	})

	// 面向user，成功登录页面
	r.Handle("GET", routePath(conf.Path, conf.PortalScriptPathFragment), func(context *gin.Context) {
		// 这个请求是wifidog重定向给用户的，本质是用户请求的，不用对其身份验证
		var gwId = context.Query("gw_id")
		// 读取 url 尝试重定向
		var url string
		url, err = context.Cookie("url")
		if err != nil {
			context.Redirect(http.StatusFound, url)
		} else {
			context.HTML(http.StatusOK, "portal.html", gin.H{
				"id": gwId,
			})
		}
	})

	// 面向user，提示信息
	r.Handle("GET", routePath(conf.Path, conf.MsgScriptPathFragment), func(context *gin.Context) {
		// 这个请求是wifidog重定向给用户的，本质是用户请求的，不用对其身份验证
		var message = context.Query("message")
		// denied
		context.HTML(http.StatusOK, "message.html", gin.H{
			"message": message,
		})
	})

	// 面向Wifidog, Ping
	r.Handle("GET", routePath(conf.Path, conf.PingScriptPathFragment), func(context *gin.Context) {
		// 需要防止外部请求这个接口导致外部修改系统信息 todo
		var (
			gwId          = context.Query("gw_id")
			sysUptime     = context.Query("sys_uptime")
			sysMemfree    = context.Query("sys_memfree")
			sysLoad       = context.Query("sys_load")
			wifidogUptime = context.Query("wifidog_uptime")
		)
		// 查询网络是否存在，注意address如果采用别的看门狗可能不一定是ip（至少wifidog是ip）
		var netId string
		var isExist = false
		err = sqlitex.ExecuteTransient(conn,
			"SELECT id FROM net_info where id = ? LIMIT 1",
			&sqlitex.ExecOptions{
				Args: []any{netId},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					netId = stmt.ColumnText(0)
					isExist = true
					return nil
				},
			},
		)
		if err != nil || !isExist {
			// 不回应非当前网络的请求
			context.Status(http.StatusInternalServerError)
			return
		}
		// 更新网络信息，忽略更新失败的情况
		err = sqlitex.ExecuteTransient(conn,
			"UPDATE net_info SET sys_uptime = ?, sys_memfree = ?, sys_load = ?, wifidog_uptime = ? WHERE id = ?",
			&sqlitex.ExecOptions{
				Args: []any{sysUptime, sysMemfree, sysLoad, wifidogUptime, gwId},
			},
		)
		context.String(http.StatusOK, "Pong")
	})

	// 面向Wifidog, 验证Auth
	r.Handle("GET", routePath(conf.Path, conf.AuthScriptPathFragment), func(context *gin.Context) {
		var (
			stage    = context.Query("stage")
			ip       = context.Query("ip")
			mac      = context.Query("mac")
			token    = context.Query("token")
			incoming = context.Query("incoming")
			outgoing = context.Query("outgoing")
			gwId     = context.Query("gw_id")
		)
		// 查询连接
		// 用户是可以拿到token的，为了防止用户在多台设备使用相同mac，这里条件要加上mac
		// 可以加上ip，伪造的可能性更小，但是如果切换vpn可能会导致断开
		var connId int
		var isExist = false
		err = sqlitex.ExecuteTransient(conn,
			"SELECT id FROM connection where token = ? and net_id = ? and ip = ? and mac = ? LIMIT 1",
			&sqlitex.ExecOptions{
				Args: []any{token, gwId, ip, mac},
				ResultFunc: func(stmt *sqlite.Stmt) error {
					connId = stmt.ColumnInt(0)
					isExist = true
					return nil
				},
			},
		)
		if err != nil {
			context.Status(http.StatusInternalServerError)
			return
		}
		if !isExist {
			context.String(http.StatusOK, "Auth: 0")
			return
		}
		// 当前时间戳（秒）
		var timestamp = time.Now().Unix()
		// 更新连接信息，忽略更新失败的情况
		if stage == "login" {
			err = sqlitex.ExecuteTransient(conn,
				"UPDATE connection SET incoming = ?, outgoing = ?, start_time = ?, end_time = ? WHERE id = ?",
				&sqlitex.ExecOptions{
					Args: []any{incoming, outgoing, timestamp, timestamp, connId},
				},
			)
			// 认证成功
			context.String(http.StatusOK, "Auth: 1")
		} else if stage == "counters" {
			err = sqlitex.ExecuteTransient(conn,
				"UPDATE connection SET incoming = ?, outgoing = ?, end_time = ? WHERE id = ?",
				&sqlitex.ExecOptions{
					Args: []any{incoming, outgoing, timestamp, connId},
				},
			)
			// 认证成功
			context.String(http.StatusOK, "Auth: 1")
		} else if stage == "logout" {
			// 考虑是否要验证ip和mac
			err = sqlitex.ExecuteTransient(conn,
				"DELETE FROM connection WHERE token = ?",
				&sqlitex.ExecOptions{
					Args: []any{token},
				},
			)
			// 退出
			context.String(http.StatusOK, "Auth: 0")
		} else {
			// 不存在当前的stage
			context.String(http.StatusOK, "Auth: 0")
		}
	})

	// 启动服务，监听端口
	err = r.Run(conf.Host + ":" + conf.Port)
	if err != nil {
		log.Fatal("服务端启动失败")
	}
}
