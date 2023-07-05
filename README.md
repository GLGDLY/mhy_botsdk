# mhy_botsdk

<div align="center">

**一款米哈游大别野机器人的Go SDK**

[![Language](https://img.shields.io/badge/language-go-green.svg?style=plastic)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-orange.svg?style=plastic)](https://github.com/GLGDLY/mhy_botsdk/blob/master/LICENSE)
[![Go](https://img.shields.io/github/v/tag/GLGDLY/mhy_botsdk.svg?style=plastic)](https://pkg.go.dev/github.com/GLGDLY/mhy_botsdk)

</div>

- 基本完善所有事件和API，并支持同时运行多个实例（支持同端口、同路径运行多个拥有不同监听器的机器人）
- 特别针对消息类型事件，配有OnCommand、Preprocessor、Reply等拓展处理器
- 具备Plugins模块，允许使用外部模块直接编写应用

*由于官方文档与实际存在不少差异，目前并不能确保所有消息事件和API完全正确*

## 简易使用

```go
package main

import (
    "strings"

    bot_api_models "github.com/GLGDLY/mhy_botsdk/api_models"
    bot_base "github.com/GLGDLY/mhy_botsdk/bot"
    bot_commands "github.com/GLGDLY/mhy_botsdk/commands"
    bot_events "github.com/GLGDLY/mhy_botsdk/events"
)

// NewBot参数: id, secret, 路径, 端口
// 下方例子会监听 localhost:8888/ 获取消息
// 并验证事件的机器人ID是否符合
var bot = bot_base.NewBot("bot_id", "bot_secret", "/", ":8888")

func msg_preprocessor(data bot_events.EventSendMessage) { // 借助preprocessor为所有消息记录log
    bot.Logger.Info("收到来自 " + data.Data.Nickname + " 的消息：" + data.GetContent(true))
}

func MyCommand1(data bot_events.EventSendMessage) {
    bot.Logger.Info("MyCommand1")
    reply, _ := bot_api_models.NewMsg(bot_api_models.MsgTypeImage)  // 创建图片类型的消息体
    reply.SetImage("https://webstatic.mihoyo.com/vila/bot/doc/message_api/img/text_case.jpg", 1080, 310, 46000) // 设置图片消息内容
    bot.Logger.Info(data.Reply(reply))
}

func MyCommand2(data bot_events.EventSendMessage) {
    bot.Logger.Info("MyCommand2")
    reply, _ := bot_api_models.NewMsg(bot_api_models.MsgTypeText) // 创建文本类型的消息体
    reply.SetText("MyCommand2") // 设置文本消息内容
    bot.Logger.Info(data.Reply(reply))
}

func msg_handler(data bot_events.EventSendMessage) { // 最后触发监听器，一般用于确保任何消息都有回复
    bot.Logger.Info("default msg handler")
    reply, _ := bot_api_models.NewMsg(bot_api_models.MsgTypeText)
    if strings.Contains(data.GetContent(true), "hello") { // 判断消息内容是否包含 "hello"
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
        bot.Logger.Info(data.Reply(reply))
    } else {
        reply.SetText("你好，我是机器人，你可以输入 hello 来和我",
            bot_api_models.MsgEntityMentionRobot{ // 艾特机器人
                Text:  "@" + data.Robot.Template.Name ,
                BotID: data.Robot.Template.Id,
            }, " 打招呼")
        bot.Logger.Info(data.Reply(reply))
    }
}

func main() {
    bot.AddPreprocessor(msg_preprocessor)
    
    bot.AddOnCommand(bot_commands.OnCommand{
        Command:        []string{"MyCommand1", "hello world"}, // 命令匹配：包含 "MyCommand1" 或 "hello world" 的消息
        Listener:       MyCommand1, // 设置回调函数为 MyCommand1
        RequireAT:      true, // 是否需要 @ 机器人才能触发
        RequireAdmin:   false, // 是否需要管理员权限才能触发
        IsShortCircuit: true, // 是否短路，即触发后不再继续匹配后续指令和监听器
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
```

## 消息结构

- SDK的事件类型模型存放在"github.com/GLGDLY/mhy_botsdk/events"中
- 事件类型EventType分为6种event：`JoinVilla`,`SendMessage`,`CreateRobot`,`DeleteRobot`,`AddQuickEmoticon`,`AuditCallback`
  - `AddListener`的注册监听器也相应分为了6种：`AddListenerJoinVilla`,`AddListenerSendMessage`,`AddListenerCreateRobot`,`AddListenerDeleteRobot`,`AddListenerAddQuickEmoticon`,`AddListenerAuditCallback`
- 事件数据结构Event细分成6个子事件：`EventJoinVilla`,`EventSendMessage`,`EventCreateRobot`,`EventDeleteRobot`,`EventAddQuickEmoticon`,`EventAuditCallback`
  - 事件数据结构将作为参数传入注册的事件回调函数
- 由于本SDK针对不同事件，设置了不同的消息监听器函数接口，我们得以减少官方事件数据中extend_data的“套娃”设计，事件的数据结构，原来的`Event`->`extend_data`->`event_data`->`JoinVilla`/`SendMessage`....将简化为`Event`->`Data`，`Data`下直接包含各个事件的扩展数据

## API

- API基本遵从官方API的结构，但存在特例：
  - `SendMessage`(`EventSendMessage`中`Reply`为对其的包装器)：最后一个msg参数要求使用"github.com/GLGDLY/mhy_botsdk/api_models"中的`NewMsg`构造并传入
    - `NewMsg`需要传入`MsgTypeText`, `MsgTypeImage`, `MsgTypePost`之一指定类型
    - `NewMsg`会返回一个`MsgInputModel`结构，其中包含仅限`MsgTypeText`的方法：`AppendText`, `SetText`, `SetTextQuote`；仅限`MsgTypeImage`的方法： `SetImage`；仅限`MsgTypePost`的方法：`SetPost`
    - 这种设计模式是为了分段式内部处理`entities`，方便用户无需执行配置消息json序列

  - `Audit`：最后一个参数要求传入"github.com/GLGDLY/mhy_botsdk/api_models"中的`UserInputAudit`结构体
    - 方便处理可选参数

## 简易插件编写

- 插件的OnCommand回调函数会增加两个参数，以使用当前机器人的api和logger
- 以下分为两个文件，其中`plugin1.go`为介绍插件的编写，`my_bot.go`为介绍加载插件

### plugin1.go：编写插件

```go
package plugin1

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
```

### my_bot.go

```go
package main

import (
    bot_base "github.com/GLGDLY/mhy_botsdk/bot"
    _ "path_to_plugin1" // 使用import导入plugin1
)

// 创建NewBot时会自动加载import了的插件
var bot = bot_base.NewBot("bot_id", "bot_secret", "/", ":8888")

func main() {
    bot.SetPluginsShortCircuitAffectMain(true) // 设置插件的短路是否影响main中注册的指令和消息处理器
    bot_base.StartAllBot() // 启动所有机器人
}
```