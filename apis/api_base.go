package apis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"time"

	base "github.com/GLGDLY/mhy_botsdk"
	models "github.com/GLGDLY/mhy_botsdk/models"
	utils "github.com/GLGDLY/mhy_botsdk/utils"
)

const open_api_url string = "https://bbs-api.miyoushe.com"

type ApiBase struct {
	Base    models.BotBase
	session http.Client
}

func MakeAPIBase(base models.BotBase, timeout time.Duration) *ApiBase {
	return &ApiBase{
		Base: base,
		session: http.Client{
			Timeout: timeout,
		},
	}
}

func (api *ApiBase) SetTimeout(timeout time.Duration) {
	api.session.Timeout = timeout
}

func (api *ApiBase) makeURL(path string) string {
	return open_api_url + path
}

func (api *ApiBase) parseParams(raw_url string, params map[string]interface{}) string {
	_url, _ := url.Parse(raw_url)
	q := _url.Query()
	for k, v := range params {
		q.Set(k, utils.String(v))
	}
	_url.RawQuery = q.Encode()
	return _url.String()
}

func (api *ApiBase) parseJSON(json_map map[string]interface{}) *bytes.Reader {
	bytesData, _ := json.Marshal(json_map)
	return bytes.NewReader(bytesData)
}

func (api *ApiBase) RequestHandler(villa_id uint64, request *http.Request, build_req_err error, resp_data interface{}) (int, error) {
	if build_req_err != nil {
		return 600, build_req_err
	}
	if reflect.TypeOf(resp_data).Kind() != reflect.Ptr {
		return 600, errors.New("resp_data is not a pointer")
	}
	resp, err := api.Request(villa_id, request)
	if err != nil {
		return 600, err
	}
	defer resp.Body.Close()
	if resp.Header.Get("Content-Type") != "application/json" {
		return resp.StatusCode, errors.New("response Content-Type is not application/json")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, err
	}

	s := reflect.ValueOf(resp_data)
	reflect.Indirect(s).FieldByName("APIBaseModel").FieldByName("RawData").SetString(string(data))

	err = json.Unmarshal(data, resp_data)
	if err != nil {
		var v map[string]interface{}
		_err := json.Unmarshal(data, &v)
		if _err != nil {
			return resp.StatusCode, _err
		}
		fmt.Println(err, "on decoding data:\n", v)
	}
	return resp.StatusCode, err
}

func (api *ApiBase) Request(villa_id uint64, request *http.Request) (*http.Response, error) {
	request.Header.Set("x-rpc-bot_id", api.Base.ID)
	request.Header.Set("x-rpc-bot_secret", api.Base.EncodedSecret)
	request.Header.Set("x-rpc-bot_villa_id", utils.String(villa_id))
	request.Header.Set("User-Agent", "github.com/GLGDLY/mhy_botsdk"+base.VERSION)
	return api.session.Do(request)
}
