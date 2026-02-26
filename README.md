# lrcAPI - Go
为[Stream Music](https://music.aqzscn.cn/)音流写的一个基于Go的歌词获取工具。特点
+ 有缓存机制（SQLite）
+ Go编写，速度嘎嘎快
+ 自动翻译，双语歌词

# 启动参数
+ `--port xxxx` - 设定端口为xxxx
+ `--pwd xxxxxxx` - 设定验证密码为xxxxxxx

# 使用方法和效果
演示是没有命中缓存的效果，命中之后更快
<div style="display: flex; justify-content: center;align-items: center;">
  <img src="https://raw.githubusercontent.com/stephen-zeng/img/master/20250409232913.png">
  <img src="https://gh.qwqwq.com.cn/stephen-zeng/img/master/e1d6503ad000a2fd92f71f3afce7d059.gif">
</div>

# 使用参数，参照截图
![](https://raw.githubusercontent.com/stephen-zeng/img/master/20250409232913.png)

# Docker部署
> 感谢[zhumao520](https://github.com/zhumao520)提醒构建docker镜像

可以使用项目中的`Dockerfile`自行构建，也可以使用下面的命令来运行
```bash
docker run -d --name lrcapi -e PWD=%AUTH_PWD% -p %EXPOSE_PORT%:1111 -v %LOCAL_DATA_PLACE%:/app/assets 0w0w0/lrcapi-go:latest

# e.g
docker run -d --name lrcapi -e PWD=123456 -p 8080:1111 -v /home/stephenzeng/dockerData/lrcAPI:/app/assets 0w0w0/lrcapi-go:latest
```
+ 镜像目前`latest`和具体版本号两种tag，建议使用`latest`。
+ arm版本的镜像为`0w0w0/lrcapi-go-arm`

# 数据存储
SQLite数据库默认存放在`assets/lyrics.db`，仍然建议将`/app/assets`挂载为持久化目录。

# 二进制部署
可以到action里面去获取最新的二进制版本，也可以去release里面获取稳定的二进制版本

# 注册为systemctl服务
下面是`/etc/systemd/system/lrcAPI.service`模板
```
[Unit]
Description = lrcAPI
After = network.target syslog.target
Wants = network.target

[Service]
Type = simple
WorkingDirectory = /root/lrcAPI/
ExecStart = /root/lrcAPI/lrcAPI --port 1111 --pwd 123456
Restart = on-failure


[Install]
WantedBy = multi-user.target
```
之后运行`systemctl daemon-reload && systemctl enable --now lrcAPI`即可注册为开机启动的服务
