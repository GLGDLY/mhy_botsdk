package bot

import (
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
)

/* http routing related */

// 新增一个http路由，用于与机器人同时运行，避免占用同一个端口，以及允许使用短路机制处理请求
//
// - path: 路由路径
//
// - addr: 路由地址
//
// - handler: 路由处理器回调函数——函数返回true则短路，返回false则继续执行（系统将在所有同路径端口的handler都不短路后再处理消息）
func AddHttpRouteHandler(path, addr string, handler func(*gin.Context) bool) {
	for _, _bot := range bot_context_manager {
		if _bot.bot.addr_key == addr {
			if _bot.svr_ctx.handles[path] == nil {
				_bot.svr_ctx.handles[path] = make([]func(*gin.Context) bool, 0)
			}
			_bot.svr_ctx.handles[path] = append(_bot.svr_ctx.handles[path], handler)
			return
		}
	}
	if other_svr_context_manager[addr] == nil {
		other_svr_context_manager[addr] = &serverContext{
			svr:        gin.Default(),
			svr_addr:   addr,
			is_running: false,
			wg:         &sync.WaitGroup{},
			bots:       make(map[string][]*Bot),
			handles:    make(map[string][]func(*gin.Context) bool),
		}
	} else {
		if other_svr_context_manager[addr].handles[path] == nil {
			other_svr_context_manager[addr].handles[path] = make([]func(*gin.Context) bool, 0)
		}
		other_svr_context_manager[addr].handles[path] = append(other_svr_context_manager[addr].handles[path], handler)
	}
}

func RemoveHttpRouteHandler(path, addr string, handler func(*gin.Context) bool) {
	for _, _bot := range bot_context_manager {
		if _bot.bot.addr_key == addr {
			if _bot.svr_ctx.handles[path] == nil {
				return
			}
			for i, h := range _bot.svr_ctx.handles[path] {
				if reflect.ValueOf(h).Pointer() == reflect.ValueOf(handler).Pointer() {
					_bot.svr_ctx.handles[path] = append(_bot.svr_ctx.handles[path][:i], _bot.svr_ctx.handles[path][i+1:]...)
					return
				}
			}
		}
	}
	if other_svr_context_manager[addr] == nil {
		return
	} else {
		if other_svr_context_manager[addr].handles[path] == nil {
			return
		}
		for i, h := range other_svr_context_manager[addr].handles[path] {
			if reflect.ValueOf(h).Pointer() == reflect.ValueOf(handler).Pointer() {
				other_svr_context_manager[addr].handles[path] = append(other_svr_context_manager[addr].handles[path][:i], other_svr_context_manager[addr].handles[path][i+1:]...)
				return
			}
		}
	}
}

func StartAllHttpServer() {
	for _, _bot := range bot_context_manager {
		if _bot.svr_ctx.is_running {
			continue
		}
		_bot.svr_ctx.is_running = true
		_bot.svr_ctx.wg.Add(1)
		b := _bot
		go func() {
			defer b.svr_ctx.wg.Done()
			b.svr_ctx.svr.Run(b.bot.addr_key)
		}()
	}
	for _, _svr := range other_svr_context_manager {
		if _svr.is_running {
			continue
		}
		_svr.is_running = true
		_svr.wg.Add(1)
		s := _svr
		go func() {
			defer s.wg.Done()
			s.svr.Run(s.svr_addr)
		}()
	}
}
