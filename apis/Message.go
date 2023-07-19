package apis

import (
	"net/http"

	models "github.com/GLGDLY/mhy_botsdk/api_models"
)

func (api *ApiBase) PinMessage(villa_id uint64, msg_uid string, is_cancel bool, room_id uint64, send_at int64) (models.EmptyModel, int, error) {
	data := map[string]interface{}{"msg_uid": msg_uid, "is_cancel": is_cancel, "room_id": room_id, "send_at": send_at}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/pinMessage"), api.parseJSON(data))
	var resp_data models.EmptyModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) RecallMessage(villa_id uint64, msg_uid string, room_id uint64, msg_time int64) (models.EmptyModel, int, error) {
	data := map[string]interface{}{"msg_uid": msg_uid, "room_id": room_id, "msg_time": msg_time}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/recallMessage"), api.parseJSON(data))
	var resp_data models.EmptyModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

// func (api *ApiBase) uploadImageForMessage(villa_id uint64, _msg *models.MsgInputModel) (int, error) {
// 	url := (*_msg)["msg_content"].(models.MsgInputModel)["content"].(models.MsgInputModel)["image_uri"].(string)
// 	resp_data, http_status, err := api.UploadImage(villa_id, url)
// 	if err != nil {
// 		return http_status, err
// 	}
// 	if http_status != 200 {
// 		return http_status, nil
// 	}
// 	(*_msg)["msg_content"].(models.MsgInputModel)["content"].(models.MsgInputModel)["image_uri"] = resp_data.Data.NewURL
// 	return http_status, nil
// }

// 使用models.NewMsg创建消息，然后使用models.SetText等方法加入内容，最后使用此函数发送
func (api *ApiBase) SendMessageCustomize(villa_id uint64, room_id uint64, _msg models.MsgInputModel) (models.SendMessageModel, int, error) {
	// if models.MsgContentType(_msg["object_name"].(models.MsgContentType)) == models.MsgTypeImage {
	// 	http_status, err := api.uploadImageForMessage(villa_id, &_msg)
	// 	if err != nil || http_status != 200 {
	// 		return models.SendMessageModel{}, http_status, err
	// 	}
	// }
	msg := _msg.Finialize(room_id)
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/sendMessage"), api.parseJSON(msg))
	var resp_data models.SendMessageModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

// 使用内嵌格式发送消息，并自动处理内部Entity（<@xxx>为艾特机器人或用户，<@everyone>为艾特全体，<#xxx>为跳转房间，<$xxx>为跳转连接）
// 艾特用户会自动获取用户昵称，跳转房间会自动获取房间名称；艾特机器人会显示文字“机器人”，艾特全体会显示“全体成员”，跳转连接会显示链接自身
// 使用\< 和 \> 可转义 < 和 >，不会被解析为Entity
func (api *ApiBase) SendMessage(villa_id uint64, room_id uint64, _msg_parts ...string) (models.SendMessageModel, int, error) {
	msg, _ := models.NewMsg(models.MsgTypeText)
	err := api.messageParser(&msg, villa_id, _msg_parts...)
	if err != nil {
		return models.SendMessageModel{}, 600, err
	}
	return api.SendMessageCustomize(villa_id, room_id, msg)
}
