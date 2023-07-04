package apis

import (
	"net/http"

	models "github.com/GLGDLY/mhy_botsdk/api_models"
)

func (api *ApiBase) OperateMemberToRole(villa_id uint64, role_id, uid uint64, is_add bool) (models.EmptyModel, int, error) {
	data := map[string]interface{}{"role_id": role_id, "uid": uid, "is_add": is_add}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/operateMemberToRole"), api.parseJSON(data))
	var resp_data models.EmptyModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) CreateMemberRole(villa_id uint64, name, color string, permissions []string) (models.CreateRoleModel, int, error) {
	data := map[string]interface{}{"name": name, "color": color, "permissions": permissions}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/createMemberRole"), api.parseJSON(data))
	var resp_data models.CreateRoleModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) EditMemberRole(villa_id uint64, id uint64, name, color string, permissions []string) (models.EmptyModel, int, error) {
	data := map[string]interface{}{"id": id, "name": name, "color": color, "permissions": permissions}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/editMemberRole"), api.parseJSON(data))
	var resp_data models.EmptyModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) DeleteMemberRole(villa_id uint64, id uint64) (models.EmptyModel, int, error) {
	data := map[string]interface{}{"id": id}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/deleteMemberRole"), api.parseJSON(data))
	var resp_data models.EmptyModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) GetMemberRoleInfo(villa_id uint64, role_id uint64) (models.GetRoleInfoModel, int, error) {
	data := map[string]interface{}{"role_id": role_id}
	request, build_req_err := http.NewRequest("GET", api.makeURL("/vila/api/bot/platform/getMemberRoleInfo"), api.parseJSON(data))
	var resp_data models.GetRoleInfoModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) GetVillaMemberRoles(villa_id uint64) (models.GetVillaMemberRolesModel, int, error) {
	data := map[string]interface{}{}
	request, build_req_err := http.NewRequest("GET", api.makeURL("/vila/api/bot/platform/getVillaMemberRoles"), api.parseJSON(data))
	var resp_data models.GetVillaMemberRolesModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}
