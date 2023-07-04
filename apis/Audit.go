package apis

import (
	"bytes"
	"encoding/json"
	"net/http"

	models "github.com/GLGDLY/mhy_botsdk/api_models"
)

func (api *ApiBase) Audit(villa_id uint64, audit_input models.UserInputAudit) (models.AuditModel, int, error) {
	bytesData, _ := json.Marshal(audit_input)
	request, build_req_err := http.NewRequest("POST", api.makeURL("/vila/api/bot/platform/audit"), bytes.NewReader(bytesData))
	var resp_data models.AuditModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}
