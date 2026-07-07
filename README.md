# lrcAPI - Go
为[Stream Music](https://music.aqzscn.cn/)音流写的一个基于Go的歌词获取工具。特点
+ 有缓存机制（SQLite）
+ Go编写，速度嘎嘎快
+ 自动翻译，双语歌词
+ 多来源聚合：**网易云音乐、QQ音乐（官方 QRC 逐词接口）、酷狗音乐（KRC 逐词接口）**
+ 返回逐词歌词的罗马音（`romaji`），并标注歌词来源（`source`）与类型（`type`）

# 歌词来源
| 来源 | `source` | 接口 | 逐词原始格式 | 凭据 |
| --- | --- | --- | --- | --- |
| 网易云音乐 | `netease` | `music.163.com/api/song/lyric`（lv/tv/rv） | LRC | 匿名可用，可选 Cookie |
| QQ音乐 | `qqmusic` | `u.y.qq.com/cgi-bin/musicu.fcg`（`GetPlayLyricInfo`，QRC）→ 回退 `fcg_query_lyric_new.fcg` | QRC（DES 加密） | 匿名可用，可选 Cookie |
| 酷狗音乐 | `kugou` | `krcs.kugou.com/search` + `lyrics.kugou.com/download` | KRC（异或+zlib） | 匿名可用 |

> QQ 的 QRC 使用其特有的「buggy DES」加密，解密依赖 [`github.com/jixunmoe-go/qrc`](https://github.com/jixunmoe-go/qrc)。

# 返回字段
每个候选包含：`id`、`title`、`artist`、`lyrics`（原文+中文翻译混合的 LRC）、`romaji`（罗马音 LRC，可为空）、`type`（`lrc`/`ttml`）、`source`（来源标识）。

# 启动参数与配置
优先级：**命令行参数 > 环境变量 / `.env` > 默认值**。

命令行：
+ `--port xxxx` - 设定端口为xxxx
+ `--pwd xxxxxxx` - 设定验证密码为xxxxxxx

环境变量 / `.env`（二进制直接运行会自动读取同目录 `.env`）：
+ `LRCAPI_PORT` - 端口（等价 `--port`）
+ `LRCAPI_PWD` - 密码（等价 `--pwd`，推荐；`PWD` 仍兼容 docker 旧用法）
+ `NETEASE_COOKIE` / `QQ_COOKIE` / `KUGOU_COOKIE` - 可选，各平台 Cookie
+ 其它预留：`APPLE_DEVELOPER_TOKEN`、`APPLE_MEDIA_USER_TOKEN`、`SPOTIFY_SP_DC`、`SODA_COOKIE`、`BILIBILI_COOKIE`、`YOUTUBE_MUSIC_COOKIE`

完整模板见 [`.env.example`](./.env.example)，复制为 `.env` 后填写即可。

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

## docker-compose部署（推荐）
项目自带 [`docker-compose.yml`](./docker-compose.yml)，配合 `.env` 管理凭据：
```bash
cp .env.example .env   # 按需填写密码 / Cookie
docker compose up -d
```
compose 通过 `env_file: .env` 注入全部环境变量，并把 `./data` 挂载为 SQLite 持久化目录。arm64 主机把 `image` 改为 `0w0w0/lrcapi-go-arm:latest` 即可。

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
