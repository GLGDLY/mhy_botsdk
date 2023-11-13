package apis

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	models "github.com/GLGDLY/mhy_botsdk/api_models"
)

func (api *ApiBase) UploadImage(villa_id uint64, url string) (models.UploadImageModel, int, error) {
	data := map[string]interface{}{"url": url}
	request, build_req_err := http.NewRequest(http.MethodPost, api.makeURL("/vila/api/bot/platform/transferImage"), api.parseJSON(data))
	var resp_data models.UploadImageModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}

func (api *ApiBase) UploadFileImage(villa_id uint64, file_path string) (models.UploadFileImageModel, int, error) {
	var resp_data models.UploadFileImageModel

	file, err := os.Open(file_path)
	if err != nil {
		return resp_data, 600, err
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return resp_data, 600, err
	}

	file_ext := strings.ToLower(file_path[strings.LastIndex(file_path, ".")+1:])

	h := md5.New()
	if _, err := h.Write(buf); err != nil {
		return resp_data, 600, err
	}

	data := map[string]interface{}{"md5": hex.EncodeToString(h.Sum(nil)), "ext": file_ext}
	fmt.Println(data["md5"])
	request, build_req_err := http.NewRequest(http.MethodGet, api.makeURL("/vila/api/bot/platform/getUploadImageParams"), api.parseJSON(data))
	var param models.GetUploadFileImageParamsModel
	http_status, err := api.RequestHandler(villa_id, request, build_req_err, &param)
	if err != nil {
		return resp_data, http_status, err
	}

	var requestBody bytes.Buffer

	multiPartWriter := multipart.NewWriter(&requestBody)

	multiPartWriter.WriteField("x:extra", param.Data.Params.CallbackVar.XExtra)
	multiPartWriter.WriteField("OSSAccessKeyId", param.Data.Params.AccessID)
	multiPartWriter.WriteField("signature", param.Data.Params.Signature)
	multiPartWriter.WriteField("success_action_status", param.Data.Params.SuccessActionStatus)
	multiPartWriter.WriteField("name", param.Data.Params.Name)
	multiPartWriter.WriteField("callback", param.Data.Params.Callback)
	multiPartWriter.WriteField("x-oss-content-type", param.Data.Params.XOSSContentType)
	multiPartWriter.WriteField("key", param.Data.Params.Key)
	multiPartWriter.WriteField("policy", param.Data.Params.Policy)

	fileWriter, err := multiPartWriter.CreateFormFile("file", data["md5"].(string)+"."+data["ext"].(string))
	if err != nil {
		return resp_data, 600, err
	}

	_, err = fileWriter.Write(buf)
	if err != nil {
		return resp_data, 600, err
	}

	err = multiPartWriter.Close()
	if err != nil {
		return resp_data, 600, err
	}

	request, build_req_err = http.NewRequest(http.MethodPost, param.Data.Params.Host, &requestBody)
	if build_req_err != nil {
		return resp_data, 600, build_req_err
	}
	request.Header.Set("Content-Type", multiPartWriter.FormDataContentType())
	fmt.Println(request.Header)

	http_status, err = api.RequestHandler(villa_id, request, build_req_err, &resp_data)
	return resp_data, http_status, err
}
