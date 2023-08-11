# Example 5: 基于反向代理的实例

- 此示例主要基于example 4的基础上，改为了使用ws的反向代理，整体架构并没有改变。

主要更改：

`bot_base.NewBot` -> `bot_base.NewWsBot`

## 如何架设反向代理？

### 添加机器人内部的反向代理

在机器人内部添加反向代理，需要在 `bot_base.NewBot()` 或 `bot_base.NewWsBot()` 创建了一个bot后，调用 `bot.AddReverseProxyHTTP()` 或 `bot.AddReverseProxyWS()` 方法。

### 使用 [mhy_glitch_proxy](https://github.com/GLGDLY/mhy_glitch_proxy) 项目

该项目为基于 Glitch 的一个简单ws反代项目，可参考该项目的README.md配置使用。后在创建机器人时调用 `bot_base.NewWsBot()` 方法即可，如：

```go
var bot = bot_base.NewWsBot("bot_id", "bot_secret", "bot_pubkey", "your_ws_uri")
```

> 注意 `your_ws_uri` 为你的反向代理的ws地址，如你在 Glitch 获取的url为 `https://xxx.glitch.me`，请在 `your_ws_uri` 中填写 `ws://xxx.glitch.me/ws/`。

## 反代的优势

- 可以在官方只支持webhook的情况下，反向代理ws，允许开发者本地链接机器人并调试
- 允许在大消息量时实现负载均衡看
- 使用 Glitch 时，可以无需自行购买服务器或架设内网穿透，可使用 Glitch 反代ws到本地，本地机器直接运行机器人。
