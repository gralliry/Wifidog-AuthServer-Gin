# Wifidog-Server-Gin

## 描述

本项目是基于openwrt软路由系统中，软件包`wifidog` `luci-app-wifidog`的认证服务器实现

## 安装

```shell
git clone https://github.com/gralliry/Wifidog-Server-Gin.git
cd Wifidog-Server-Gin

# GOOS和GOARCH对应关系：https://freshman.tech/snippets/go/cross-compile-go-programs/
# gin运行模式：debug | release | test
```

### Linux

```shell
env GOOS=linux GOARCH=amd64 GIN_MODE=debug go build -o authserver
```

### windows

```shell
set GOOS=windows
set GOARCH=amd64
set GIN_MODE=debug
go build -o authserver.exe
```
### darwin(MacOS)

```shell
env GOOS=darwin GOARCH=amd64 GIN_MODE=debug go build -o authserver
```

### openwrt(linux-mipsel)

```shell
env GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -ldflags="-s -w" -o authserver
```

## 使用

打开`openwrt`的`服务`->`wifodog配置`

有几项需要与`config.toml`中对应：

* `ListenHost`对应`认证服务器：主机名`
* `ListenPort`对应`认证服务器：web服务端口` (注意在`config.toml`中是字符串)

然后执行以下
```shell
sqlite3 ./data/database.db
```
```sqlite
-- `通用配置`->`设备ID`(一般是路由器mac地址，对应上你wifidog的配置页面内容即可)
INSERT INTO net(id, address, port) VALUES ('设备ID', '认证服务器：主机名', '认证服务器：web服务端口')
```

打开`认证服务器配置`：

* `认证服务器：url路径` -> `/wifidog/`
* `服务器login接口脚本url路径段` -> `login/?`
* `服务器portal接口脚本url路径段` -> `portal/?`
* `服务器ping接口脚本url路径段` -> `ping/?`
* `服务器auth接口脚本url路径段` -> `auth/?`
* `服务器消息接口脚本url路径段` -> `gw_message.php?`

注意：在`config.toml`不需要添加`?`

## 作者留言

如果是路由器本身作为认证服务器，极力建议使用可执行文件（而不是使用go命令运行源码，其环境和程序占用都过大）

```shell
./authserver
```

如果路由器存在go语言环境，你也可以直接运行源码

```shell
go run main.go
```

## 问题

部分存在无法编译的问题可能是因为缺少对应的gcc库，尤其是对于openwrt中linux-mipsel架构

使用官方的sqlite3驱动是依赖CGO的，不适合低存储低内存的场景，这里使用了其他的sqlite驱动，但是该驱动并不支持`mipsel`的架构

目前作者在寻找适配的、能快速部署的gcc编译器和sqlite驱动，如果你有好的想法可以在issue中提出建议，作者会一一回复