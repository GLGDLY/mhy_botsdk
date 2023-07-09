package main

import (
	bot_events "github.com/GLGDLY/mhy_botsdk/events"
	bot_plugins "github.com/GLGDLY/mhy_botsdk/plugins"
)

func command1(data bot_events.EventSendMessage, bot *bot_plugins.AbstractBot) {
	bot.Logger.Info("plugin1::command1")
	bot.Logger.Info(data.Reply("plugin1::command1"))
}

func command2(data bot_events.EventSendMessage, bot *bot_plugins.AbstractBot) {
	bot.Logger.Info("plugin1::command2")
	bot.Logger.Info(data.Reply("plugin1::command2"))
}

func init() {
	bot_plugins.RegisterPlugin( // 注册插件
		"plugin1", // 插件名
		&bot_plugins.Plugin{ // 插件内容
			OnCommand: []bot_plugins.OnCommand{
				{
					Command:        []string{"command1"},
					Listener:       command1,
					RequireAT:      true,
					RequireAdmin:   false,
					IsShortCircuit: true,
				},
				{
					Command:        []string{"command2"},
					Listener:       command2,
					RequireAT:      true,
					RequireAdmin:   false,
					IsShortCircuit: true,
				},
			},
		},
	)
}
