# Wifidog-Server-Gin

## 描述

本项目是基于openwrt软路由系统中，软件包`wifidog` `luci-app-wifidog`的认证服务器实现

## 使用

```shell
git clone https://github.com/gralliry/Wifidog-Server-Gin.git
# or `python run.py --port {yourport}`
```

打开`服务`->`wifodog配置`

有几项需要与`app.py`中对应：

`通用配置`->`设备ID`

`认证服务器配置`：
* `认证服务器：url路径`
* `服务器login接口脚本url路径段`
* `服务器portal接口脚本url路径段`
* `服务器ping接口脚本url路径段`
* `服务器auth接口脚本url路径段`
* `服务器消息接口脚本url路径段`

以上参数请一一和`app.py`中的全局变量对应

## 作者留言

如果是外部服务器作为认证服务器，可以采用python
如果是路由器本身作为认证服务器，不建议使用python，其环境和程序占用都过大

可以去看看我另一个用Go写的认证服务器

本项目除去bug修复和说明，不再对功能进行扩展

## 常见问题

* nginx: request header or cookie too large

```
# nginx配置，因为本程序的session是基于cookies加密实现的，cookies的数据量会很大
# nginx对headers大小有限制
server
{
    ...
    # 增加默认请求头部缓冲区大小
    client_header_buffer_size 16k;
    # 增加允许的请求头部缓冲区数量
    large_client_header_buffers 4 16k;
    ...
}
```