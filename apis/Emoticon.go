package apis

import (
	"net/http"

	models "github.com/GLGDLY/mhy_botsdk/api_models"
)

func (api *ApiBase) GetAllEmoticons(villa_id uint64) (models.GetAllEmoticonsModel, int, error) {
	data := map[string]interface{}{}
	request, build_req_err := http.NewRequest(http.MethodGet, api.makeURL("/vila/api/bot/platform/getAllEmoticons"), api.parseJSON(data))
	var resp_data models.GetAllEmoticonsModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}
