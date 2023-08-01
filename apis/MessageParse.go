package apis

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	models "github.com/GLGDLY/mhy_botsdk/api_models"
)

func (api *ApiBase) MessageParser(msg *models.MsgInputModel, villa_id uint64, _msg_parts ...string) error {
	/* for parsing */
	msg_buf := bytes.NewBufferString("")
	is_entity := false
	var entity_type *models.MsgEntityType = nil
	/* for appending entity cache */
	usernames := make(map[uint64]string)
	roomnames := make(map[uint64]string)

	skip_next_word := false
	for i, msg_part := range _msg_parts {
		for j, word := range msg_part {
			// check for escape
			if skip_next_word {
				skip_next_word = false
				msg_buf.WriteRune(word)
				continue
			}
			if word == '\\' {
				skip_next_word = true
				continue

			}
			// check for entity
			if word == '<' {
				if is_entity {
					return fmt.Errorf(`invalid format, unexcepted "<" in position %d on msg arg %d`, j, i)
				}
				is_entity = true
				msg.AppendText(msg_buf.String())
				msg_buf.Reset()
				entity_type = nil
			} else if word == '>' {
				if !is_entity {
					return fmt.Errorf(`invalid format, unexcepted ">" in position %d on msg arg %d`, j, i)
				}
				entity_content := msg_buf.String()
				switch *entity_type {
				case models.MsgEntityMentionUserType:
					if entity_content == "everyone" {
						msg.AppendText(models.MsgEntityMentionAll{
							Text: "@全体成员",
						})
						break
					} else if strings.HasPrefix(entity_content, "bot_") {
						msg.AppendText(models.MsgEntityMentionRobot{
							Text:  "@机器人",
							BotID: entity_content,
						})
						break
					}
					uid, error := strconv.ParseUint(entity_content, 10, 64)
					if error != nil {
						return fmt.Errorf(`invalid format, invalid user id "%s" in position %d on msg arg %d`, entity_content, j, i)
					}
					username, ok := usernames[uid]
					if !ok {
						resp, http, err := api.GetMember(villa_id, uid)
						if err != nil {
							return err
						}
						if http != 200 || resp.Retcode != 0 {
							return fmt.Errorf(`failed to fetch user info, http code: %d, retcode: %d`, http, resp.Retcode)
						}
						username = resp.Data.Member.Basic.Nickname
						usernames[uid] = username
					}
					msg.AppendText(models.MsgEntityMentionUser{
						Text:   "@" + username,
						UserID: uid,
					})
				case models.MsgEntityVillaRoomLinkType:
					room_id, error := strconv.ParseUint(entity_content, 10, 64)
					if error != nil {
						return fmt.Errorf(`invalid format, invalid room id "%s" in position %d on msg arg %d`, entity_content, j, i)
					}
					roomname, ok := roomnames[room_id]
					if !ok {
						resp, http, err := api.GetRoom(villa_id, room_id)
						if err != nil {
							return err
						}
						if http != 200 || resp.Retcode != 0 {
							return fmt.Errorf(`failed to fetch room info, http code: %d, retcode: %d`, http, resp.Retcode)
						}
						roomname = resp.Data.Room.RoomName
						roomnames[room_id] = roomname
					}
					msg.AppendText(models.MsgEntityVillaRoomLink{
						Text:    "#" + roomname,
						VillaID: villa_id,
						RoomID:  room_id,
					})
				case models.MsgEntityLinkType:
					msg.AppendText(models.MsgEntityLink{
						URL:                    entity_content,
						RequiresBotAccessToken: false,
					})
				}
				is_entity = false
				msg_buf.Reset()
			} else {
				if is_entity && entity_type == nil {
					var _entity_type models.MsgEntityType
					switch word {
					case '@':
						_entity_type = models.MsgEntityMentionUserType
					case '#':
						_entity_type = models.MsgEntityVillaRoomLinkType
					case '$':
						_entity_type = models.MsgEntityLinkType
					}
					entity_type = &_entity_type
				} else {
					msg_buf.WriteRune(word)
				}
			}
		}
	}
	if is_entity {
		return fmt.Errorf(`invalid format, unexcepted EOF till last msg arg (start with "<" but not end with ">")`)
	}
	msg.AppendText(msg_buf.String())
	return nil
}
