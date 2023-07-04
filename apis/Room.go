package apis

import (
	models "mhy_botsdk/api_models"
	"net/http"
)

func (api *ApiBase) CreateGroup(villa_id uint64, group_name string) (models.CreateRoomModel, int, error) {
	data := map[string]interface{}{"group_name": group_name}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/createGroup"), api.parseJSON(data))
	var resp_data models.CreateRoomModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) EditGroup(villa_id uint64, group_id uint64, group_name string) (models.EmptyModel, int, error) {
	data := map[string]interface{}{"group_id": group_id, "group_name": group_name}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/editGroup"), api.parseJSON(data))
	var resp_data models.EmptyModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) DeleteGroup(villa_id uint64, group_id uint64) (models.EmptyModel, int, error) {
	data := map[string]interface{}{"group_id": group_id}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/deleteGroup"), api.parseJSON(data))
	var resp_data models.EmptyModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) GetGroupList(villa_id uint64) (models.GetGroupListModel, int, error) {
	data := map[string]interface{}{}
	request, build_req_err := http.NewRequest("GET", api.makeURL("/vila/api/bot/platform/getGroupList"), api.parseJSON(data))
	var resp_data models.GetGroupListModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) EditRoom(villa_id uint64, room_id uint64, room_name string) (models.EmptyModel, int, error) {
	data := map[string]interface{}{"room_id": room_id, "room_name": room_name}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/editRoom"), api.parseJSON(data))
	var resp_data models.EmptyModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) DeleteRoom(villa_id uint64, room_id uint64) (models.EmptyModel, int, error) {
	data := map[string]interface{}{"room_id": room_id}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/deleteRoom"), api.parseJSON(data))
	var resp_data models.EmptyModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) GetRoom(villa_id uint64, room_id uint64) (models.GetRoomModel, int, error) {
	data := map[string]interface{}{"room_id": room_id}
	request, build_req_err := http.NewRequest("GET", api.makeURL("/vila/api/bot/platform/getRoom"), api.parseJSON(data))
	var resp_data models.GetRoomModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) GetVillaGroupRoomList(villa_id uint64) (models.GetVillaGroupRoomListModel, int, error) {
	data := map[string]interface{}{}
	request, build_req_err := http.NewRequest("GET", api.makeURL("/vila/api/bot/platform/getVillaGroupRoomList"), api.parseJSON(data))
	var resp_data models.GetVillaGroupRoomListModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}
