package apis

import (
	"net/http"

	models "github.com/GLGDLY/mhy_botsdk/api_models"
)

func (api *ApiBase) GetVilla(villa_id uint64) (models.GetVillaModel, int, error) {
	request, build_req_err := http.NewRequest("GET", api.makeURL("/vila/api/bot/platform/getVilla"), nil)
	var resp_data models.GetVillaModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}
