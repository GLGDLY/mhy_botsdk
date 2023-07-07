package plugins

import (
	apis "github.com/GLGDLY/mhy_botsdk/apis"
	events "github.com/GLGDLY/mhy_botsdk/events"
	logger "github.com/GLGDLY/mhy_botsdk/logger"
	models "github.com/GLGDLY/mhy_botsdk/models"
)

// 用于为插件提供基础机器人功能的抽象类
type AbstractBot struct {
	Api            *apis.ApiBase
	Logger         logger.LoggerInterface
	WaitForCommand func(reg models.WaitForCommandRegister) (*events.EventSendMessage, error)
}
