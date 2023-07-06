package api_models

import (
	"bytes"
	"encoding/json"
	"errors"

	utils "github.com/GLGDLY/mhy_botsdk/utils"
)

/* user input message struct wrapper */
type MsgContentType string
type MsgMentionType uint64

const (
	MsgTypeText  MsgContentType = "MHY:Text"
	MsgTypeImage MsgContentType = "MHY:Image"
	MsgTypePost  MsgContentType = "MHY:Post"
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
	switch msg_type {
	case MsgTypeText, MsgTypeImage, MsgTypePost:
		msg := MsgInputModel{"object_name": msg_type, "msg_content": MsgInputModel{}}
		msg["msg_content"].(MsgInputModel)["content"] = MsgInputModel{"text": bytes.NewBufferString(""), "entities": []MsgInputModel{}}
		msg["msg_content"].(MsgInputModel)["mentionedInfo"] = MsgInputModel{"type": MentionUser, "userIdList": []string{}}
		return msg, nil
	}
	return nil, errors.New("不支持的消息类型")
}

func (msg MsgInputModel) appendText(text_len int, args ...interface{}) { // internal processor
	for _, arg := range args {
		switch arg.(type) {
		case string:
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(*bytes.Buffer).WriteString(arg.(string))
			text_len += len([]rune(arg.(string)))
		case MsgEntityMentionRobot:
			this_len := len([]rune(arg.(MsgEntityMentionRobot).Text))
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"] = append(msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"].([]MsgInputModel),
				MsgInputModel{"entity": MsgInputModel{"type": "mentioned_robot", "bot_id": arg.(MsgEntityMentionRobot).BotID},
					"offset": text_len, "length": this_len})
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(*bytes.Buffer).WriteString(arg.(MsgEntityMentionRobot).Text)
			msg["msg_content"].(MsgInputModel)["mentionedInfo"].(MsgInputModel)["userIdList"] = append(msg["msg_content"].(MsgInputModel)["mentionedInfo"].(MsgInputModel)["userIdList"].([]string),
				arg.(MsgEntityMentionRobot).BotID)
			text_len += this_len
		case MsgEntityMentionUser:
			this_len := len([]rune(arg.(MsgEntityMentionUser).Text))
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"] = append(msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"].([]MsgInputModel),
				MsgInputModel{"entity": MsgInputModel{"type": "mentioned_user", "user_id": utils.String(arg.(MsgEntityMentionUser).UserID)},
					"offset": text_len, "length": this_len})
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(*bytes.Buffer).WriteString(arg.(MsgEntityMentionUser).Text)
			msg["msg_content"].(MsgInputModel)["mentionedInfo"].(MsgInputModel)["userIdList"] = append(msg["msg_content"].(MsgInputModel)["mentionedInfo"].(MsgInputModel)["userIdList"].([]string),
				utils.String(arg.(MsgEntityMentionUser).UserID))
			text_len += this_len
		case MsgEntityMentionAll:
			this_len := len([]rune(arg.(MsgEntityMentionAll).Text))
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"] = append(msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"].([]MsgInputModel),
				MsgInputModel{"entity": MsgInputModel{"type": "mention_all"}, "offset": text_len, "length": this_len})
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(*bytes.Buffer).WriteString(arg.(MsgEntityMentionAll).Text)
			msg["msg_content"].(MsgInputModel)["mentionedInfo"].(MsgInputModel)["type"] = MentionAll
			text_len += this_len
		case MsgEntityVillaRoomLink:
			this_len := len([]rune(arg.(MsgEntityVillaRoomLink).Text))
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"] = append(msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"].([]MsgInputModel),
				MsgInputModel{"entity": MsgInputModel{"type": "villa_room_link", "villa_id": utils.String(arg.(MsgEntityVillaRoomLink).VillaID),
					"room_id": utils.String(arg.(MsgEntityVillaRoomLink).RoomID)},
					"offset": text_len, "length": this_len})
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(*bytes.Buffer).WriteString(arg.(MsgEntityVillaRoomLink).Text)
			text_len += this_len
		case MsgEntityLink:
			this_len := len([]rune(arg.(MsgEntityLink).Text))
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"] = append(msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["entities"].([]MsgInputModel),
				MsgInputModel{"entity": MsgInputModel{"type": "link", "url": arg.(MsgEntityLink).URL,
					"requires_bot_access_token": arg.(MsgEntityLink).RequiresBotAccessToken},
					"offset": text_len, "length": this_len})
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(*bytes.Buffer).WriteString(arg.(MsgEntityLink).Text)
			text_len += this_len
		default:
			content_string := utils.String(arg)
			msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(*bytes.Buffer).WriteString(content_string)
			text_len += len([]rune(content_string))
		}
	}
}

func (msg MsgInputModel) AppendText(args ...interface{}) error {
	if MsgContentType(msg["object_name"].(MsgContentType)) != MsgTypeText {
		return errors.New("消息类型不是文本消息，请使用NewMsg(MsgText)创建文本消息后SetText")
	}
	s, ok := msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(string)
	var text_len int
	if !ok {
		s = msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"].(*bytes.Buffer).String()
	}
	text_len = len([]rune(s))
	msg["msg_content"].(MsgInputModel)["content"].(MsgInputModel)["text"] = bytes.NewBufferString(s)
	msg.appendText(text_len, args...)
	return nil
}

// 设置文本消息内容
func (msg MsgInputModel) SetText(args ...interface{}) error {
	if MsgContentType(msg["object_name"].(MsgContentType)) != MsgTypeText {
		return errors.New("消息类型不是文本消息，请使用NewMsg(MsgText)创建文本消息后SetText")
	}
	msg["msg_content"].(MsgInputModel)["content"] = MsgInputModel{"text": bytes.NewBufferString(""), "entities": []MsgInputModel{}}
	msg["msg_content"].(MsgInputModel)["mentionedInfo"] = MsgInputModel{"type": MentionUser, "userIdList": []string{}}
	text_len := 0
	msg.appendText(text_len, args...)
	return nil
}

// 设置引用回复消息，接受被引用消息的id和发送时间
func (msg MsgInputModel) SetTextQuote(quoted_message_id string, quoted_message_send_time int64, original_message_id string, original_message_send_time int64) error {
	// if MsgContentType(msg["object_name"].(string)) != MsgText {
	// 	return errors.New("消息类型不是文本消息，请使用NewMsg(MsgText)创建文本消息后SetText")
	// }
	msg["msg_content"].(MsgInputModel)["quote"] = MsgInputModel{"originalMessageId": original_message_id, "originalMessageSendTime": original_message_send_time,
		"quotedMessageId": quoted_message_id, "quotedMessageSendTime": quoted_message_send_time}
	return nil
}

// 设置图片消息内容，接受图片url, 图片宽度, 图片高度, 图片大小 4种类型的参数，宽高单位为像素，图片大小单位为字节，不应超过10M
func (msg MsgInputModel) SetImage(url string, width int, height int, file_size int) error {
	if MsgContentType(msg["object_name"].(MsgContentType)) != MsgTypeImage {
		return errors.New("消息类型不是图片消息，请使用NewMsg(MsgTypeImage)创建图片消息后SetImage")
	}
	msg["msg_content"] = MsgInputModel{"content": MsgInputModel{"image_uri": url, "width": width, "height": height}}
	return nil
}

func (msg MsgInputModel) SetPost(post_id string) error {
	if MsgContentType(msg["object_name"].(MsgContentType)) != MsgTypePost {
		return errors.New("消息类型不是动态消息，请使用NewMsg(MsgTypePost)创建动态消息后SetPost")
	}
	msg["msg_content"] = MsgInputModel{"content": MsgInputModel{"post_id": post_id}}
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

	if MsgContentType(_msg["object_name"].(MsgContentType)) == MsgTypeText {
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
