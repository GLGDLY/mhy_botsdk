package plugins

import (
	"regexp"
	"strings"

	api_models "github.com/GLGDLY/mhy_botsdk/api_models"
	apis "github.com/GLGDLY/mhy_botsdk/apis"
	events "github.com/GLGDLY/mhy_botsdk/events"
	logger "github.com/GLGDLY/mhy_botsdk/logger"
	"github.com/GLGDLY/mhy_botsdk/utils"
)

/* Plugins version of commands.go */
type plugin_msg_listener func(data events.EventSendMessage, _api *apis.ApiBase, _logger logger.LoggerInterface)

type Preprocessor plugin_msg_listener
type OnCommand struct {
	Command        []string       // 可触发事件的指令列表，与正则 Regex 互斥，优先使用此项
	Regex          string         // 可触发指令的正则表达式，与指令表 Command 互斥
	regex          *regexp.Regexp // internal use
	Listener       plugin_msg_listener
	RequireAT      bool   // 是否要求必须@机器人才能触发指令
	RequireAdmin   bool   // 是否要求频道主或或管理才可触发指令
	AdminErrorMsg  string // 当RequireAT，而触发用户的权限不足时，如此项不为空，返回此消息并短路；否则不进行短路
	IsShortCircuit bool   // 如果触发指令成功是否短路不运行后续指令（将根据注册顺序排序指令的短路机制，且插件中的短路是否影响主程序会根据bot的is_plugins_short_circuit_affect_main决定）
}

type Plugin struct {
	IsEnable      bool
	Preprocessors []Preprocessor
	OnCommand     []OnCommand
}

func (p *OnCommand) processCommand(data events.EventSendMessage, _logger logger.LoggerInterface, _api *apis.ApiBase) bool {
	at := "@" + data.Robot.Template.Name
	if p.RequireAT && !strings.Contains(data.GetContent(false), at) {
		return false
	}
	if p.RequireAdmin {
		res, http_code, err := _api.GetMember(data.Robot.VillaId, data.Data.FromUserId)
		if err != nil {
			_logger.Error("command listener {", utils.GetFunctionName(p.Listener), "} get member role info error: ", err)
			return false
		} else if http_code != 200 || res.Retcode != 0 {
			_logger.Error("command listener {", utils.GetFunctionName(p.Listener), "} get member role info error: ", res)
			return false
		}
		is_admin := false
		for _, v := range res.Data.Member.RoleList {
			if v.RoleType == "MEMBER_ROLE_TYPE_ADMIN" || v.RoleType == "MEMBER_ROLE_TYPE_OWNER" {
				is_admin = true
				break
			}
		}
		if !is_admin {
			if p.AdminErrorMsg != "" {
				msg, _ := api_models.NewMsg(api_models.MsgTypeText)
				msg.SetText(p.AdminErrorMsg)
				_api.SendMessage(data.Robot.VillaId, data.Data.RoomId, msg)
				return true
			}
			return false
		}
	}
	utils.Try(func() { p.Listener(data, _api, _logger) }, func(err interface{}) {
		_logger.Error("plugins listener {", utils.GetFunctionName(p.Listener), "} error: ", err)
	})
	return p.IsShortCircuit
}

// 内部检查当前消息是否符合触发条件
func (p *OnCommand) CheckCommand(data events.EventSendMessage, _logger logger.LoggerInterface, _api *apis.ApiBase) bool {
	msg := data.GetContent(false)
	if p.Command != nil {
		for _, v := range p.Command {
			if strings.Contains(msg, v) {
				if p.processCommand(data, _logger, _api) {
					return true
				}
			}
		}
	}
	if p.regex == nil && p.Regex != "" {
		p.regex = regexp.MustCompile(p.Regex)
	}
	if p.regex != nil {
		if p.regex.MatchString(msg) {
			if p.processCommand(data, _logger, _api) {
				return true
			}
		}
	}
	return false
}

func (p *OnCommand) Equals(_p OnCommand) bool {
	if p.Command != nil && _p.Command != nil {
		if len(p.Command) != len(_p.Command) {
			return false
		}
		p_map := make(map[string]bool)
		for _, v := range p.Command {
			p_map[v] = true
		}
		for _, v := range _p.Command {
			if _, ok := p_map[v]; !ok {
				return false
			}
		}
		return true
	}
	if p.Regex != "" && _p.Regex != "" {
		return p.Regex == _p.Regex
	}
	return false
}

/* Context Managment for Plugins */
var context_manager = make(map[string]*Plugin)

// 注册插件到下一个Bot实例
func RegisterPlugin(name string, context *Plugin) {
	context.IsEnable = true
	context_manager[name] = context
}

func FetchPlugins() map[string]*Plugin {
	context_manager_copy := make(map[string]*Plugin)
	for k, v := range context_manager {
		_v := *v
		context_manager_copy[k] = &_v
	}
	return context_manager_copy
}

func ClearPlugins() {
	context_manager = make(map[string]*Plugin)
}
