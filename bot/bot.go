package bot

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	apis "github.com/GLGDLY/mhy_botsdk/apis"
	commands "github.com/GLGDLY/mhy_botsdk/commands"
	events "github.com/GLGDLY/mhy_botsdk/events"
	logger "github.com/GLGDLY/mhy_botsdk/logger"
	models "github.com/GLGDLY/mhy_botsdk/models"
	plugin "github.com/GLGDLY/mhy_botsdk/plugins"
	utils "github.com/GLGDLY/mhy_botsdk/utils"

	_ "github.com/fatih/color" // used init for support color output in console
)

type Bot struct {
	Base           models.BotBase // 机器人基本信息
	path_key       string
	addr_key       string
	svr            *gin.Engine
	filter_manager *filterManager      // used to filter event that passed repeatly in a short time
	abstract_bot   *plugin.AbstractBot // bot的抽象类，用于为插件提供基础机器人功能
	is_running     bool                // 是否正在运行
	/* 事件监听器开始 */
	listeners_join_villa         []events.BotListenerJoinVilla
	listeners_send_message       []events.BotListenerSendMessage
	listeners_create_robot       []events.BotListenerCreateRobot
	listeners_delete_robot       []events.BotListenerDeleteRobot
	listeners_add_quick_emoticon []events.BotListenerAddQuickEmoticon
	listeners_audit_callback     []events.BotListenerAuditCallback
	listeners_raw_request        []events.BotListenerRawRequest
	/* 事件监听器结束 */
	/* reverse proxy start */
	reverse_proxy_http_msg_chan []chan []byte
	reverse_proxy_ws_msg_chan   []chan []byte
	/* reverse proxy end */
	use_default_logger                   bool                      // 是否使用默认的日志记录器，默认为false
	is_plugins_short_circuit_affect_main bool                      // 插件中的指令短路是否会影响主程序其余指令和监听器的执行，默认为false
	is_filter_self_msg                   bool                      // 是否过滤自己发送的消息，默认为true
	is_verify_msg_signature              bool                      // 是否验证接受到事件的签名，默认为true
	plugins                              map[string]*plugin.Plugin // 插件列表
	on_commands                          []commands.OnCommand      // 处理消息事件的指令列表
	preprocessors                        []commands.Preprocessor   // 消息事的预处理器，用于在运行指令列表和监听器之前处理事件
	wait_for_command_registers           []waitForCommandRegister  // 用户处理消息时暂停等待指令的处理列表
	Api                                  *apis.ApiBase             // api接口
	Logger                               logger.LoggerInterface    // 日志记录器
}

/* context managers start */
type server_context struct {
	svr        *gin.Engine
	svr_addr   string
	is_running bool
	wg         *sync.WaitGroup
	bots       map[string][]*Bot                    // path: bot
	handles    map[string][]func(*gin.Context) bool // path: handler
}

type bot_context struct {
	bot     *Bot
	svr_ctx *server_context
}

var bot_context_manager map[string]*bot_context = make(map[string]*bot_context) // id: bot
var other_svr_context_manager map[string]*server_context = make(map[string]*server_context)

/* context managers end */

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
		other_svr_context_manager[addr] = &server_context{
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

/* bot related */

// NewBot 创建一个机器人实例，bot_id 为机器人的id，bot_secret 为机器人的secret，path 为接收事件的路径（如"/"），addr 为接收事件的地址（如":8888"）
//
// 机器人实例创建后，需要调用 Start() 方法启动机器人，但建议使用 StartAll() 或 StartAllBot() 方法直接启动所有机器人
//
// - 对于消息处理，可以通过 AddPreprocessor() 方法添加预处理器，通过 AddOnCommand() 方法添加命令处理器，通过 AddListener() 方法添加事件监听器
//
// - 对于插件，可以通过 AddPlugin() 方法添加插件
//
// 整体消息处理的运行与短路顺序为： [main]预处理器 -> [插件]预处理器 -> [插件]令处理器 -> [main]命令处理器 -> [main]事件监听器
func NewBot(bot_id, bot_secret, bot_pubkey, path, addr string) *Bot {
	bot_pubkey = parsePubKey(bot_pubkey) // parse pubkey to appropriate format
	bot_base := models.BotBase{ID: bot_id, Secret: bot_secret, PubKey: bot_pubkey, EncodedSecret: pubKeyEncryptSecret(bot_pubkey, bot_secret)}
	_bot := Bot{
		Base:                                 bot_base,
		addr_key:                             addr,
		path_key:                             path,
		filter_manager:                       &filterManager{entries: make(map[string]time.Time)},
		listeners_join_villa:                 []events.BotListenerJoinVilla{},
		listeners_send_message:               []events.BotListenerSendMessage{},
		listeners_create_robot:               []events.BotListenerCreateRobot{},
		listeners_delete_robot:               []events.BotListenerDeleteRobot{},
		listeners_add_quick_emoticon:         []events.BotListenerAddQuickEmoticon{},
		listeners_audit_callback:             []events.BotListenerAuditCallback{},
		use_default_logger:                   false,
		is_plugins_short_circuit_affect_main: false,
		is_filter_self_msg:                   true,
		is_verify_msg_signature:              true,
		on_commands:                          []commands.OnCommand{},
		preprocessors:                        []commands.Preprocessor{},
		wait_for_command_registers:           []waitForCommandRegister{},
		Api:                                  apis.MakeAPIBase(bot_base, 1*time.Minute),
		Logger:                               logger.NewDefaultLogger(bot_id),
	}
	_bot.abstract_bot = &plugin.AbstractBot{
		Api:            _bot.Api,
		Logger:         _bot.Logger,
		WaitForCommand: _bot.WaitForCommand,
	}

	go _bot.filter_manager.loop()

	port_already_exists := false
	var port_exists_svr_ptr *server_context
	port_path_already_exists := false

	if other_svr_context_manager[addr] != nil {
		port_already_exists = true
		port_exists_svr_ptr = other_svr_context_manager[addr]
		delete(other_svr_context_manager, addr)
	} else {
		for _, _bot_ctx := range bot_context_manager {
			if _bot_ctx.bot.Base.ID == bot_id {
				panic(fmt.Sprintf("bot id of %s already exists", bot_id))
			} else if _bot_ctx.bot.addr_key == addr {
				if _bot_ctx.bot.path_key == path {
					port_path_already_exists = true
				}
				if _bot_ctx.bot.svr != nil {
					port_already_exists = true
					port_exists_svr_ptr = _bot_ctx.svr_ctx
					break
				}
			}
		}
	}

	bot_context_manager[bot_id] = &bot_context{bot: &_bot}

	if port_path_already_exists {
		port_exists_svr_ptr.bots[path] = append(port_exists_svr_ptr.bots[path], &_bot)
		_bot.svr = port_exists_svr_ptr.svr
		bot_context_manager[bot_id].svr_ctx = port_exists_svr_ptr
	} else if port_already_exists {
		port_exists_svr_ptr.bots[path] = []*Bot{&_bot}
		_bot.svr = port_exists_svr_ptr.svr
		bot_context_manager[bot_id].svr_ctx = port_exists_svr_ptr
	} else {
		_bots := map[string][]*Bot{path: {&_bot}}
		svr := gin.Default()
		_bot.svr = svr
		bot_context_manager[bot_id].svr_ctx = &server_context{svr: svr, svr_addr: addr, is_running: false, wg: &sync.WaitGroup{}, bots: _bots, handles: map[string][]func(*gin.Context) bool{}}
	}

	return &_bot
}

// 设置API的超时时间，默认为1分钟
func (_bot *Bot) SetAPITimeout(timeout time.Duration) {
	_bot.Api.SetTimeout(timeout)
}

// 设置bot的日志记录器，默认为os.Stdout+一个log档案
func (_bot *Bot) SetLogger(logger logger.LoggerInterface) {
	_bot.Logger = logger
}

// 设置是否使用默认的日志记录器，默认为false
func (_bot *Bot) SetUseDefaultLogger(is_use bool) {
	_bot.use_default_logger = is_use
}

// 设置插件中的指令短路是否会影响主程序其余指令和监听器的执行，默认为false
func (_bot *Bot) SetPluginsShortCircuitAffectMain(is_affect bool) {
	_bot.is_plugins_short_circuit_affect_main = is_affect
}

// 设置是否启用某一插件
func (_bot *Bot) SetPluginEnabled(plugin_name string, is_enable bool) {
	if _bot.plugins == nil {
		_bot.plugins = plugin.FetchPlugins() // load plugins from plugins context manager
	}
	_bot.plugins[plugin_name].IsEnable = is_enable
}

func (_bot *Bot) GetPluginNames() []string {
	if _bot.plugins == nil {
		_bot.plugins = plugin.FetchPlugins() // load plugins from plugins context manager
	}
	plugin_names := []string{}
	for plugin_name := range _bot.plugins {
		plugin_names = append(plugin_names, plugin_name)
	}
	return plugin_names
}

// 设置是否过滤自己发送的消息，默认为true
func (_bot *Bot) SetFilterSelfMsg(is_filter bool) {
	_bot.is_filter_self_msg = is_filter
}

// 设置是否验证消息签名，默认为true
func (_bot *Bot) SetVerifyMsgSignature(is_verify bool) {
	_bot.is_verify_msg_signature = is_verify
}

func (_bot *Bot) AddListenerJoinVilla(listener events.BotListenerJoinVilla) {
	_bot.listeners_join_villa = append(_bot.listeners_join_villa, listener)
}

func (_bot *Bot) AddListenerSendMessage(listener events.BotListenerSendMessage) {
	_bot.listeners_send_message = append(_bot.listeners_send_message, listener)
}

func (_bot *Bot) AddListenerCreateRobot(listener events.BotListenerCreateRobot) {
	_bot.listeners_create_robot = append(_bot.listeners_create_robot, listener)
}

func (_bot *Bot) AddListenerDeleteRobot(listener events.BotListenerDeleteRobot) {
	_bot.listeners_delete_robot = append(_bot.listeners_delete_robot, listener)
}

func (_bot *Bot) AddListenerAddQuickEmoticon(listener events.BotListenerAddQuickEmoticon) {
	_bot.listeners_add_quick_emoticon = append(_bot.listeners_add_quick_emoticon, listener)
}

func (_bot *Bot) AddListenerAuditCallback(listener events.BotListenerAuditCallback) {
	_bot.listeners_audit_callback = append(_bot.listeners_audit_callback, listener)
}

// 不对回调请求进行任何处理，直接返回到这里注册的监听器，允许用户自行处理回调请求（注意：将根据端口和路径发送回调请求，如使用同端口同路径多机器人，请自行分辨机器人）
func (_bot *Bot) AddlistenerRawRequest(listener events.BotListenerRawRequest) {
	_bot.listeners_raw_request = append(_bot.listeners_raw_request, listener)
}

func (_bot *Bot) RemoveListenerJoinVilla(listener events.BotListenerJoinVilla) {
	for i, l := range _bot.listeners_join_villa {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_join_villa = append(_bot.listeners_join_villa[:i], _bot.listeners_join_villa[i+1:]...)
			return
		}
	}
}

func (_bot *Bot) RemoveListenerSendMessage(listener events.BotListenerSendMessage) {
	for i, l := range _bot.listeners_send_message {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_send_message = append(_bot.listeners_send_message[:i], _bot.listeners_send_message[i+1:]...)
			return
		}
	}
}

func (_bot *Bot) RemoveListenerCreateRobot(listener events.BotListenerCreateRobot) {
	for i, l := range _bot.listeners_create_robot {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_create_robot = append(_bot.listeners_create_robot[:i], _bot.listeners_create_robot[i+1:]...)
			return
		}
	}
}

func (_bot *Bot) RemoveListenerDeleteRobot(listener events.BotListenerDeleteRobot) {
	for i, l := range _bot.listeners_delete_robot {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_delete_robot = append(_bot.listeners_delete_robot[:i], _bot.listeners_delete_robot[i+1:]...)
			return
		}
	}
}

func (_bot *Bot) RemoveListenerAddQuickEmoticon(listener events.BotListenerAddQuickEmoticon) {
	for i, l := range _bot.listeners_add_quick_emoticon {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_add_quick_emoticon = append(_bot.listeners_add_quick_emoticon[:i], _bot.listeners_add_quick_emoticon[i+1:]...)
			return
		}
	}
}

func (_bot *Bot) RemoveListenerAuditCallback(listener events.BotListenerAuditCallback) {
	for i, l := range _bot.listeners_audit_callback {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_audit_callback = append(_bot.listeners_audit_callback[:i], _bot.listeners_audit_callback[i+1:]...)
			return
		}
	}
}

func (_bot *Bot) RemovelistenerRawRequest(listener events.BotListenerRawRequest) {
	for i, l := range _bot.listeners_raw_request {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_raw_request = append(_bot.listeners_raw_request[:i], _bot.listeners_raw_request[i+1:]...)
			return
		}
	}
}

func (_bot *Bot) AddOnCommand(plugin commands.OnCommand) error {
	_bot.on_commands = append(_bot.on_commands, plugin)
	return nil
}

func (_bot *Bot) RemoveOnCommand(plugin commands.OnCommand) error {
	for i, p := range _bot.on_commands {
		if p.Equals(plugin) {
			_bot.on_commands = append(_bot.on_commands[:i], _bot.on_commands[i+1:]...)
			return nil
		}
	}
	return errors.New("plugin not found")
}

func (_bot *Bot) AddPreprocessor(preprocessor commands.Preprocessor) error {
	_bot.preprocessors = append(_bot.preprocessors, preprocessor)
	return nil
}

func (_bot *Bot) RemovePreprocessor(preprocessor commands.Preprocessor) error {
	for i, p := range _bot.preprocessors {
		if reflect.ValueOf(p).Pointer() == reflect.ValueOf(preprocessor).Pointer() {
			_bot.preprocessors = append(_bot.preprocessors[:i], _bot.preprocessors[i+1:]...)
			return nil
		}
	}
	return errors.New("preprocessor not found")
}

func (_bot *Bot) processHandlesBeforeStart() {
	svr_ctx := bot_context_manager[_bot.Base.ID].svr_ctx
	if svr_ctx == nil {
		panic("server context not found")
	}
	// process handles that collides with bot's path
	for p, b := range svr_ctx.bots {
		_p := p
		_b := b
		handles, ok := svr_ctx.handles[_p]
		if ok {
			handles = append(handles, func(c *gin.Context) bool {
				hook(_b, c)
				return false
			})
			delete(svr_ctx.handles, _p)
		} else {
			handles = []func(*gin.Context) bool{func(c *gin.Context) bool {
				hook(_b, c)
				return false
			}}
		}
		utils.Try(func() {
			svr_ctx.svr.Any(_p, func(c *gin.Context) {
				for _, handle := range handles {
					if handle(c) { // perform short circuit middleware
						return
					}
				}
			})
		}, func(e interface{}, s string) {
			_bot.Logger.Error("failed to add handle: ", e)
		})
	}
	// process other handles
	for p, h := range svr_ctx.handles {
		_p := p
		_h := h
		utils.Try(func() {
			svr_ctx.svr.Any(_p, func(c *gin.Context) {
				for _, handle := range _h {
					if handle(c) { // perform short circuit middleware
						return
					}
				}
				c.String(http.StatusNotFound, "<h1>404</h1><p>page not found</p>")
			})
		}, func(e interface{}, s string) {
			_bot.Logger.Error("failed to add handle: ", e)
		})
	}
}

func (_bot *Bot) Start() error {
	_bot.is_running = true
	if _bot.plugins == nil {
		_bot.plugins = plugin.FetchPlugins() // load plugins from plugins context manager
	}

	for _plugin_name, _plugin := range _bot.plugins {
		var _enable string
		if _plugin.IsEnable {
			_enable = "启用"
		} else {
			_enable = "禁用"
		}
		_bot.Logger.Infof("机器人 {%v} 加载了插件 %s (%s)\n", _bot.Base.ID, _plugin_name, _enable)
	}

	_bot.Logger.Infof("机器人 {%v} 于 localhost%v 开始运行\n", _bot.Base.ID, _bot.addr_key)
	var _bot_ctx *bot_context
	for _, bot_ctx := range bot_context_manager {
		if bot_ctx.bot == _bot {
			_bot_ctx = bot_ctx
			break
		}
	}
	if _bot_ctx.svr_ctx.is_running {
		_bot_ctx.svr_ctx.wg.Wait()
		return nil
	} else {
		_bot_ctx.svr_ctx.is_running = true
		_bot.processHandlesBeforeStart()
		_bot_ctx.svr_ctx.wg.Add(1)
		defer _bot_ctx.svr_ctx.wg.Done()
		err := _bot.svr.Run(_bot.addr_key)
		_bot.Logger.Error("机器人 {%v} 停止监听 localhost%v : %v\n", _bot.Base.ID, _bot.addr_key, err)
		return err
	}
}

func StartAllBot() {
	var wg sync.WaitGroup
	for _, bot_ctx := range bot_context_manager {
		wg.Add(1)
		go func(_bot *Bot) {
			defer wg.Done()
			_bot.Start()
		}(bot_ctx.bot)
	}
	wg.Wait()
}

/* Global */

// 开始运行所有机器人和 HTTP 服务器
func StartAll() {
	StartAllBot()
	StartAllHttpServer()
}

func init() {
	gin.SetMode(gin.ReleaseMode) // default to release mode
}
