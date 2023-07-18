package main

import (
	"fmt"
	"strings"

	bot_api_models "github.com/GLGDLY/mhy_botsdk/api_models"
	bot_base "github.com/GLGDLY/mhy_botsdk/bot"
	bot_commands "github.com/GLGDLY/mhy_botsdk/commands"
	bot_events "github.com/GLGDLY/mhy_botsdk/events"
)

// NewBot参数: id, secret, 路径, 端口
// 下方例子会监听 localhost:8888/ 获取消息
// 并验证事件的机器人ID是否符合
var bot = bot_base.NewBot("bot_id", "bot_secret", "bot_pubkey", "/", ":8888")

func msg_preprocessor(data bot_events.EventSendMessage) { // 借助preprocessor为所有消息记录log
	bot.Logger.Info("收到来自 " + data.Data.Nickname + " 的消息：" + data.GetContent(true))
}

func MyCommand1(data bot_events.EventSendMessage) {
	bot.Logger.Info("MyCommand1")
	reply, _ := bot_api_models.NewMsg(bot_api_models.MsgTypeImage)                                              // 创建图片类型的消息体
	reply.SetImage("https://webstatic.mihoyo.com/vila/bot/doc/message_api/img/text_case.jpg", 1080, 310, 46000) // 设置图片消息内容
	bot.Logger.Info(data.ReplyCustomize(reply))
}

func MyCommand2(data bot_events.EventSendMessage) {
	bot.Logger.Info("MyCommand2")
	fmt.Print(fmt.Sprintf("MyCommand2 <@%v> <@%v> <@everyone> <#%v>",
		data.Robot.Template.Id, data.Data.FromUserId, data.Data.RoomId))
	bot.Logger.Info(data.Reply(fmt.Sprintf("MyCommand2 <@%v> <@%v> <@everyone> <#%v>",
		data.Robot.Template.Id, data.Data.FromUserId, data.Data.RoomId))) // 使用内嵌格式发送文本消息，内嵌格式按顺序为：@机器人（自己）、@发送者、@全体、#跳转房间
}

func msg_handler(data bot_events.EventSendMessage) { // 最后触发监听器，一般用于确保任何消息都有回复
	bot.Logger.Info("default msg handler")
	reply, _ := bot_api_models.NewMsg(bot_api_models.MsgTypeText) // 创建文本类型的消息体
	if strings.Contains(data.GetContent(true), "hello") {         // 判断消息内容是否包含 "hello"
		reply.SetText("Hello World!",
			bot_api_models.MsgEntityMentionUser{ // 为回复的消息加入@发送者的消息
				Text:   "@" + data.Data.Nickname,
				UserID: data.Data.FromUserId,
			},
			"\n",
		)
		reply.AppendText(bot_api_models.MsgEntityVillaRoomLink{ // 为回复的消息加入大别野房间链接
			Text:    "#跳转房间",
			VillaID: data.Robot.VillaId,
			RoomID:  data.Data.RoomId,
		})
		reply.AppendText(bot_api_models.MsgEntityMentionAll{
			Text: "@全体成员",
		})
		bot.Logger.Info(data.ReplyCustomize(reply))
	} else {
		reply.SetText("你好，我是机器人，你可以输入 hello 来和我",
			bot_api_models.MsgEntityMentionRobot{ // 艾特机器人
				Text:  "@" + data.Robot.Template.Name,
				BotID: data.Robot.Template.Id,
			}, " 打招呼")
		bot.Logger.Info(data.ReplyCustomize(reply))
	}
}

func main() {
	bot.AddPreprocessor(msg_preprocessor)

	bot.AddOnCommand(bot_commands.OnCommand{
		Command:        []string{"MyCommand1", "hello world"}, // 命令匹配：包含 "MyCommand1" 或 "hello world" 的消息
		Listener:       MyCommand1,                            // 设置回调函数为 MyCommand1
		RequireAT:      true,                                  // 是否需要 @ 机器人才能触发
		RequireAdmin:   false,                                 // 是否需要管理员权限才能触发
		IsShortCircuit: true,                                  // 是否短路，即触发后不再继续匹配后续指令和监听器
	})
	bot.AddOnCommand(bot_commands.OnCommand{
		Regex:          "/?MyCommand2", // 正则匹配：内容为 "MyCommand2" 或 "/MyCommand2" 的消息
		Listener:       MyCommand2,
		RequireAT:      true,
		RequireAdmin:   false,
		IsShortCircuit: true,
	})
	bot.AddListenerSendMessage(msg_handler)
	/* NewBot 创建一个机器人实例，bot_id 为机器人的id，bot_secret 为机器人的secret，path 为接收事件的路径（如"/"），addr 为接收事件的地址（如":8888"）；
	 * 机器人实例创建后，需要调用 Run() 方法启动机器人；
	 * 对于消息处理，可以通过 AddPreprocessor() 方法添加预处理器，通过 AddOnCommand() 方法添加命令处理器，通过 AddListener() 方法添加事件监听器；
	 * 对于插件，可以通过 AddPlugin() 方法添加插件；
	 * 整体消息处理的运行与短路顺序为： [main]预处理器 -> [插件]预处理器 -> [插件]令处理器 -> [main]命令处理器 -> [main]事件监听器；
	 * 如以上例子，如输入"hello world"，将会执行MyCommand1，然后短路，不执行msg_handler的"hello"指令；而如果输入"hello 123"，则会执行msg_handler的"hello"指令 */

	bot_base.StartAllBot() // 启动所有机器人
}
