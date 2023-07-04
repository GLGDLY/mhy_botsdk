package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"

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
	base     models.BotBase // 机器人基本信息
	path_key string
	addr_key string
	svr      *http.Server
	/* 事件监听器开始 */
	listeners_join_villa         []events.BotListenerJoinVilla
	listeners_send_message       []events.BotListenerSendMessage
	listeners_create_robot       []events.BotListenerCreateRobot
	listeners_delete_robot       []events.BotListenerDeleteRobot
	listeners_add_quick_emoticon []events.BotListenerAddQuickEmoticon
	listeners_audit_callback     []events.BotListenerAuditCallback
	/* 事件监听器结束 */
	use_default_logger                   bool                      // 是否使用默认的日志记录器，默认为false
	is_plugins_short_circuit_affect_main bool                      // 插件中的指令短路是否会影响主程序其余指令和监听器的执行，默认为false
	is_filter_self_msg                   bool                      // 是否过滤自己发送的消息，默认为true
	plugins                              map[string]*plugin.Plugin // 插件列表
	on_commands                          []commands.OnCommand      // 处理消息事件的指令列表
	preprocessors                        []commands.Preprocessor   // 消息事的预处理器，用于在运行指令列表和监听器之前处理事件
	Api                                  *apis.ApiBase             // api接口
	Logger                               logger.LoggerInterface    // 日志记录器
}

/* context managers start */
type server_context struct {
	svr        *http.Server
	is_running bool
	wg         *sync.WaitGroup
}

type bot_context struct {
	bot     *bot
	svr_ctx *server_context
}

var bot_context_manager map[string]*bot_context = make(map[string]*bot_context) // id: bot

/* context managers end */

func hook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" || r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message":"bad request","retcode":-1}`))
		return
	}
	var event events.Event
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		var _bot *bot
		for _, b := range bot_context_manager {
			_bot = b.bot
			break
		}
		var raw_data []byte
		_, err_read := r.Body.Read(raw_data)
		if err != nil {
			_bot.Logger.Error("decode event error (" + err.Error() + "): " + string(raw_data))
		} else {
			_bot.Logger.Error("decode event error (" + err.Error() + ", " + err_read.Error() + ")")
		}
	} else {
		_id := event.Event.Robot.Template.Id
		_bot_ctx, ok := bot_context_manager[_id] // find bot ctx by id to allow multiple bot running on same port &|| path
		if !ok {
			fmt.Println("bot id: ", _id, " not found in ctx: ", bot_context_manager)
			return
		}
		_bot := _bot_ctx.bot
		if _bot.use_default_logger {
			_bot.Logger.Debugf("receive event: %v+\n", event)
		}
		event_type := event.Event.Type
	switch_label:
		switch event_type {
		case events.JoinVilla:
			event := events.Event2EventJoinVilla(event)
			for _, listener := range _bot.listeners_join_villa {
				utils.Try(func() { listener(event) }, func(err interface{}) {
					_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err)
				})
			}
		case events.SendMessage:
			event := events.Event2EventSendMessage(event, _bot.Api)
			if _bot.is_filter_self_msg && event.Data.Content.User.Id == _bot.base.ID {
				break switch_label
			}
			for _, _preprocessor := range _bot.preprocessors {
				utils.Try(func() { _preprocessor(event) }, func(err interface{}) {
					_bot.Logger.Error("preprocessor {", utils.GetFunctionName(_preprocessor), "} error: ", err)
				})
			}
			for _, p := range _bot.plugins {
				if p.IsEnable {
					for _, _preprocessor := range p.Preprocessors {
						utils.Try(func() { _preprocessor(event, _bot.Api, _bot.Logger) }, func(err interface{}) {
							_bot.Logger.Error("preprocessor {", utils.GetFunctionName(_preprocessor), "} error: ", err)
						})
					}
					_is_short_circuit := false
					for _, _command := range p.OnCommand {
						if _command.CheckCommand(event, _bot.Logger, _bot.Api) {
							_is_short_circuit = true
							break // short circuit for plugin's internal commands
						}
					}
					if _is_short_circuit && _bot.is_plugins_short_circuit_affect_main {
						break switch_label // short circuit for all commands
					}
				}
			}
			for _, _command := range _bot.on_commands {
				if _command.CheckCommand(event, _bot.Logger, _bot.Api) {
					break switch_label // short circuit
				}
			}
			for _, listener := range _bot.listeners_send_message {
				utils.Try(func() { listener(event) }, func(err interface{}) {
					_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err)
				})
			}
		case events.CreateRobot:
			event := events.Event2EventCreateRobot(event)
			for _, listener := range _bot.listeners_create_robot {
				utils.Try(func() { listener(event) }, func(err interface{}) {
					_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err)
				})
			}
		case events.DeleteRobot:
			event := events.Event2EventDeleteRobot(event)
			for _, listener := range _bot.listeners_delete_robot {
				utils.Try(func() { listener(event) }, func(err interface{}) {
					_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err)
				})
			}
		case events.AddQuickEmoticon:
			event := events.Event2EventAddQuickEmoticon(event)
			for _, listener := range _bot.listeners_add_quick_emoticon {
				utils.Try(func() { listener(event) }, func(err interface{}) {
					_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err)
				})
			}
		case events.AuditCallback:
			event := events.Event2EventAuditCallback(event)
			for _, listener := range _bot.listeners_audit_callback {
				utils.Try(func() { listener(event) }, func(err interface{}) {
					_bot.Logger.Error("listener {", utils.GetFunctionName(listener), "} error: ", err)
				})
			}
		default:
			_bot.Logger.Warnf("unknown event type: %v\n", event_type)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"","retcode":0}`))
}

// NewBot 创建一个机器人实例，bot_id 为机器人的id，bot_secret 为机器人的secret，path 为接收事件的路径（如"/"），addr 为接收事件的地址（如":8888"）；
// 机器人实例创建后，需要调用 Run() 方法启动机器人；
// 对于消息处理，可以通过 AddPreprocessor() 方法添加预处理器，通过 AddOnCommand() 方法添加命令处理器，通过 AddListener() 方法添加事件监听器；
// 对于插件，可以通过 AddPlugin() 方法添加插件；
// 整体消息处理的运行与短路顺序为： [main]预处理器 -> [插件]预处理器 -> [插件]令处理器 -> [main]命令处理器 -> [main]事件监听器；
func NewBot(bot_id, bot_secret, path, addr string) *bot {
	bot_base := models.BotBase{ID: bot_id, Secret: bot_secret}
	_bot := bot{
		base:                                 bot_base,
		addr_key:                             addr,
		path_key:                             path,
		listeners_join_villa:                 []events.BotListenerJoinVilla{},
		listeners_send_message:               []events.BotListenerSendMessage{},
		listeners_create_robot:               []events.BotListenerCreateRobot{},
		listeners_delete_robot:               []events.BotListenerDeleteRobot{},
		listeners_add_quick_emoticon:         []events.BotListenerAddQuickEmoticon{},
		listeners_audit_callback:             []events.BotListenerAuditCallback{},
		use_default_logger:                   false,
		is_plugins_short_circuit_affect_main: false,
		is_filter_self_msg:                   true,
		plugins:                              plugin.FetchPlugins(), // load plugins from plugins context manager
		on_commands:                          []commands.OnCommand{},
		preprocessors:                        []commands.Preprocessor{},
		Api:                                  &apis.ApiBase{Base: bot_base},
		Logger:                               logger.NewDefaultLogger(bot_id),
	}

	port_already_exists := false
	var port_exists_svr_ptr *server_context
	port_path_already_exists := false
	for _, _bot_ctx := range bot_context_manager {
		if _bot_ctx.bot.base.ID == bot_id {
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

	bot_context_manager[bot_id] = &bot_context{bot: &_bot}

	if port_path_already_exists {
		_bot.svr = port_exists_svr_ptr.svr
		bot_context_manager[bot_id].svr_ctx = port_exists_svr_ptr
	} else if port_already_exists {
		mux := port_exists_svr_ptr.svr.Handler.(*http.ServeMux)
		mux.HandleFunc(path, hook)
		_bot.svr = port_exists_svr_ptr.svr
		bot_context_manager[bot_id].svr_ctx = port_exists_svr_ptr
	} else {
		mux := http.NewServeMux()
		mux.HandleFunc(path, hook)
		svr := &http.Server{
			Addr:    addr,
			Handler: mux,
		}
		_bot.svr = svr
		bot_context_manager[bot_id].svr_ctx = &server_context{svr: svr, is_running: false, wg: &sync.WaitGroup{}}
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
	_bot.plugins[plugin_name].IsEnable = is_enable
}

// 设置是否过滤自己发送的消息，默认为true
func (_bot *bot) SetFilterSelfMsg(is_filter bool) {
	_bot.is_filter_self_msg = is_filter
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

func (_bot *bot) Start() error {
	_bot.Logger.Infof("机器人 {%v} 于 localhost%v 开始运行\n", _bot.base.ID, _bot.svr.Addr)
	var _bot_ctx *bot_context
	for _, bot_ctx := range bot_context_manager {
		if bot_ctx.bot == _bot {
			_bot_ctx = bot_ctx
			break
		}
	}
	var err error
	if _bot_ctx.svr_ctx.is_running {
		_bot_ctx.svr_ctx.wg.Wait()
		return nil
	} else {
		_bot_ctx.svr_ctx.is_running = true
		_bot_ctx.svr_ctx.wg.Add(1)
		defer _bot_ctx.svr_ctx.wg.Done()
		err = _bot.svr.ListenAndServe()
	}
	return err
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
