package apis

import (
	"net/http"

	models "github.com/GLGDLY/mhy_botsdk/api_models"
)

func (api *ApiBase) UploadImage(villa_id uint64, url string) (models.UploadImageModel, int, error) {
	data := map[string]interface{}{"url": url}
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/transferImage"), api.parseJSON(data))
	var resp_data models.UploadImageModel
	http_status, err := api.RequestHandler(0, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}
