package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

type bot struct {
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
	bots       map[string][]*bot                    // path: bot
	handles    map[string][]func(*gin.Context) bool // path: handler
}

type bot_context struct {
	bot     *bot
	svr_ctx *server_context
}

var bot_context_manager map[string]*bot_context = make(map[string]*bot_context) // id: bot
var other_svr_context_manager map[string]*server_context = make(map[string]*server_context)

/* context managers end */

/* http routing related */

// 新增一个http路由，用于与机器人同时运行，避免占用同一个端口，以及允许使用短路机制处理请求
// path: 路由路径
// addr: 路由地址
// handler: 路由处理器，返回true则短路，返回false则继续执行（系统将在所有同路径端口的handler都不短路后再处理消息）
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
			bots:       make(map[string][]*bot),
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

func processEvent(_bot *bot, event events.Event) {
	event_id := event.Event.Id
	if _bot.filter_manager.needFilter(event_id) {
		_bot.Logger.Debugf("filter repeat event: %v+\n", event)
		return
	} else {
		_bot.filter_manager.add(event_id)
	}

	if _bot.use_default_logger {
		_bot.Logger.Debugf("receive event: %v+\n", event)
	}
	event_type := event.Event.Type
switch_label:
	switch event_type {
	case events.JoinVilla:
		event := events.Event2EventJoinVilla(event)
		for _, listener := range _bot.listeners_join_villa {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	case events.SendMessage:
		event := events.Event2EventSendMessage(event, _bot.Api)
		if _bot.is_filter_self_msg && event.Data.Content.User.Id == _bot.Base.ID {
			break switch_label
		}
		// 1. run preprocessors
		for _, _preprocessor := range _bot.preprocessors {
			utils.Try(func() { _preprocessor(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("preprocessor {", utils.GetFunctionName(_preprocessor), "} error: ", err, "\n", tb)
			})
		}
		// 2. run wait_for command registers
		if _bot.checkWaifForCommand(event) {
			break switch_label
		}
		// 3. run plugins

		// 3_1. run plugins preprocessors
		for _, p := range _bot.plugins {
			if p.IsEnable {
				for _, _preprocessor := range p.Preprocessors {
					utils.Try(func() { _preprocessor(event, _bot.abstract_bot) }, func(err interface{}, tb string) {
						_bot.Logger.Error("preprocessor {", utils.GetFunctionName(_preprocessor), "} error: ", err, "\n", tb)
					})
				}
			}
		}

		// 3_2. run plugins commands
		for _, p := range _bot.plugins {
			if p.IsEnable {
				_is_short_circuit := false
				for _, _command := range p.OnCommand {
					if _command.CheckCommand(event, _bot.abstract_bot) {
						_is_short_circuit = true
						break // short circuit for plugin's internal commands
					}
				}
				if _is_short_circuit && _bot.is_plugins_short_circuit_affect_main {
					break switch_label // short circuit for all commands
				}
			}
		}

		// 4. run on commands
		for _, _command := range _bot.on_commands {
			if _command.CheckCommand(event, _bot.Logger, _bot.Api) {
				break switch_label // short circuit
			}
		}
		// 5. run normal listeners
		for _, listener := range _bot.listeners_send_message {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	case events.CreateRobot:
		event := events.Event2EventCreateRobot(event)
		for _, listener := range _bot.listeners_create_robot {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	case events.DeleteRobot:
		event := events.Event2EventDeleteRobot(event)
		for _, listener := range _bot.listeners_delete_robot {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	case events.AddQuickEmoticon:
		event := events.Event2EventAddQuickEmoticon(event)
		for _, listener := range _bot.listeners_add_quick_emoticon {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	case events.AuditCallback:
		event := events.Event2EventAuditCallback(event)
		for _, listener := range _bot.listeners_audit_callback {
			utils.Try(func() { listener(event) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	default:
		_bot.Logger.Warnf("unknown event type: %v\n", event_type)
	}
}

func hook(_bots []*bot, c *gin.Context) {
	// send raw request to listeners on bot of current path and port
	for _, _bot := range _bots {
		if !_bot.is_running {
			continue
		}
		for _, listener := range _bot.listeners_raw_request {
			go utils.Try(func() { listener(c) }, func(err interface{}, tb string) {
				_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err, "\n", tb)
			})
		}
	}

	// handle event
	if c.Request.Method != "POST" || c.Request.Header.Get("Content-Type") != "application/json" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
			"retcode": -1,
		})
		return
	}

	defer func() { // ensure response
		c.JSON(http.StatusOK, gin.H{
			"message": "",
			"retcode": 0,
		})
	}()

	raw_body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println("new event read body error: ", err)
		return
	}

	var event events.Event
	err = json.Unmarshal(raw_body, &event)
	if err != nil {
		fmt.Println("decode event error (" + err.Error() + "): " + string(raw_body))
		return
	}

	_id := event.Event.Robot.Template.Id
	_bot_ctx, ok := bot_context_manager[_id] // find bot ctx by id to allow multiple bot running on same port &|| path
	if !ok {
		// fmt.Println("bot id: ", _id, " not found in ctx: ", bot_context_manager)
		return
	}

	_bot := _bot_ctx.bot

	if !_bot.is_running {
		return
	}
	if _bot.is_verify_msg_signature {
		sign := c.Request.Header.Get("x-rpc-bot_sign")
		verify, err := pubKeyVerify(sign, string(raw_body), _bot.Base.Secret, _bot.Base.PubKey)
		if (!verify) || (err != nil) {
			_bot.Logger.Debug("new event verify error, rejected: ", err)
			return
		}
	}
	go processEvent(_bot, event) // use goroutine to avoid blocking (especially handle wait_for)

}

// NewBot 创建一个机器人实例，bot_id 为机器人的id，bot_secret 为机器人的secret，path 为接收事件的路径（如"/"），addr 为接收事件的地址（如":8888"）；
// 机器人实例创建后，需要调用 Run() 方法启动机器人；
// 对于消息处理，可以通过 AddPreprocessor() 方法添加预处理器，通过 AddOnCommand() 方法添加命令处理器，通过 AddListener() 方法添加事件监听器；
// 对于插件，可以通过 AddPlugin() 方法添加插件；
// 整体消息处理的运行与短路顺序为： [main]预处理器 -> [插件]预处理器 -> [插件]令处理器 -> [main]命令处理器 -> [main]事件监听器；
func NewBot(bot_id, bot_secret, bot_pubkey, path, addr string) *bot {
	bot_pubkey = parsePubKey(bot_pubkey) // parse pubkey to appropriate format
	bot_base := models.BotBase{ID: bot_id, Secret: bot_secret, PubKey: bot_pubkey, EncodedSecret: pubKeyEncryptSecret(bot_pubkey, bot_secret)}
	_bot := bot{
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
		Api:                                  &apis.ApiBase{Base: bot_base},
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
		port_exists_svr_ptr.bots[path] = []*bot{&_bot}
		_bot.svr = port_exists_svr_ptr.svr
		bot_context_manager[bot_id].svr_ctx = port_exists_svr_ptr
	} else {
		_bots := map[string][]*bot{path: {&_bot}}
		svr := gin.Default()
		_bot.svr = svr
		bot_context_manager[bot_id].svr_ctx = &server_context{svr: svr, svr_addr: addr, is_running: false, wg: &sync.WaitGroup{}, bots: _bots, handles: map[string][]func(*gin.Context) bool{}}
	}

	return &_bot
}

// 设置bot的日志记录器，默认为os.Stdout+一个log档案
func (_bot *bot) SetLogger(logger logger.LoggerInterface) {
	_bot.Logger = logger
}

// 设置是否使用默认的日志记录器，默认为false
func (_bot *bot) SetUseDefaultLogger(is_use bool) {
	_bot.use_default_logger = is_use
}

// 设置插件中的指令短路是否会影响主程序其余指令和监听器的执行，默认为false
func (_bot *bot) SetPluginsShortCircuitAffectMain(is_affect bool) {
	_bot.is_plugins_short_circuit_affect_main = is_affect
}

// 设置是否启用某一插件
func (_bot *bot) SetPluginEnabled(plugin_name string, is_enable bool) {
	if _bot.plugins == nil {
		_bot.plugins = plugin.FetchPlugins() // load plugins from plugins context manager
	}
	_bot.plugins[plugin_name].IsEnable = is_enable
}

func (_bot *bot) GetPluginNames() []string {
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
func (_bot *bot) SetFilterSelfMsg(is_filter bool) {
	_bot.is_filter_self_msg = is_filter
}

// 设置是否验证消息签名，默认为true
func (_bot *bot) SetVerifyMsgSignature(is_verify bool) {
	_bot.is_verify_msg_signature = is_verify
}

func (_bot *bot) AddListenerJoinVilla(listener events.BotListenerJoinVilla) {
	_bot.listeners_join_villa = append(_bot.listeners_join_villa, listener)
}

func (_bot *bot) AddListenerSendMessage(listener events.BotListenerSendMessage) {
	_bot.listeners_send_message = append(_bot.listeners_send_message, listener)
}

func (_bot *bot) AddListenerCreateRobot(listener events.BotListenerCreateRobot) {
	_bot.listeners_create_robot = append(_bot.listeners_create_robot, listener)
}

func (_bot *bot) AddListenerDeleteRobot(listener events.BotListenerDeleteRobot) {
	_bot.listeners_delete_robot = append(_bot.listeners_delete_robot, listener)
}

func (_bot *bot) AddListenerAddQuickEmoticon(listener events.BotListenerAddQuickEmoticon) {
	_bot.listeners_add_quick_emoticon = append(_bot.listeners_add_quick_emoticon, listener)
}

func (_bot *bot) AddListenerAuditCallback(listener events.BotListenerAuditCallback) {
	_bot.listeners_audit_callback = append(_bot.listeners_audit_callback, listener)
}

// 不对回调请求进行任何处理，直接返回到这里注册的监听器，允许用户自行处理回调请求（注意：将根据端口和路径发送回调请求，如使用同端口同路径多机器人，请自行分辨机器人）
func (_bot *bot) AddlistenerRawRequest(listener events.BotListenerRawRequest) {
	_bot.listeners_raw_request = append(_bot.listeners_raw_request, listener)
}

func (_bot *bot) RemoveListenerJoinVilla(listener events.BotListenerJoinVilla) {
	for i, l := range _bot.listeners_join_villa {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_join_villa = append(_bot.listeners_join_villa[:i], _bot.listeners_join_villa[i+1:]...)
			return
		}
	}
}

func (_bot *bot) RemoveListenerSendMessage(listener events.BotListenerSendMessage) {
	for i, l := range _bot.listeners_send_message {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_send_message = append(_bot.listeners_send_message[:i], _bot.listeners_send_message[i+1:]...)
			return
		}
	}
}

func (_bot *bot) RemoveListenerCreateRobot(listener events.BotListenerCreateRobot) {
	for i, l := range _bot.listeners_create_robot {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_create_robot = append(_bot.listeners_create_robot[:i], _bot.listeners_create_robot[i+1:]...)
			return
		}
	}
}

func (_bot *bot) RemoveListenerDeleteRobot(listener events.BotListenerDeleteRobot) {
	for i, l := range _bot.listeners_delete_robot {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_delete_robot = append(_bot.listeners_delete_robot[:i], _bot.listeners_delete_robot[i+1:]...)
			return
		}
	}
}

func (_bot *bot) RemoveListenerAddQuickEmoticon(listener events.BotListenerAddQuickEmoticon) {
	for i, l := range _bot.listeners_add_quick_emoticon {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_add_quick_emoticon = append(_bot.listeners_add_quick_emoticon[:i], _bot.listeners_add_quick_emoticon[i+1:]...)
			return
		}
	}
}

func (_bot *bot) RemoveListenerAuditCallback(listener events.BotListenerAuditCallback) {
	for i, l := range _bot.listeners_audit_callback {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_audit_callback = append(_bot.listeners_audit_callback[:i], _bot.listeners_audit_callback[i+1:]...)
			return
		}
	}
}

func (_bot *bot) RemovelistenerRawRequest(listener events.BotListenerRawRequest) {
	for i, l := range _bot.listeners_raw_request {
		if reflect.ValueOf(l).Pointer() == reflect.ValueOf(listener).Pointer() {
			_bot.listeners_raw_request = append(_bot.listeners_raw_request[:i], _bot.listeners_raw_request[i+1:]...)
			return
		}
	}
}

func (_bot *bot) AddOnCommand(plugin commands.OnCommand) error {
	_bot.on_commands = append(_bot.on_commands, plugin)
	return nil
}

func (_bot *bot) RemoveOnCommand(plugin commands.OnCommand) error {
	for i, p := range _bot.on_commands {
		if p.Equals(plugin) {
			_bot.on_commands = append(_bot.on_commands[:i], _bot.on_commands[i+1:]...)
			return nil
		}
	}
	return errors.New("plugin not found")
}

func (_bot *bot) AddPreprocessor(preprocessor commands.Preprocessor) error {
	_bot.preprocessors = append(_bot.preprocessors, preprocessor)
	return nil
}

func (_bot *bot) RemovePreprocessor(preprocessor commands.Preprocessor) error {
	for i, p := range _bot.preprocessors {
		if reflect.ValueOf(p).Pointer() == reflect.ValueOf(preprocessor).Pointer() {
			_bot.preprocessors = append(_bot.preprocessors[:i], _bot.preprocessors[i+1:]...)
			return nil
		}
	}
	return errors.New("preprocessor not found")
}

func (_bot *bot) processHandlesBeforeStart() {
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

func (_bot *bot) Start() error {
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
		go func(_bot *bot) {
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
