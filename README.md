<p align="right">
   <strong>中文</strong> 
</p>
<div align="center">

# boodlebox2api

_觉得有点意思的话 别忘了点个 ⭐_

<a href="https://t.me/+LGKwlC_xa-E5ZDk9">
    <img src="https://telegram.org/img/website_icon.svg" width="16" height="16" style="vertical-align: middle;">
    <span style="text-decoration: none; font-size: 12px; color: #0088cc; vertical-align: middle;">Telegram 交流群</span>
</a>

<sup><i>(原`coze-discord-proxy`交流群, 此项目仍可进此群**交流** / **反馈bug**)</i></sup>
<sup><i>(群内提供公益API、AI机器人)</i></sup>

</div>

## 功能

- [x] 支持对话接口(流式/非流式)(`/chat/completions`),详情查看[支持模型](#支持模型)
- [x] 支持识别**图片**多轮对话
- [x] 支持自定义请求头校验值(Authorization)
- [x] 支持cookie池(随机),详情查看[获取cookie](#cookie获取方式)
- [x] 支持请求失败自动切换cookie重试(需配置cookie池)
- [x] 可配置代理请求(环境变量`PROXY_URL`)

### 接口文档:

略

### 示例:

略

## 如何使用

略

## 如何集成NextChat

略

## 如何集成one-api

略

## 部署

### 基于 Docker-Compose(All In One) 进行部署

```shell
docker-compose pull && docker-compose up -d
```

#### docker-compose.yml

```docker
version: '3.4'

services:
  boodlebox2api:
    image: deanxv/boodlebox2api:latest
    container_name: boodlebox2api
    restart: always
    ports:
      - "10066:10066"
    volumes:
      - ./data:/app/boodlebox2api/data
    environment:
      - BB_COOKIE=******  # cookie (多个请以,分隔)
      - API_SECRET=123456  # [可选]接口密钥-修改此行为请求头校验的值(多个请以,分隔)
      - TZ=Asia/Shanghai
```

### 基于 Docker 进行部署

```docker
docker run --name boodlebox2api -d --restart always \
-p 10066:10066 \
-v $(pwd)/data:/app/boodlebox2api/data \
-e BB_COOKIE=***** \
-e API_SECRET="123456" \
-e TZ=Asia/Shanghai \
deanxv/boodlebox2api
```

其中`API_SECRET`、`BB_COOKIE`修改为自己的。

如果上面的镜像无法拉取,可以尝试使用 GitHub 的 Docker 镜像,将上面的`deanxv/boodlebox2api`替换为
`ghcr.io/deanxv/boodlebox2api`即可。

### 部署到第三方平台

<details>
<summary><strong>部署到 Zeabur</strong></summary>
<div>

[![Deployed on Zeabur](https://zeabur.com/deployed-on-zeabur-dark.svg)](https://zeabur.com?referralCode=deanxv&utm_source=deanxv)

> Zeabur 的服务器在国外,自动解决了网络的问题,~~同时免费的额度也足够个人使用~~

1. 首先 **fork** 一份代码。
2. 进入 [Zeabur](https://zeabur.com?referralCode=deanxv),使用github登录,进入控制台。
3. 在 Service -> Add Service,选择 Git（第一次使用需要先授权）,选择你 fork 的仓库。
4. Deploy 会自动开始,先取消。
5. 添加环境变量

   `BB_COOKIE:******`  cookie (多个请以,分隔)

   `API_SECRET:123456` [可选]接口密钥-修改此行为请求头校验的值(多个请以,分隔)(与openai-API-KEY用法一致)

保存。

6. 选择 Redeploy。

</div>


</details>

<details>
<summary><strong>部署到 Render</strong></summary>
<div>

> Render 提供免费额度,绑卡后可以进一步提升额度

Render 可以直接部署 docker 镜像,不需要 fork 仓库：[Render](https://dashboard.render.com)

</div>
</details>

## 配置

### 环境变量

1. `PORT=10066`  [可选]端口,默认为10066
2. `DEBUG=true`  [可选]DEBUG模式,可打印更多信息[true:打开、false:关闭]
3. `API_SECRET=123456`  [可选]接口密钥-修改此行为请求头(Authorization)校验的值(同API-KEY)(多个请以,分隔)
4. `BB_COOKIE=******`  cookie (多个请以,分隔)
5. `REQUEST_RATE_LIMIT=60`  [可选]每分钟下的单ip请求速率限制,默认:60次/min
6. `PROXY_URL=http://127.0.0.1:10801`  [可选]代理
6. `USER_AGENT=Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome`  [可选]请求标识,用自己的(可能)防封,默认使用作者的。
7. `ROUTE_PREFIX=hf`  [可选]路由前缀,默认为空,添加该变量后的接口示例:`/hf/v1/chat/completions`
8. `RATE_LIMIT_COOKIE_LOCK_DURATION=600`  [可选]到达速率限制的cookie禁用时间,默认为60s

### cookie获取方式

1. 打开[kilocode](https://kilocode.ai/profile)。
2. 打开**F12**开发者工具。
3. 使用Google登录
<span><img src="docs/img.png" width="800"/></span>
4. 右侧开发者工具-控制台，执行如下代码。

```
document.querySelector('a[href^="vscode://kilocode.kilo-code/kilocode?token="]').href.split('token=')[1]
```

5. 打印的值即所需cookie值,即环境变量`BB_COOKIE`。
<span><img src="docs/img2.png" width="800"/></span>

## 进阶配置

略

## 支持模型

> 新用户绑卡即可获赠 $20 使用额度。如未自动到账，请前往[官方 Discord 服务器](https://discord.com/invite/Ja6BkfyTzJ)，在[指定频道](https://discord.com/channels/1349288496988160052/1355527689272033391)提交注册邮箱，审核通过后额度将添加至账户。

| 模型名称                                | 
|-------------------------------------|
| claude-3-7-sonnet-20250219          |
| claude-3-7-sonnet-20250219-thinking |
| gemini-2.5-pro-preview-03-25        |
| gpt-4.1                             |

> 可通过下图获取并保存`__Secure-next-auth.session-token`随时在[kilo查询平台](https://kl.aytsao.cn/)查询余额。

<span><img src="docs/img1.png" width="400"/></span>

<span><img src="docs/img3.png" width="400"/></span>


## 报错排查

略

## 其他

略