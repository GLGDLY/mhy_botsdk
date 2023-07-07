package apis

import (
	"fmt"
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

// 使用models.NewMsg创建消息，然后使用models.SetText等方法加入内容，最后使用此函数发送
func (api *ApiBase) SendMessage(villa_id uint64, room_id uint64, _msg models.MsgInputModel) (models.SendMessageModel, int, error) {
	msg := _msg.Finialize(room_id)
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/sendMessage"), api.parseJSON(msg))
	b, _ := request.GetBody()
	body_bytes := make([]byte, 1024)
	b.Read(body_bytes)
	fmt.Println(string(body_bytes))
	var resp_data models.SendMessageModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}
