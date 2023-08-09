package main

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	bot_base "github.com/GLGDLY/mhy_botsdk/bot"
	bot_commands "github.com/GLGDLY/mhy_botsdk/commands"
	bot_events "github.com/GLGDLY/mhy_botsdk/events"
	bot_models "github.com/GLGDLY/mhy_botsdk/models"
)

var bot = bot_base.NewBot("bot_id", "bot_secret", "bot_pubkey", "/", ":8888")

func msg_preprocessor(data bot_events.EventSendMessage) { // 借助preprocessor为所有消息记录log
	bot.Logger.Info("收到来自 " + data.Data.Nickname + " 的消息：" + data.GetContent(true))
}

func GuessingGame(data bot_events.EventSendMessage) {
	bot.Logger.Info("GuessingGame")

	reply := fmt.Sprintf("<@%v> 猜数字游戏开始，输入 1-100 之间的数字", data.Data.FromUserId)
	bot.Logger.Info(data.Reply(reply))

	var identify string = fmt.Sprintf("guessing_game_%v", data.Data.FromUserId) // 用于标识此次游戏的唯一标识符
	bot.CancelWaitForCommand(identify)                                          // 取消之前的等待指令（如不存在会返回error）

	// 创建目标数字和相应保存的min、max
	min := 1
	max := 100
	target := min + rand.Intn(max-min)
	// 创建一个匹配1-100之间数字的正则表达式
	num_reg := `([1-9]\d?|100)`

	// 等待用户输入
	var timeout time.Duration = 5 * time.Minute // 超时时间
	var new_data *bot_events.EventSendMessage
	var err error
	for {
		new_data, err = bot.WaitForCommand(bot_models.WaitForCommandRegister{
			Scope: bot_models.ScopeVilla | bot_models.ScopeRoom | bot_models.ScopeUser, // 作用域：villa、room、user，代表必须同时满足在当前别野、当前房间、当前用户才能触发
			Command: bot_models.CommandBase{
				Regex:          num_reg,
				RequireAT:      false,
				IsShortCircuit: true,
			},
			Data:     &data,
			Identify: &identify, // 用于标识该注册的字符串，用于验证是否重复或取消注册
			Timeout:  &timeout,  // 超时时间，如果为0则不超时，nil默认60秒
		})
		if err != nil {
			reply := fmt.Sprintf("<@%v>", data.Data.FromUserId)
			switch err.Error() {
			case "timeout":
				reply += "超时了，游戏结束"
			case "cancel":
				reply += "游戏结束"
			default:
				bot.Logger.Error(err)
				reply += "发生错误，游戏结束"
			}
			bot.Logger.Info(data.Reply(reply))
			return
		}
		num, conv_err := strconv.Atoi(regexp.MustCompile(num_reg).FindString(new_data.GetContent(true)))
		if conv_err != nil {
			bot.Logger.Error(conv_err)
			continue
		}
		if target == num {
			reply := fmt.Sprintf("<@%v> 恭喜你猜对了！", data.Data.FromUserId)
			bot.Logger.Info(data.Reply(reply))
			return
		} else if target > num {
			if num > min {
				min = num
			}
			reply := fmt.Sprintf("<@%v> [%v]太小了，范围 %v-%v", data.Data.FromUserId, num, min, max)
			bot.Logger.Info(data.Reply(reply))
		} else {
			if num < max {
				max = num
			}
			reply := fmt.Sprintf("<@%v> [%v]太大了，范围 %v-%v", data.Data.FromUserId, num, min, max)
			bot.Logger.Info(data.Reply(reply))
		}
	}
}

func msg_handler(data bot_events.EventSendMessage) { // 最后触发监听器，一般用于确保任何消息都有回复
	bot.Logger.Info("default msg handler")
	reply := fmt.Sprintf("你好，我是机器人，你可以输入 猜数字 来和我 <@%v> 玩游戏呢", data.Robot.Template.Id)
	bot.Logger.Info(data.Reply(reply))
}

func main() {
	rand.Seed(time.Now().UnixNano())

	bot.AddPreprocessor(msg_preprocessor)

	bot.AddOnCommand(bot_commands.OnCommand{
		Command:        []string{"猜数字"},
		Listener:       GuessingGame,
		RequireAT:      true,
		RequireAdmin:   false,
		IsShortCircuit: true,
	})
	bot.AddListenerSendMessage(msg_handler)

	bot_base.StartAllBot() // 启动所有机器人
}
