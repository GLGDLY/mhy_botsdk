package main

import (
	"strings"

	bot_api_models "github.com/GLGDLY/mhy_botsdk/api_models"
	bot_base "github.com/GLGDLY/mhy_botsdk/bot"
	bot_commands "github.com/GLGDLY/mhy_botsdk/commands"
	bot_events "github.com/GLGDLY/mhy_botsdk/events"
	bot_logger "github.com/GLGDLY/mhy_botsdk/logger"
)

var bot_id1 = "bot_id1"
var bot_id2 = "bot_id2"

// SDK支持同端口同路径运行多个机器人，会根据事件的机器人ID判断并分发消息
var test_bot = bot_base.NewBot(bot_id1, "bot_secret1", "/", ":8888") // test bot
var bot = bot_base.NewBot(bot_id2, "bot_secret2", "/", ":8888")

func logger_dispatcher(data bot_events.EventSendMessage) bot_logger.LoggerInterface {
	switch data.Robot.Template.Id {
	case bot_id1:
		return test_bot.Logger
	case bot_id2:
		return bot.Logger
	}
	return bot_logger.NewDefaultLogger("default")
}

func msg_preprocessor(data bot_events.EventSendMessage) {
	switch data.Robot.Template.Id {
	case bot_id1:
		logger_dispatcher(data).Info("收到来自 " + data.Data.Nickname + " 的消息：" + data.GetContent(true))
	case bot_id2:
		logger_dispatcher(data).Info("收到来自 " + data.Data.Nickname + " 的消息：" + data.GetContent(true))
	}
}

func msg_handler(data bot_events.EventSendMessage) {
	// d, _, _ := test_bot.Api.GetVilla(data.Robot.VillaId)
	// logger_dispatcher(data).Info(d.Data.Villa.CategoryID)
	logger_dispatcher(data).Info("default msg handler")
	reply, _ := bot_api_models.NewMsg(bot_api_models.MsgTypeText)
	if strings.Contains(data.GetContent(true), "hello") {
		reply.SetText(
			"Hello World!",
			bot_api_models.MsgEntityMentionUser{
				Text:   "@" + data.Data.Nickname,
				UserID: data.Data.FromUserId,
			},
			"\n",
		)
		reply.AppendText(bot_api_models.MsgEntityVillaRoomLink{
			Text:    "#跳转房间",
			VillaID: data.Robot.VillaId,
			RoomID:  data.Data.RoomId,
		})
		reply.AppendText(bot_api_models.MsgEntityMentionAll{
			Text: "@全体成员",
		})
		logger_dispatcher(data).Info(data.Reply(reply))
	} else {
		reply.SetText("你好，我是机器人，你可以输入 hello 来和我",
			bot_api_models.MsgEntityMentionRobot{
				Text:  "@" + data.Robot.Template.Name,
				BotID: data.Robot.Template.Id,
			}, " 打招呼")
		logger_dispatcher(data).Info(data.Reply(reply))
	}
}

func MyCommand1(data bot_events.EventSendMessage) {
	logger_dispatcher(data).Info("MyCommand1")
	reply, _ := bot_api_models.NewMsg(bot_api_models.MsgTypeImage)
	reply.SetImage("https://webstatic.mihoyo.com/vila/bot/doc/message_api/img/text_case.jpg", 1080, 310, 46000)
	logger_dispatcher(data).Info(data.Reply(reply))
}

func MyCommand2(data bot_events.EventSendMessage) {
	logger_dispatcher(data).Info("MyCommand2")
	reply, _ := bot_api_models.NewMsg(bot_api_models.MsgTypeText)
	reply.SetText("MyCommand2")
	logger_dispatcher(data).Info(data.Reply(reply))
}

func main() {
	test_bot.SetPluginsShortCircuitAffectMain(true)

	/* 以下为添加监听器的方式 */
	test_bot.AddPreprocessor(msg_preprocessor)
	bot.AddPreprocessor(msg_preprocessor)

	test_bot.AddOnCommand(bot_commands.OnCommand{
		Command:        []string{"MyCommand1", "hello world"}, // 命令匹配：包含 "MyCommand1" 或 "hello world" 的消息
		Listener:       MyCommand1,
		RequireAT:      true,
		RequireAdmin:   true,
		IsShortCircuit: true,
	})
	test_bot.AddOnCommand(bot_commands.OnCommand{
		Regex:          "/?MyCommand2", // 正则匹配：内容为 "MyCommand2" 或 "/MyCommand2" 的消息
		Listener:       MyCommand2,
		RequireAT:      true,
		RequireAdmin:   false,
		IsShortCircuit: true,
	})

	test_bot.AddListenerSendMessage(msg_handler)
	/* NewBot 创建一个机器人实例，bot_id 为机器人的id，bot_secret 为机器人的secret，path 为接收事件的路径（如"/"），addr 为接收事件的地址（如":8888"）；
	 * 机器人实例创建后，需要调用 Run() 方法启动机器人；
	 * 对于消息处理，可以通过 AddPreprocessor() 方法添加预处理器，通过 AddOnCommand() 方法添加命令处理器，通过 AddListener() 方法添加事件监听器；
	 * 对于插件，可以通过 AddPlugin() 方法添加插件；
	 * 整体消息处理的运行与短路顺序为： [main]预处理器 -> [插件]预处理器 -> [插件]令处理器 -> [main]命令处理器 -> [main]事件监听器；
	 * 如以上例子，如输入"hello world"，将会执行MyCommand1，然后短路，不执行msg_handler的"hello"指令；而如果输入"hello 123"，则会执行msg_handler的"hello"指令 */

	// fmt.Printf("%+v\n", test_bot)
	bot_base.StartAllBot()
}
