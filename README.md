# lrcAPI - Go
为Stream Music音流写的一个基于Go的歌词获取工具。特点
+ 有缓存机制
+ Go编写，速度嘎嘎快

# 启动参数
+ `--port xxxx` - 设定端口为xxxx
+ `--pwd xxxxxxx` - 设定验证密码为xxxxxxx

注意，Linux需要安装`musl`运行环境，请自行谷歌

# 使用参数，参照截图
![](https://raw.githubusercontent.com/stephen-zeng/img/master/20250409232913.png)

# Docker部署
> 感谢[zhumao520](https://github.com/zhumao520)提醒构建docker镜像

可以使用项目中的`Dockerfile`自行构建，也可以使用下面的命令来运行
```bash
docker run -d --name lrcapi -e PWD=%AUTH_PWD% -p %EXPOSE_PORT%:1111 -v %LOCAL_DATA_PLACE%:/app/data 0w0w0/lrcapi-go:latest

# e.g
docker run -d --name lrcapi -e PWD=123456 -p 8080:1111 -v /home/stephenzeng/dockerData/lrcAPI:/app/data 0w0w0/lrcapi-go:latest
```
+ 镜像目前`latest`和具体版本号两种tag，建议使用`latest`。
+ arm版本的镜像为`0w0w0/lrcapi-go-arm`