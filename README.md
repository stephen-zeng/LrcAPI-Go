# lrcAPI - Go
为[Stream Music](https://music.aqzscn.cn/)音流写的一个基于Go的歌词获取工具。特点
+ 有缓存机制
+ Go编写，速度嘎嘎快
+ 自动翻译，双语歌词

# 启动参数
+ `--port xxxx` - 设定端口为xxxx
+ `--pwd xxxxxxx` - 设定验证密码为xxxxxxx

注意，Linux需要安装`musl`运行环境，请自行谷歌

# 使用方法和效果
演示是没有命中缓存的效果，命中之后更快
<div style="display: flex; justify-content: center;align-items: center;">
  <img src="https://raw.githubusercontent.com/stephen-zeng/img/master/20250409232913.png">
  <img src="https://gh.qwqwq.com.cn/stephen-zeng/img/master/e1d6503ad000a2fd92f71f3afce7d059.gif">
</div>

# Docker部署
详见衍生项目[LrcAPI-Go-main](https://github.com/zhumao520/LrcAPI-Go-main)，我也不懂这哥们为什么不PR，PR挺好的😅
