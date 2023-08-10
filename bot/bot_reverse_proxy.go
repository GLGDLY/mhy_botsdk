package bot

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const ws_heartbeat_interval = 30 * time.Second

var ws_heartbeat_msg = []byte("{\"type\":\"hb\"}")

func defaultWSProxyLoop(ws *websocket.Conn, msg_chan chan []byte) {
	heartbeat := time.NewTicker(ws_heartbeat_interval)
	defer heartbeat.Stop()

	done := make(chan byte)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		for {
			mtype, _, err := ws.ReadMessage()
			if err != nil || mtype == websocket.CloseMessage {
				done <- 1
				return
			}
		}
	}()

	for {
		select {
		case msg := <-msg_chan:
			ws.WriteMessage(websocket.TextMessage, msg)
		case <-heartbeat.C:
			ws.WriteMessage(websocket.PingMessage, ws_heartbeat_msg)
		case <-done:
			return
		case <-interrupt:
			ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}

func (_bot *Bot) defaultWSConnectLoop(_url string, msg_chan chan []byte) {
	do_once_flag := true
	for {
		func() {
			conn, _, err := websocket.DefaultDialer.Dial(_url, nil)
			if err != nil {
				_bot.Logger.Errorf("反向代理 %v 连接失败：%s", _url, err.Error())
				time.Sleep(5 * time.Second)
				return
			}
			defer conn.Close()
			if do_once_flag {
				do_once_flag = false
				_bot.Logger.Infof("反向代理 %v 连接成功", _url)
			} else {
				_bot.Logger.Infof("反向代理 %v 重新连接成功", _url)
			}
			defaultWSProxyLoop(conn, msg_chan)
			time.Sleep(5 * time.Second)
		}()
	}
}

func (_bot *Bot) defaultHttpProxyLoop(_url string, msg_chan chan []byte) {
	for msg := range msg_chan {
		_, err := http.Post(_url, "application/json", bytes.NewReader(msg))
		if err != nil {
			_bot.Logger.Errorf("反向代理 %v 发送失败：%s", _url, err.Error())
			continue
		}
	}
}

func (_bot *Bot) addReverseProxyWS(_url url.URL) {
	var msg_chan = make(chan []byte)
	_bot.reverse_proxy_ws_msg_chan = append(_bot.reverse_proxy_ws_msg_chan, msg_chan)
	go _bot.defaultWSConnectLoop(_url.String(), msg_chan)
}

func (_bot *Bot) addReverseProxyHTTP(_url url.URL) {
	var msg_chan = make(chan []byte)
	_bot.reverse_proxy_http_msg_chan = append(_bot.reverse_proxy_http_msg_chan, msg_chan)
	go _bot.defaultHttpProxyLoop(_url.String(), msg_chan)
}

// 添加事件的反向代理，支持ws和http，其中ws会每隔30秒发送一次心跳包
//
// - 如使用ws协议，另一端可使用NewWsBot()创建Bot实例
//
// - 如使用http协议，另一端可正常使用NewBot()创建Bot实例
func (_bot *Bot) AddReverseProxy(path string) error {
	_url, err := url.Parse(path)
	if err != nil {
		return err
	}
	switch _url.Scheme {
	case "ws":
		_bot.addReverseProxyWS(*_url)
	case "http":
		_bot.addReverseProxyHTTP(*_url)
	default:
		return errors.New("无法识别的反向代理类型（必须以ws或http开头）")
	}
	return nil
}
