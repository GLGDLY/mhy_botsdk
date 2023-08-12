package bot

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const ws_heartbeat_interval = 30 * time.Second

var ws_heartbeat_msg = []byte("{\"type\":\"hb\"}")
var upgrader = websocket.Upgrader{}

func (_bot *Bot) defaultWSProxyLoop(ws *websocket.Conn, msg_chan chan [2][]byte) {
	heartbeat := time.NewTicker(ws_heartbeat_interval)
	defer heartbeat.Stop()

	done_in2out := make(chan byte)
	defer close(done_in2out)
	done_out2in := false
	defer func() {
		done_out2in = true
	}()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		for {
			if done_out2in {
				return
			}
			mtype, _, err := ws.ReadMessage()
			if err != nil || mtype == websocket.CloseMessage {
				done_in2out <- 1
				return
			}
			if mtype == websocket.PingMessage {
				ws.WriteMessage(websocket.PongMessage, nil)
			}
		}
	}()

	for {
		select {
		case msg := <-msg_chan:
			body := strings.TrimSpace(string(msg[0]))
			body = body[:len(body)-1] + ",\"sign\":\"" + string(msg[1]) + "\"}"
			err := ws.WriteMessage(websocket.TextMessage, []byte(body))
			if err != nil {
				_bot.Logger.Errorf("反向代理 %v 发送失败：%s", ws.RemoteAddr().String(), err.Error())
			} else {
				_bot.Logger.Debugf("反向代理 %v 发送成功", ws.RemoteAddr().String())
			}
		case <-heartbeat.C:
			ws.WriteMessage(websocket.PingMessage, ws_heartbeat_msg)
		case <-done_in2out:
			return
		case <-interrupt:
			ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return
		}
	}
}

func (_bot *Bot) defaultWSProxyHook(msg_chan chan [2][]byte, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	_bot.Logger.Infof("反向代理 %v 连接成功", c.Request.URL.String())
	_bot.defaultWSProxyLoop(conn, msg_chan)
}

// 添加事件的WebSocket反向代理（服务端server），会每隔30秒发送一次心跳包
//
// - 另一端可使用 NewWsBot() 创建Bot实例
func (_bot *Bot) AddReverseProxyWS(path, addr string) {
	var msg_chan = make(chan [2][]byte)

	is_added := false

	// check if exists in http routes
	if svr_ctx_ptr := other_svr_context_manager[addr]; svr_ctx_ptr != nil {
		if svr_ctx_ptr.handles[path] != nil {
			_bot.Logger.Errorf("相关端口 %v 和路径 %v 已被其他服务占用，无法添加反向代理", addr, path)
			close(msg_chan)
			return
		} else {
			is_added = true
			svr_ctx_ptr.handles[path] = []func(*gin.Context) bool{func(c *gin.Context) bool {
				_bot.defaultWSProxyHook(msg_chan, c)
				return true
			}}
		}
	}
	// check if exists in bot routes
	for _, _bot_ctx := range bot_context_manager {
		if _bot_ctx.bot.addr_key == addr {
			if _bot_ctx.bot.path_key == path {
				_bot.Logger.Errorf("相关端口 %v 和路径 %v 已被其他服务占用，无法添加反向代理", addr, path)
				close(msg_chan)
				return
			} else if _bot_ctx.svr_ctx.handles[path] != nil {
				_bot.Logger.Errorf("相关端口 %v 和路径 %v 已被其他服务占用，无法添加反向代理", addr, path)
				close(msg_chan)
				return
			} else {
				is_added = true
				_bot_ctx.svr_ctx.handles[path] = []func(*gin.Context) bool{func(c *gin.Context) bool {
					_bot.defaultWSProxyHook(msg_chan, c)
					return true
				}}
			}
			break
		}
	}
	// else, create new svr_ctx to other_svr_context_manager
	if !is_added {
		svr_ctx_handles := make(map[string][]func(*gin.Context) bool)
		svr_ctx_handles[path] = []func(*gin.Context) bool{func(c *gin.Context) bool {
			_bot.defaultWSProxyHook(msg_chan, c)
			return true
		}}
		other_svr_context_manager[addr] = &serverContext{
			svr:        gin.Default(),
			svr_addr:   addr,
			is_running: false,
			wg:         &sync.WaitGroup{},
			bots:       make(map[string][]*Bot),
			handles:    svr_ctx_handles,
		}
	}

	_bot.reverse_proxy_ws_msg_chan = append(_bot.reverse_proxy_ws_msg_chan, msg_chan)
}

func (_bot *Bot) defaultHttpProxyLoop(_url *url.URL, msg_chan chan [2][]byte) {
	for msg := range msg_chan {
		req := http.Request{
			Method: http.MethodPost,
			URL:    _url,
			Header: http.Header{
				"Content-Type":   {"application/json"},
				"x-rpc-bot_sign": {string(msg[1])},
			},
			Body: io.NopCloser(bytes.NewReader(msg[0])),
		}
		resp, err := http.DefaultClient.Do(&req)
		if err != nil {
			_bot.Logger.Errorf("反向代理 %v 发送失败：%s", _url, err.Error())
		} else {
			resp.Body.Close()
			_bot.Logger.Debugf("反向代理 %v 发送成功", _url)
		}
	}
}

// 添加事件的http服务端反向代理
//
// - 另一端可正常使用 NewBot() 创建Bot实例
func (_bot *Bot) AddReverseProxyHTTP(raw_url string) {
	_url, err := url.Parse(raw_url)
	if err != nil {
		_bot.Logger.Errorf("反向代理 %v 添加失败：%s", raw_url, err.Error())
		return
	}

	var msg_chan = make(chan [2][]byte)
	_bot.reverse_proxy_http_msg_chan = append(_bot.reverse_proxy_http_msg_chan, msg_chan)
	go _bot.defaultHttpProxyLoop(_url, msg_chan)
}
