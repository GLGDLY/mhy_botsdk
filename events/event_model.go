package events

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	api_models "github.com/GLGDLY/mhy_botsdk/api_models"
	apis "github.com/GLGDLY/mhy_botsdk/apis"
)

/* --------- enum EventType start --------- */
type EventType uint8

const (
	JoinVilla        EventType = 1
	SendMessage      EventType = 2
	CreateRobot      EventType = 3
	DeleteRobot      EventType = 4
	AddQuickEmoticon EventType = 5
	AuditCallback    EventType = 6
)

/* --------- enum EventType end --------- */

/* event specific */

type JoinVillaData struct {
	JoinUid          uint64 `json:"join_uid"`
	JoinUserNickname string `json:"join_user_nickname"`
	JoinAt           uint64 `json:"join_at"`
	VillaId          uint64 `json:"villa_id"`
}

type SendMessageData struct {
	ContentRaw string `json:"content"`
	Content    struct {
		Trace struct {
			VisualRoomVersion string `json:"visual_room_version"`
			AppVersion        string `json:"app_version"`
			ActionType        uint8  `json:"action_type"`
			BotMsgId          string `json:"bot_msg_id"`
			Client            string `json:"client"`
			Env               string `json:"env"`
			RongSdkVersion    string `json:"rong_sdk_version"`
		} `json:"trace"`
		MentionedInfo struct {
			MentionedContent string   `json:"mentioned_content"`
			UserIdList       []string `json:"user_id_list"`
			Type             uint8    `json:"type"`
		} `json:"mentioned_info"`
		User struct {
			PortraitUri string `json:"portrait_uri"`
			Extra       struct {
				MemberRoles struct {
					Name          string `json:"name"`
					Color         string `json:"color"`
					WebColor      string `json:"web_color"`
					RoleFontColor string `json:"role_font_color"`
					RoleBgColor   string `json:"role_bg_color"`
				} `json:"member_roles"`
				State interface{} `json:"state"`
			} `json:"extra"`
			Name     string `json:"name"`
			Alias    string `json:"alias"`
			Id       string `json:"id"`
			Portrait string `json:"portrait"`
		} `json:"user"`
		Content struct {
			Images []struct { // just guessing its data structure
				Url  string `json:"url"`
				Size struct {
					Width  uint64 `json:"width"`
					Height uint64 `json:"height"`
				} `json:"size"`
				FileSize uint64 `json:"file_size"`
			} `json:"images"`
			Entities []struct {
				Offset uint8 `json:"offset"`
				Length uint8 `json:"length"`
				Entity struct {
					Type  string `json:"type"`
					BotId string `json:"bot_id"`
				} `json:"entity"`
			} `json:"entities"`
			Text string `json:"text"`
		}
	}
	FromUserId uint64 `json:"from_user_id"`
	SendAt     uint64 `json:"send_at"`
	ObjectName uint64 `json:"object_name"`
	RoomId     uint64 `json:"room_id"`
	Nickname   string `json:"nickname"`
	MsgUid     string `json:"msg_uid"`
	BotMsgId   string `json:"bot_msg_id"`
	VillaId    uint64 `json:"villa_id"`
}

type CreateRobotData struct {
	VillaId uint64 `json:"villa_id"`
}

type DeleteRobotData struct {
	VillaId uint64 `json:"villa_id"`
}

type AddQuickEmoticonData struct {
	VillaId    uint64 `json:"villa_id"`
	RoomId     uint64 `json:"room_id"`
	Uid        uint64 `json:"uid"`
	EmoticonId uint16 `json:"emoticon_id"`
	Emoticon   string `json:"emoticon"`
	MsgUid     string `json:"msg_uid"`
	BotMsgId   string `json:"bot_msg_id"`
	IsCancel   bool   `json:"is_cancel"`
}

type AuditCallbackData struct {
	AuditId     string `json:"audit_id"`
	BotTplId    string `json:"bot_tpl_id"`
	VillaId     uint64 `json:"villa_id"`
	RoomId      uint64 `json:"room_id"`
	UserId      uint64 `json:"user_id"`
	PassThrough string `json:"pass_through"`
	AuditResult uint   `json:"audit_result"`
}

/* general */

type Robot struct {
	Template struct {
		Id       string `json:"id"`
		Name     string `json:"name"`
		Desc     string `json:"desc"`
		Icon     string `json:"icon"`
		Commands []struct {
			Name string `json:"name"`
			Desc string `json:"desc"`
		} `json:"commands"`
	} `json:"template"`
	VillaId uint64 `json:"villa_id"`
}

type EventBase struct {
	Robot     Robot     `json:"robot"`
	Type      EventType `json:"type"`
	CreatedAt uint64    `json:"created_at"`
	Id        string    `json:"id"`
	SendAt    uint64    `json:"send_at"`
}

/* event */

type EventJoinVilla struct {
	EventBase
	Data JoinVillaData
}

type EventSendMessage struct {
	EventBase
	Data SendMessageData
	api  *apis.ApiBase // used for reply helper function
}

type EventCreateRobot struct {
	EventBase
	Data CreateRobotData
}

type EventDeleteRobot struct {
	EventBase
	Data DeleteRobotData
}

type EventAddQuickEmoticon struct {
	EventBase
	Data AddQuickEmoticonData
}

type EventAuditCallback struct {
	EventBase
	Data AuditCallbackData
}

type Event struct {
	Event struct {
		EventBase
		ExtendData struct {
			EventData struct {
				JoinVilla        JoinVillaData        `json:"JoinVilla,omitempty"`
				SendMessage      SendMessageData      `json:"SendMessage,omitempty"`
				CreateRobot      CreateRobotData      `json:"CreateRobot,omitempty"`
				DeleteRobot      DeleteRobotData      `json:"DeleteRobot,omitempty"`
				AddQuickEmoticon AddQuickEmoticonData `json:"AddQuickEmoticon,omitempty"`
				AuditCallback    AuditCallbackData    `json:"AuditCallback,omitempty"`
			} `json:"EventData"`
		} `json:"extend_data"`
	} `json:"event"`
	Type string `json:"type,omitempty"` // for sdk reverse proxy packet handling
	Sign string `json:"sign,omitempty"` // for sdk reverse proxy packet handling
}

/* helper functions for EventSendMessage */

// 获取消息内容，is_treat为true时，如果消息内容以/开头，则去掉/
func (e *EventSendMessage) GetContent(is_treat bool) string {
	content := e.Data.Content.Content.Text
	if is_treat {
		at := "@" + e.Robot.Template.Name
		content = strings.Replace(content, at, "", -1)
		content = strings.TrimSpace(content)
		content = strings.TrimLeft(content, "/")
		content = strings.TrimSpace(content)
	}
	return content
}

// 在相应的房间回复消息 i.e. wrapper for api.SendMessage
// 使用内嵌格式发送消息，并自动处理内部Entity（<@xxx>为艾特机器人或用户，<@everyone>为艾特全体，<#xxx>为跳转房间，<$xxx>为跳转连接）
// 艾特用户会自动获取用户昵称，跳转房间会自动获取房间名称；艾特机器人会显示文字“机器人”，艾特全体会显示“全体成员”，跳转连接会显示链接自身
// 使用\< 和 \> 可转义 < 和 >，不会被解析为Entity
func (e *EventSendMessage) Reply(msg ...string) (api_models.SendMessageModel, int, error) {
	return e.api.SendMessage(e.Robot.VillaId, e.Data.RoomId, msg...)
}

// 在相应的房间回复消息 i.e. wrapper for api.SendMessageCustomize
// 使用models.NewMsg创建消息，然后使用models.SetText等方法加入内容，最后使用此函数发送
func (e *EventSendMessage) ReplyCustomize(msg api_models.MsgInputModel) (api_models.SendMessageModel, int, error) {
	return e.api.SendMessageCustomize(e.Robot.VillaId, e.Data.RoomId, msg)
}

/* helpher functions for internal converting */

func Event2EventJoinVilla(event Event) EventJoinVilla {
	var eventJoinVilla EventJoinVilla
	eventJoinVilla.EventBase = event.Event.EventBase
	eventJoinVilla.Data = event.Event.ExtendData.EventData.JoinVilla
	return eventJoinVilla
}

func Event2EventSendMessage(event Event, api *apis.ApiBase) EventSendMessage {
	var eventSendMessage EventSendMessage
	eventSendMessage.EventBase = event.Event.EventBase
	eventSendMessage.Data = event.Event.ExtendData.EventData.SendMessage
	json.Unmarshal([]byte(eventSendMessage.Data.ContentRaw), &eventSendMessage.Data.Content)
	eventSendMessage.api = api
	return eventSendMessage
}

func Event2EventCreateRobot(event Event) EventCreateRobot {
	var eventCreateRobot EventCreateRobot
	eventCreateRobot.EventBase = event.Event.EventBase
	eventCreateRobot.Data = event.Event.ExtendData.EventData.CreateRobot
	return eventCreateRobot
}

func Event2EventDeleteRobot(event Event) EventDeleteRobot {
	var eventDeleteRobot EventDeleteRobot
	eventDeleteRobot.EventBase = event.Event.EventBase
	eventDeleteRobot.Data = event.Event.ExtendData.EventData.DeleteRobot
	return eventDeleteRobot
}

func Event2EventAddQuickEmoticon(event Event) EventAddQuickEmoticon {
	var eventAddQuickEmoticon EventAddQuickEmoticon
	eventAddQuickEmoticon.EventBase = event.Event.EventBase
	eventAddQuickEmoticon.Data = event.Event.ExtendData.EventData.AddQuickEmoticon
	return eventAddQuickEmoticon
}

func Event2EventAuditCallback(event Event) EventAuditCallback {
	var eventAuditCallback EventAuditCallback
	eventAuditCallback.EventBase = event.Event.EventBase
	eventAuditCallback.Data = event.Event.ExtendData.EventData.AuditCallback
	return eventAuditCallback
}

/* listeners for each event */
type BotListenerJoinVilla func(data EventJoinVilla)
type BotListenerSendMessage func(data EventSendMessage)
type BotListenerCreateRobot func(data EventCreateRobot)
type BotListenerDeleteRobot func(data EventDeleteRobot)
type BotListenerAddQuickEmoticon func(data EventAddQuickEmoticon)
type BotListenerAuditCallback func(data EventAuditCallback)

/* raw request listener */
type BotListenerRawRequest func(c *gin.Context)
