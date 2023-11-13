package api_models

import (
	"bytes"
	"encoding/json"
	"errors"
	"unicode/utf16"

	utils "github.com/GLGDLY/mhy_botsdk/utils"
)

/* user input message struct wrapper */
type MsgContentType string
type MsgMentionType uint64
type MsgEntityType string

const (
	MsgTypeText  MsgContentType = "MHY:Text"
	MsgTypeImage MsgContentType = "MHY:Image"
	MsgTypePost  MsgContentType = "MHY:Post"
)

const (
	MsgEntityMentionRobotType  MsgEntityType = "mentioned_robot"
	MsgEntityMentionUserType   MsgEntityType = "mentioned_user"
	MsgEntityMentionAllType    MsgEntityType = "mention_all"
	MsgEntityVillaRoomLinkType MsgEntityType = "villa_room_link"
	MsgEntityLinkType          MsgEntityType = "link"
)

const (
	MentionAll  MsgMentionType = 1
	MentionUser MsgMentionType = 2
)

type MsgEntityMentionRobot struct {
	Text  string
	BotID string
}

type MsgEntityMentionUser struct {
	Text   string
	UserID uint64
}

type MsgEntityMentionAll struct {
	Text string
}

type MsgEntityVillaRoomLink struct {
	Text    string
	VillaID uint64
	RoomID  uint64
}

type MsgEntityLink struct {
	Text                   string
	URL                    string
	RequiresBotAccessToken bool
}

type MsgInputModel map[string]interface{}

func NewMsg(msg_type MsgContentType) (MsgInputModel, error) {
	msg := MsgInputModel{"object_name": msg_type, "msg_content": MsgInputModel{}}
	switch msg_type {
	case MsgTypeText:
		msg["msg_content"].(MsgInputModel)["content"] = MsgInputModel{"text": bytes.NewBufferString(""), "entities": []MsgInputModel{}}
		msg["msg_content"].(MsgInputModel)["mentionedInfo"] = MsgInputModel{"type": MentionUser, "userIdList": []string{}}
	case MsgTypeImage:
		msg["msg_content"].(MsgInputModel)["content"] = MsgInputModel{"url": ""}
	case MsgTypePost:
		msg["msg_content"] = MsgInputModel{"content": MsgInputModel{"post_id": ""}}
	default:
		return nil, errors.New("不支持的消息类型")
	}
	return msg, nil
}

func (msg MsgInputModel) appendText(text_len int, args ...interface{}) { // internal processor
	msg_content := msg["msg_content"].(MsgInputModel)
	msg_content_inner := msg_content["content"].(MsgInputModel)
	msg_context_text := msg_content_inner["text"].(*bytes.Buffer)
	msg_content_mentionedInfo := msg_content["mentionedInfo"].(MsgInputModel)
	for _, arg := range args {
		switch typed_arg := arg.(type) {
		case string:
			msg_context_text.WriteString(typed_arg)
			text_len += len(utf16.Encode([]rune(typed_arg)))
		case MsgEntityMentionRobot:
			this_len := len(utf16.Encode([]rune(typed_arg.Text)))
			msg_content_inner["entities"] = append(msg_content_inner["entities"].([]MsgInputModel),
				MsgInputModel{"entity": MsgInputModel{"type": MsgEntityMentionRobotType, "bot_id": typed_arg.BotID},
					"offset": text_len, "length": this_len})
			msg_context_text.WriteString(typed_arg.Text)
			msg_content_mentionedInfo["userIdList"] = append(msg_content_mentionedInfo["userIdList"].([]string),
				typed_arg.BotID)
			text_len += this_len
		case MsgEntityMentionUser:
			this_len := len(utf16.Encode([]rune(typed_arg.Text)))
			msg_content_inner["entities"] = append(msg_content_inner["entities"].([]MsgInputModel),
				MsgInputModel{"entity": MsgInputModel{"type": MsgEntityMentionUserType, "user_id": utils.String(typed_arg.UserID)},
					"offset": text_len, "length": this_len})
			msg_context_text.WriteString(typed_arg.Text)
			msg_content_mentionedInfo["userIdList"] = append(msg_content_mentionedInfo["userIdList"].([]string),
				utils.String(typed_arg.UserID))
			text_len += this_len
		case MsgEntityMentionAll:
			this_len := len(utf16.Encode([]rune(typed_arg.Text)))
			msg_content_inner["entities"] = append(msg_content_inner["entities"].([]MsgInputModel),
				MsgInputModel{"entity": MsgInputModel{"type": MsgEntityMentionAllType}, "offset": text_len, "length": this_len})
			msg_context_text.WriteString(typed_arg.Text)
			msg_content_mentionedInfo["type"] = MentionAll
			text_len += this_len
		case MsgEntityVillaRoomLink:
			this_len := len(utf16.Encode([]rune(typed_arg.Text)))
			msg_content_inner["entities"] = append(msg_content_inner["entities"].([]MsgInputModel),
				MsgInputModel{"entity": MsgInputModel{"type": MsgEntityVillaRoomLinkType, "villa_id": utils.String(typed_arg.VillaID),
					"room_id": utils.String(typed_arg.RoomID)},
					"offset": text_len, "length": this_len})
			msg_context_text.WriteString(typed_arg.Text)
			text_len += this_len
		case MsgEntityLink:
			text := typed_arg.Text
			if text == "" {
				text = typed_arg.URL
			}
			this_len := len(utf16.Encode([]rune(text)))
			msg_content_inner["entities"] = append(msg_content_inner["entities"].([]MsgInputModel),
				MsgInputModel{"entity": MsgInputModel{"type": MsgEntityLinkType, "url": typed_arg.URL,
					"requires_bot_access_token": typed_arg.RequiresBotAccessToken},
					"offset": text_len, "length": this_len})
			msg_context_text.WriteString(text)
			text_len += this_len
		default:
			content_string := utils.String(arg)
			msg_context_text.WriteString(content_string)
			text_len += len(utf16.Encode([]rune(content_string)))
		}
	}
}

func (msg MsgInputModel) AppendText(args ...interface{}) error {
	// if MsgContentType(msg["object_name"].(MsgContentType)) != MsgTypeText {
	// 	return errors.New("消息类型不是文本消息，请使用NewMsg(MsgText)创建文本消息后SetText")
	// }
	s, ok := msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(string)
	var text_len int
	if !ok {
		s = msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(*bytes.Buffer).String()
	}
	text_len = len(utf16.Encode([]rune(s)))
	msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"] = bytes.NewBufferString(s)
	msg.appendText(text_len, args...)
	return nil
}

// 设置文本消息内容
func (msg MsgInputModel) SetText(args ...interface{}) error {
	// if MsgContentType(msg["object_name"].(MsgContentType)) != MsgTypeText {
	// 	return errors.New("消息类型不是文本消息，请使用NewMsg(MsgText)创建文本消息后SetText")
	// }
	msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"] = bytes.NewBufferString("")
	msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"] = []MsgInputModel{}
	msg["msg_content"].(MsgInputModel)["mentionedInfo"] = MsgInputModel{"type": MentionUser, "userIdList": []string{}}
	text_len := 0
	msg.appendText(text_len, args...)
	return nil
}

// 设置引用回复消息，接受被引用消息的id和发送时间
func (msg MsgInputModel) SetTextQuote(quoted_message_id string, quoted_message_send_time uint64) error {
	// if MsgContentType(msg["object_name"].(string)) != MsgText {
	// 	return errors.New("消息类型不是文本消息，请使用NewMsg(MsgText)创建文本消息后SetText")
	// }
	msg["msg_content"].(MsgInputModel)["quote"] = MsgInputModel{"original_message_id": quoted_message_id, "original_message_send_time": quoted_message_send_time,
		"quoted_message_id": quoted_message_id, "quoted_message_send_time": quoted_message_send_time}
	return nil
}

// 设置图片消息内容，接受图片url, 图片宽度, 图片高度, 图片大小 4种类型的参数，宽高单位为像素，图片大小单位为字节，不应超过10M
func (msg MsgInputModel) SetImage(url string, args ...interface{}) error {
	// if MsgContentType(msg["object_name"].(MsgContentType)) != MsgTypeImage {
	// 	return errors.New("消息类型不是图片消息，请使用NewMsg(MsgTypeImage)创建图片消息后SetImage")
	// }
	msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["url"] = url
	return nil
}

func (msg MsgInputModel) SetPost(post_id string) error {
	// if MsgContentType(msg["object_name"].(MsgContentType)) != MsgTypePost {
	// 	return errors.New("消息类型不是动态消息，请使用NewMsg(MsgTypePost)创建动态消息后SetPost")
	// }
	msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["post_id"] = post_id
	return nil
}

// recursively copy whole map
func (msg MsgInputModel) deepCopy() MsgInputModel {
	_msg := make(MsgInputModel)
	for k, v := range msg {
		_, ok := v.(MsgInputModel)
		if ok {
			_msg[k] = v.(MsgInputModel).deepCopy()
		} else {
			_msg[k] = v
		}
	}
	return _msg
}

// return a deep copy of msg
func (msg MsgInputModel) Finialize(room_id uint64) MsgInputModel {
	_msg := msg.deepCopy()

	content_type, ok := _msg["object_name"].(MsgContentType)
	if ok && content_type == MsgTypeText {
		_, ok := _msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(string)
		if !ok {
			_msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"] =
				_msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(*bytes.Buffer).String()
		}
		if len(_msg["msg_content"].(MsgInputModel)["mentionedInfo"].(MsgInputModel)["userIdList"].([]string)) == 0 &&
			_msg["msg_content"].(MsgInputModel)["mentionedInfo"].(MsgInputModel)["type"].(MsgMentionType) == MentionUser {
			delete(_msg["msg_content"].(MsgInputModel), "mentionedInfo")
		}
	}

	_msg["room_id"] = room_id

	bytesData, _ := json.Marshal(_msg["msg_content"].(MsgInputModel))
	_msg["msg_content"] = string(bytesData)
	return _msg
}

/* user input audit struct wrapper */
type UserInputAudit struct {
	AuditContent string `json:"audit_content,omitempty"` // 待审核内容，必填
	PassThrough  string `json:"pass_through,omitempty"`  // 透传信息，该字段会在审核结果回调时携带给开发者，选填
	UID          uint64 `json:"room_id,omitempty"`       // 用户id, 必填
	RoomID       uint64 `json:"uid,omitempty"`           // 房间id，选填
}
