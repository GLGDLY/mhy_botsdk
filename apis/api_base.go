package apis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	base "mhy_botsdk"
	models "mhy_botsdk/models"
	utils "mhy_botsdk/utils"
	"net/http"
	"net/url"
	"reflect"
)

const open_api_url string = "https://bbs-api.miyoushe.com"

type ApiBase struct {
	Base    models.BotBase
	session http.Client
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
		return resp.StatusCode, errors.New("Response Content-Type is not application/json")
	}
	data, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, resp_data)
	if err != nil {
		var v map[string]interface{}
		json.Unmarshal(data, &v)
		fmt.Println(err, "on data:\n", v)
	}
	return resp.StatusCode, err
}

func (api *ApiBase) Request(villa_id uint64, request *http.Request) (*http.Response, error) {
	request.Header.Set("x-rpc-bot_id", api.Base.ID)
	request.Header.Set("x-rpc-bot_secret", api.Base.Secret)
	request.Header.Set("x-rpc-bot_villa_id", utils.String(villa_id))
	request.Header.Set("User-Agent", "mhy_botsdk"+base.VERSION)
	return api.session.Do(request)
}
