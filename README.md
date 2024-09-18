# Wifidog-Server-Gin

## 描述

本项目是基于openwrt软路由系统中，软件包`wifidog` `luci-app-wifidog`的认证服务器实现

## 使用

```shell
git clone https://github.com/gralliry/Wifidog-Server-Gin.git
cd Wifidog-Server-Gin
go run main.go
```

打开`服务`->`wifodog配置`有几项需要与`config.toml`中对应：

`Port`(当前程序运行的端口，注意在`config.toml`中是字符串)

`通用配置`->`设备ID`(一般是路由器mac地址，对应上你wifidog的配置页面内容即可)

`认证服务器配置`：

* `认证服务器：url路径` -> `/wifidog/`
* `服务器login接口脚本url路径段` -> `login/?`
* `服务器portal接口脚本url路径段` -> `portal/?`
* `服务器ping接口脚本url路径段` -> `ping/?`
* `服务器auth接口脚本url路径段` -> `auth/?`
* `服务器消息接口脚本url路径段` -> `gw_message.php?`

设置你的网关地址`GWAdress`(一般是10.0.0.1)和端口`GWPort`(一般是2060，注意在`config.toml`中是字符串)

## 作者留言

如果是路由器本身作为认证服务器，极力建议使用可执行文件（而不是使用go命令运行源码，其环境和程序占用都过大）

```shell
./wifidog-auth-server
```

如果路由器存在go语言环境，你也可以直接运行源码

```shell
go run main.go
```

自编译

```shell
go env -w GOOS=linux CGO_ENABLED=1
go build -ldflags="-s -w" -o auth-server main.go
```
