package main

import (
	bot_api_models "github.com/GLGDLY/mhy_botsdk/api_models"
	bot_apis "github.com/GLGDLY/mhy_botsdk/apis"
	bot_events "github.com/GLGDLY/mhy_botsdk/events"
	bot_logger "github.com/GLGDLY/mhy_botsdk/logger"
	bot_plugins "github.com/GLGDLY/mhy_botsdk/plugins"
)

func command1(data bot_events.EventSendMessage, api *bot_apis.ApiBase, logger bot_logger.LoggerInterface) {
	logger.Info("plugin1::command1")
	reply, _ := bot_api_models.NewMsg(bot_api_models.MsgTypeText)
	reply.SetText("plugin1::command1")
	logger.Info(data.Reply(reply))
}

func command2(data bot_events.EventSendMessage, api *bot_apis.ApiBase, logger bot_logger.LoggerInterface) {
	logger.Info("plugin1::command2")
	reply, _ := bot_api_models.NewMsg(bot_api_models.MsgTypeText)
	reply.SetText("plugin1::command2")
	logger.Info(data.Reply(reply))
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
