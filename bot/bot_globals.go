package bot

import (
	"sync"

	apis "github.com/GLGDLY/mhy_botsdk/apis"
	commands "github.com/GLGDLY/mhy_botsdk/commands"
	events "github.com/GLGDLY/mhy_botsdk/events"
	logger "github.com/GLGDLY/mhy_botsdk/logger"
	models "github.com/GLGDLY/mhy_botsdk/models"
	plugin "github.com/GLGDLY/mhy_botsdk/plugins"

	"github.com/gin-gonic/gin"
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
	reverse_proxy_http_msg_chan []chan [2][]byte // [body, sign]
	reverse_proxy_ws_msg_chan   []chan [2][]byte // [body, sign]
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
type serverContext struct {
	svr        *gin.Engine
	svr_addr   string
	is_running bool
	wg         *sync.WaitGroup
	bots       map[string][]*Bot                    // path: bot
	handles    map[string][]func(*gin.Context) bool // path: handler
}

type wsContext struct {
	uri        string
	is_running bool
	wg         *sync.WaitGroup
	bots       map[string][]*Bot // path: bot
}

type botContext struct {
	bot     *Bot
	svr_ctx *serverContext
	ws_ctx  *wsContext
}

var bot_context_manager map[string]*botContext = make(map[string]*botContext) // id: bot
var other_svr_context_manager map[string]*serverContext = make(map[string]*serverContext)

/* context managers end */

/* Global */

// 开始运行所有机器人和 HTTP 服务器
func StartAll() {
	StartAllBot()
	StartAllHttpServer()
}

func init() {
	gin.SetMode(gin.ReleaseMode) // default to release mode
}
