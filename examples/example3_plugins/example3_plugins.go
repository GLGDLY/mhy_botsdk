package main

import (
	bot_base "github.com/GLGDLY/mhy_botsdk/bot"
)

// 创建NewBot时会自动加载import了的插件
var bot = bot_base.NewBot("bot_id", "bot_secret", "/", "bot_pubkey", ":8888")

func main() {
	bot.SetPluginsShortCircuitAffectMain(true)        // 设置插件的短路是否影响main中注册的指令和消息处理器
	bot.Logger.Info("Plugins:", bot.GetPluginNames()) // 获取所有插件名
	// bot.SetPluginEnabled("plugin1", false) 		 // 设置插件是否启用
	bot_base.StartAllBot() // 启动所有机器人
}
