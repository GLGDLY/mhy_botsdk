package apis

import (
	models "mhy_botsdk/api_models"
	"net/http"
)

func (api *ApiBase) GetMember(villa_id uint64, uid uint64) (models.GetMemberModel, int, error) {
	query := map[string]interface{}{"uid": uid}
	request, build_req_err := http.NewRequest("GET", api.parseParams(api.makeURL("/vila/api/bot/platform/getMember"), query), nil)
	var resp_data models.GetMemberModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) GetVillaMembers(villa_id uint64, offset_str string, size uint64) (models.GetVillaMembersModel, int, error) {
	query := map[string]interface{}{"offset": offset_str, "size": size}
	request, build_req_err := http.NewRequest("GET", api.parseParams(api.makeURL("/vila/api/bot/platform/getVillaMembers"), query), nil)
	var resp_data models.GetVillaMembersModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) GetVillaMembersDefault(villa_id uint64) (models.GetVillaMembersModel, int, error) {
	query := map[string]interface{}{"offset": "", "size": "18446744073709551615"} // MaxUint64
	request, build_req_err := http.NewRequest("GET", api.parseParams(api.makeURL("/vila/api/bot/platform/getVillaMembers"), query), nil)
	var resp_data models.GetVillaMembersModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) DeleteVillaMember(villa_id uint64, uid uint64) (models.EmptyModel, int, error) {
	data := map[string]interface{}{"uid": uid}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/deleteVillaMember"), api.parseJSON(data))
	var resp_data models.EmptyModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}
