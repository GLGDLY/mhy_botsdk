package apis

import (
	"net/http"

	models "github.com/GLGDLY/mhy_botsdk/api_models"
)

func (api *ApiBase) CheckMemberBotAccessToken(villa_id uint64, token string) (models.CheckMemberBotAccessTokenModel, int, error) {
	data := map[string]interface{}{"token": token}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/checkMemberBotAccessToken"), api.parseJSON(data))
	var resp_data models.CheckMemberBotAccessTokenModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}
